package use_cases

import (
	"context"
	"errors"
	"time"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/in"
	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/value_objects"
)

const (
	defaultReceiptPollingInterval = 15 * time.Second
	defaultReceiptPollingClaimTTL = 30 * time.Second
)

type runReceiptPollingCycleUseCase struct {
	unitOfWork      outport.UnitOfWork
	observer        outport.BlockchainReceiptObserver
	clock           outport.Clock
	lifecyclePolicy policies.PaymentReceiptTrackingLifecyclePolicy
}

func NewRunReceiptPollingCycleUseCase(
	unitOfWork outport.UnitOfWork,
	observer outport.BlockchainReceiptObserver,
	clock outport.Clock,
	lifecyclePolicy policies.PaymentReceiptTrackingLifecyclePolicy,
) inport.RunReceiptPollingCycleUseCase {
	return &runReceiptPollingCycleUseCase{
		unitOfWork:      unitOfWork,
		observer:        observer,
		clock:           clock,
		lifecyclePolicy: lifecyclePolicy,
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
	err = uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
		trackingStore := txScope.PaymentReceiptTracking
		if trackingStore == nil {
			return errors.New("payment receipt tracking store is not configured")
		}

		claimedTrackings, err := trackingStore.ClaimDue(ctx, outport.ClaimPaymentReceiptTrackingsInput{
			Now:        now,
			Limit:      input.BatchSize,
			ClaimUntil: now.Add(claimTTL),
			Chain:      chainFilter,
			Network:    networkFilter,
			Statuses:   entities.PollablePaymentReceiptStatuses(),
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
			if err := uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
				trackingStore := txScope.PaymentReceiptTracking
				if trackingStore == nil {
					return errors.New("payment receipt tracking store is not configured")
				}
				return trackingStore.Save(ctx, trackingWithError, now, now.Add(pollInterval))
			}); err != nil {
				return output, err
			}
			output.ProcessingErrorCount++
			continue
		}
		expiredTracking, expired, err := uc.lifecyclePolicy.ExpireIfDue(tracking, now)
		if err != nil {
			return output, err
		}
		if expired {
			statusChangedEvent, statusChanged, err := expiredTracking.StatusChangedEvent(tracking.Status, now)
			if err != nil {
				return output, err
			}
			if err := uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
				trackingStore := txScope.PaymentReceiptTracking
				if trackingStore == nil {
					return errors.New("payment receipt tracking store is not configured")
				}
				if err := trackingStore.Save(
					ctx,
					expiredTracking,
					now,
					now.Add(pollInterval),
				); err != nil {
					return err
				}
				if !statusChanged {
					return nil
				}
				notificationOutbox := txScope.PaymentReceiptStatusNotificationOutbox
				if notificationOutbox == nil {
					return errors.New("payment receipt status notification outbox is not configured")
				}
				return notificationOutbox.EnqueueStatusChanged(ctx, statusChangedEvent)
			}); err != nil {
				return output, err
			}
			output.TerminalFailedCount++
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
			if err := uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
				trackingStore := txScope.PaymentReceiptTracking
				if trackingStore == nil {
					return errors.New("payment receipt tracking store is not configured")
				}
				return trackingStore.Save(ctx, trackingWithError, now, now.Add(pollInterval))
			}); err != nil {
				return output, err
			}
			output.ProcessingErrorCount++
			continue
		}

		updatedTracking, err := uc.lifecyclePolicy.ApplyObservation(tracking, value_objects.PaymentReceiptObservation{
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
			if err := uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
				trackingStore := txScope.PaymentReceiptTracking
				if trackingStore == nil {
					return errors.New("payment receipt tracking store is not configured")
				}
				return trackingStore.Save(ctx, trackingWithError, now, now.Add(pollInterval))
			}); err != nil {
				return output, err
			}
			output.ProcessingErrorCount++
			continue
		}
		statusChangedEvent, statusChanged, err := updatedTracking.StatusChangedEvent(tracking.Status, now)
		if err != nil {
			return output, err
		}

		nextPollAt := now.Add(pollInterval)

		if err := uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
			trackingStore := txScope.PaymentReceiptTracking
			if trackingStore == nil {
				return errors.New("payment receipt tracking store is not configured")
			}
			if err := trackingStore.Save(ctx, updatedTracking, now, nextPollAt); err != nil {
				return err
			}
			if !statusChanged {
				return nil
			}
			notificationOutbox := txScope.PaymentReceiptStatusNotificationOutbox
			if notificationOutbox == nil {
				return errors.New("payment receipt status notification outbox is not configured")
			}
			return notificationOutbox.EnqueueStatusChanged(ctx, statusChangedEvent)
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
