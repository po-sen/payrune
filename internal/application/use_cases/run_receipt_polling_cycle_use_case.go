package use_cases

import (
	"context"
	"errors"
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
	expiredReceiptReason          = "payment window expired"

	defaultReceiptPaidUnconfirmedExpiryExtension = 7 * 24 * time.Hour
)

type runReceiptPollingCycleUseCase struct {
	unitOfWork                     outport.UnitOfWork
	observer                       outport.BlockchainReceiptObserver
	clock                          outport.Clock
	paidUnconfirmedExpiryExtension time.Duration
}

type RunReceiptPollingCycleUseCaseConfig struct {
	PaidUnconfirmedExpiryExtension time.Duration
}

func NewRunReceiptPollingCycleUseCase(
	unitOfWork outport.UnitOfWork,
	observer outport.BlockchainReceiptObserver,
	clock outport.Clock,
	config RunReceiptPollingCycleUseCaseConfig,
) inport.RunReceiptPollingCycleUseCase {
	paidUnconfirmedExpiryExtension := defaultReceiptPaidUnconfirmedExpiryExtension
	if config.PaidUnconfirmedExpiryExtension > 0 {
		paidUnconfirmedExpiryExtension = config.PaidUnconfirmedExpiryExtension
	}

	return &runReceiptPollingCycleUseCase{
		unitOfWork:                     unitOfWork,
		observer:                       observer,
		clock:                          clock,
		paidUnconfirmedExpiryExtension: paidUnconfirmedExpiryExtension,
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
			trackingWithError, markErr := tracking.MarkPollingError("issued at is required")
			if markErr != nil {
				return output, markErr
			}
			if err := uc.unitOfWork.WithinTransaction(ctx, func(txRepositories outport.TxRepositories) error {
				repository := txRepositories.PaymentReceiptTracking
				if repository == nil {
					return errors.New("payment receipt tracking repository is not configured")
				}
				return repository.SavePollingError(
					ctx,
					trackingWithError.PaymentAddressID,
					trackingWithError.LastError,
					now,
					now.Add(pollInterval),
				)
			}); err != nil {
				return output, err
			}
			output.FailedCount++
			continue
		}
		if tracking.IsExpired(now) {
			expiredTracking, err := tracking.MarkExpired(expiredReceiptReason)
			if err != nil {
				return output, err
			}
			if err := uc.unitOfWork.WithinTransaction(ctx, func(txRepositories outport.TxRepositories) error {
				repository := txRepositories.PaymentReceiptTracking
				if repository == nil {
					return errors.New("payment receipt tracking repository is not configured")
				}
				return repository.SaveObservation(
					ctx,
					expiredTracking,
					now,
					now.Add(pollInterval),
				)
			}); err != nil {
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
			trackingWithError, markErr := tracking.MarkPollingError(observeErr.Error())
			if markErr != nil {
				return output, markErr
			}
			if err := uc.unitOfWork.WithinTransaction(ctx, func(txRepositories outport.TxRepositories) error {
				repository := txRepositories.PaymentReceiptTracking
				if repository == nil {
					return errors.New("payment receipt tracking repository is not configured")
				}
				return repository.SavePollingError(
					ctx,
					trackingWithError.PaymentAddressID,
					trackingWithError.LastError,
					now,
					now.Add(pollInterval),
				)
			}); err != nil {
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
			trackingWithError, markErr := tracking.MarkPollingError(err.Error())
			if markErr != nil {
				return output, markErr
			}
			if err := uc.unitOfWork.WithinTransaction(ctx, func(txRepositories outport.TxRepositories) error {
				repository := txRepositories.PaymentReceiptTracking
				if repository == nil {
					return errors.New("payment receipt tracking repository is not configured")
				}
				return repository.SavePollingError(
					ctx,
					trackingWithError.PaymentAddressID,
					trackingWithError.LastError,
					now,
					now.Add(pollInterval),
				)
			}); err != nil {
				return output, err
			}
			output.FailedCount++
			continue
		}
		updatedTracking = updatedTracking.ExtendExpiryOnTransitionToPaidUnconfirmed(
			tracking.Status,
			now,
			uc.paidUnconfirmedExpiryExtension,
		)

		nextPollAt := now.Add(pollInterval)

		if err := uc.unitOfWork.WithinTransaction(ctx, func(txRepositories outport.TxRepositories) error {
			repository := txRepositories.PaymentReceiptTracking
			if repository == nil {
				return errors.New("payment receipt tracking repository is not configured")
			}
			return repository.SaveObservation(ctx, updatedTracking, now, nextPollAt)
		}); err != nil {
			return output, err
		}
		output.UpdatedCount++
	}

	return output, nil
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
