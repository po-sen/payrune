package use_cases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/in"
	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/value_objects"
)

const (
	defaultReceiptPollingInterval = 15 * time.Second
	defaultReceiptPollingClaimTTL = 30 * time.Second
)

type runReceiptPollingCycleUseCase struct {
	unitOfWork outport.UnitOfWork
	observer   outport.BlockchainReceiptObserver
	clock      outport.Clock
}

func NewRunReceiptPollingCycleUseCase(
	unitOfWork outport.UnitOfWork,
	observer outport.BlockchainReceiptObserver,
	clock outport.Clock,
) inport.RunReceiptPollingCycleUseCase {
	return &runReceiptPollingCycleUseCase{
		unitOfWork: unitOfWork,
		observer:   observer,
		clock:      clock,
	}
}

func (uc *runReceiptPollingCycleUseCase) Execute(
	ctx context.Context,
	input dto.RunReceiptPollingCycleInput,
) (dto.RunReceiptPollingCycleOutput, error) {
	if uc.unitOfWork == nil {
		return dto.RunReceiptPollingCycleOutput{}, errors.New("unit of work is not configured")
	}
	if uc.observer == nil {
		return dto.RunReceiptPollingCycleOutput{}, errors.New("blockchain receipt observer is not configured")
	}
	if uc.clock == nil {
		return dto.RunReceiptPollingCycleOutput{}, errors.New("clock is not configured")
	}
	if input.BatchSize <= 0 {
		return dto.RunReceiptPollingCycleOutput{}, errors.New("batch size must be greater than zero")
	}

	now := uc.clock.NowUTC()
	pollInterval := input.PollInterval
	if pollInterval <= 0 {
		pollInterval = defaultReceiptPollingInterval
	}
	claimTTL := input.ClaimTTL
	if claimTTL <= 0 {
		claimTTL = defaultReceiptPollingClaimTTL
	}
	chainFilter, networkFilter, err := resolveReceiptPollingScope(input.Chain, input.Network)
	if err != nil {
		return dto.RunReceiptPollingCycleOutput{}, err
	}

	var trackings []entities.PaymentReceiptTracking
	err = uc.unitOfWork.WithinTransaction(ctx, func(txRepositories outport.TxRepositories) error {
		repository := txRepositories.PaymentReceiptTracking
		if repository == nil {
			return errors.New("payment receipt tracking repository is not configured")
		}

		claimedTrackings, err := repository.ClaimDue(ctx, outport.ClaimPaymentReceiptTrackingsInput{
			Now:        now,
			Limit:      input.BatchSize,
			ClaimUntil: now.Add(claimTTL),
			Chain:      chainFilter,
			Network:    networkFilter,
		})
		if err != nil {
			return err
		}

		trackings = claimedTrackings
		return nil
	})
	if err != nil {
		return dto.RunReceiptPollingCycleOutput{}, err
	}

	output := dto.RunReceiptPollingCycleOutput{ClaimedCount: len(trackings)}

	for _, tracking := range trackings {
		if tracking.IssuedAt.IsZero() {
			if err := uc.savePollingError(
				ctx,
				tracking,
				errors.New("issued at is required"),
				now,
				now.Add(pollInterval),
			); err != nil {
				return output, err
			}
			output.FailedCount++
			continue
		}

		observation, observeErr := uc.observer.ObserveAddress(ctx, outport.ObserveChainPaymentAddressInput{
			Chain:                 tracking.Chain,
			Network:               tracking.Network,
			Address:               tracking.Address,
			IssuedAt:              tracking.IssuedAt,
			RequiredConfirmations: tracking.RequiredConfirmations,
			SinceBlockHeight:      tracking.LastObservedBlockHeight,
		})
		if observeErr != nil {
			if err := uc.savePollingError(
				ctx,
				tracking,
				observeErr,
				now,
				now.Add(pollInterval),
			); err != nil {
				return output, err
			}
			output.FailedCount++
			continue
		}

		updatedTracking, err := tracking.ApplyObservation(value_objects.PaymentReceiptObservation{
			ObservedTotalMinor:    observation.ObservedTotalMinor,
			ConfirmedTotalMinor:   observation.ConfirmedTotalMinor,
			UnconfirmedTotalMinor: observation.UnconfirmedTotalMinor,
			ConflictTotalMinor:    observation.ConflictTotalMinor,
			LatestBlockHeight:     observation.LatestBlockHeight,
		}, now)
		if err != nil {
			if saveErr := uc.savePollingError(
				ctx,
				tracking,
				err,
				now,
				now.Add(pollInterval),
			); saveErr != nil {
				return output, saveErr
			}
			output.FailedCount++
			continue
		}

		nextPollAt := now.Add(pollInterval)
		if updatedTracking.Status == value_objects.PaymentReceiptStatusPaidConfirmed {
			nextPollAt = now.Add(24 * time.Hour)
		}

		if err := uc.saveObservation(ctx, updatedTracking, now, nextPollAt); err != nil {
			return output, err
		}
		output.UpdatedCount++
	}

	return output, nil
}

func (uc *runReceiptPollingCycleUseCase) savePollingError(
	ctx context.Context,
	tracking entities.PaymentReceiptTracking,
	observeErr error,
	now time.Time,
	nextPollAt time.Time,
) error {
	trackingWithError, markErr := tracking.MarkPollingError(observeErr.Error())
	if markErr != nil {
		return fmt.Errorf("mark polling error: %w", markErr)
	}
	return uc.unitOfWork.WithinTransaction(ctx, func(txRepositories outport.TxRepositories) error {
		repository := txRepositories.PaymentReceiptTracking
		if repository == nil {
			return errors.New("payment receipt tracking repository is not configured")
		}
		return repository.SavePollingError(
			ctx,
			trackingWithError.PaymentAddressID,
			trackingWithError.LastError,
			now,
			nextPollAt,
		)
	})
}

func (uc *runReceiptPollingCycleUseCase) saveObservation(
	ctx context.Context,
	tracking entities.PaymentReceiptTracking,
	now time.Time,
	nextPollAt time.Time,
) error {
	return uc.unitOfWork.WithinTransaction(ctx, func(txRepositories outport.TxRepositories) error {
		repository := txRepositories.PaymentReceiptTracking
		if repository == nil {
			return errors.New("payment receipt tracking repository is not configured")
		}
		return repository.SaveObservation(ctx, tracking, now, nextPollAt)
	})
}

func resolveReceiptPollingScope(rawChain string, rawNetwork string) (string, string, error) {
	var (
		chainFilter   string
		networkFilter string
	)

	if rawChain != "" {
		chain, ok := value_objects.ParseChainID(rawChain)
		if !ok {
			return "", "", errors.New("poll chain is invalid")
		}
		chainFilter = string(chain)
	}

	if rawNetwork != "" {
		network, ok := value_objects.ParseNetworkID(rawNetwork)
		if !ok {
			return "", "", errors.New("poll network is invalid")
		}
		networkFilter = string(network)
	}

	if networkFilter != "" && chainFilter == "" {
		return "", "", errors.New("poll chain is required when poll network is set")
	}

	return chainFilter, networkFilter, nil
}
