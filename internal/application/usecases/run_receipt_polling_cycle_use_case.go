package usecases

import (
	"context"
	"errors"
	"time"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
)

type runReceiptPollingCycleUseCase struct {
	unitOfWork      outport.UnitOfWork
	observer        outport.BlockchainReceiptObserver
	clock           outport.Clock
	lifecyclePolicy policies.PaymentReceiptTrackingLifecyclePolicy
}

type receiptPollingScopeKey struct {
	chain   valueobjects.ChainID
	network valueobjects.NetworkID
}

type receiptPollingTrackingResult int

const (
	receiptPollingTrackingUpdated receiptPollingTrackingResult = iota + 1
	receiptPollingTrackingTerminalFailed
	receiptPollingTrackingProcessingError
)

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
		return dto.RunReceiptPollingCycleOutput{}, inport.ErrUnitOfWorkNotConfigured
	}
	if uc.observer == nil {
		return dto.RunReceiptPollingCycleOutput{}, inport.ErrBlockchainReceiptObserverNotConfigured
	}
	if uc.clock == nil {
		return dto.RunReceiptPollingCycleOutput{}, inport.ErrClockNotConfigured
	}
	if input.BatchSize <= 0 {
		return dto.RunReceiptPollingCycleOutput{}, inport.ErrBatchSizeMustBeGreaterThanZero
	}

	now := uc.clock.NowUTC()
	if input.Network != "" && input.Chain == "" {
		return dto.RunReceiptPollingCycleOutput{}, inport.ErrPollChainRequiredWhenPollNetworkSet
	}

	var trackings []entities.PaymentReceiptTracking
	err := uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
		trackingStore := txScope.PaymentReceiptTracking
		if trackingStore == nil {
			return inport.ErrPaymentReceiptTrackingStoreNotConfigured
		}

		claimedTrackings, err := trackingStore.ClaimDue(ctx, outport.ClaimPaymentReceiptTrackingsInput{
			Now:        now,
			Limit:      input.BatchSize,
			ClaimUntil: now.Add(input.ClaimTTL),
			Chain:      string(input.Chain),
			Network:    string(input.Network),
			Statuses:   entities.PollablePaymentReceiptStatuses(),
		})
		if err != nil {
			return inport.ErrDependencyFailure
		}

		trackings = claimedTrackings
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, inport.ErrPaymentReceiptTrackingStoreNotConfigured):
			return dto.RunReceiptPollingCycleOutput{}, err
		case errors.Is(err, inport.ErrDependencyFailure):
			return dto.RunReceiptPollingCycleOutput{}, err
		default:
			return dto.RunReceiptPollingCycleOutput{}, inport.ErrDependencyFailure
		}
	}

	output := dto.RunReceiptPollingCycleOutput{ClaimedCount: len(trackings)}
	latestBlockHeights := make(map[receiptPollingScopeKey]int64)
	latestBlockHeightErrs := make(map[receiptPollingScopeKey]error)
	nextPollAt := now.Add(input.RescheduleInterval)

	for _, tracking := range trackings {
		result, err := uc.processTracking(
			ctx,
			tracking,
			now,
			nextPollAt,
			latestBlockHeights,
			latestBlockHeightErrs,
		)
		if err != nil {
			return output, err
		}
		switch result {
		case receiptPollingTrackingUpdated:
			output.UpdatedCount++
		case receiptPollingTrackingTerminalFailed:
			output.TerminalFailedCount++
		case receiptPollingTrackingProcessingError:
			output.ProcessingErrorCount++
		}
	}

	return output, nil
}

func (uc *runReceiptPollingCycleUseCase) processTracking(
	ctx context.Context,
	tracking entities.PaymentReceiptTracking,
	now time.Time,
	nextPollAt time.Time,
	latestBlockHeights map[receiptPollingScopeKey]int64,
	latestBlockHeightErrs map[receiptPollingScopeKey]error,
) (receiptPollingTrackingResult, error) {
	if tracking.IssuedAt.IsZero() {
		if err := uc.savePollingError(ctx, tracking, "issued at is required", now, nextPollAt); err != nil {
			return 0, err
		}
		return receiptPollingTrackingProcessingError, nil
	}

	latestBlockHeight, err := uc.fetchLatestBlockHeightForTracking(
		ctx,
		tracking,
		latestBlockHeights,
		latestBlockHeightErrs,
	)
	if err != nil {
		if saveErr := uc.savePollingError(ctx, tracking, err.Error(), now, nextPollAt); saveErr != nil {
			return 0, saveErr
		}
		return receiptPollingTrackingProcessingError, nil
	}

	observation, err := uc.observer.ObserveAddress(ctx, outport.ObserveChainPaymentAddressInput{
		Chain:                 tracking.Chain,
		Network:               tracking.Network,
		Address:               tracking.Address,
		IssuedAt:              tracking.IssuedAt,
		RequiredConfirmations: tracking.RequiredConfirmations,
		LatestBlockHeight:     latestBlockHeight,
		SinceBlockHeight:      tracking.LastObservedBlockHeight,
	})
	if err != nil {
		if saveErr := uc.savePollingError(ctx, tracking, err.Error(), now, nextPollAt); saveErr != nil {
			return 0, saveErr
		}
		return receiptPollingTrackingProcessingError, nil
	}

	updatedTracking, err := uc.lifecyclePolicy.ApplyObservation(tracking, valueobjects.PaymentReceiptObservation{
		ObservedTotalMinor:    observation.ObservedTotalMinor,
		ConfirmedTotalMinor:   observation.ConfirmedTotalMinor,
		UnconfirmedTotalMinor: observation.UnconfirmedTotalMinor,
		LatestBlockHeight:     observation.LatestBlockHeight,
	}, now)
	if err != nil {
		if saveErr := uc.savePollingError(ctx, tracking, err.Error(), now, nextPollAt); saveErr != nil {
			return 0, saveErr
		}
		return receiptPollingTrackingProcessingError, nil
	}

	finalTracking, expired, err := uc.lifecyclePolicy.ExpireIfDue(updatedTracking, now)
	if err != nil {
		return 0, inport.ErrInternalFailure
	}
	if expired {
		if err := uc.saveTrackingAndMaybeEnqueueStatusChanged(
			ctx,
			finalTracking,
			tracking.Status,
			now,
			nextPollAt,
		); err != nil {
			return 0, err
		}
		return receiptPollingTrackingTerminalFailed, nil
	}

	if err := uc.saveTrackingAndMaybeEnqueueStatusChanged(
		ctx,
		updatedTracking,
		tracking.Status,
		now,
		nextPollAt,
	); err != nil {
		return 0, err
	}
	return receiptPollingTrackingUpdated, nil
}

func (uc *runReceiptPollingCycleUseCase) savePollingError(
	ctx context.Context,
	tracking entities.PaymentReceiptTracking,
	reason string,
	now time.Time,
	nextPollAt time.Time,
) error {
	trackingWithError, err := tracking.MarkPollingError(reason)
	if err != nil {
		return inport.ErrInternalFailure
	}

	err = uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
		trackingStore := txScope.PaymentReceiptTracking
		if trackingStore == nil {
			return inport.ErrPaymentReceiptTrackingStoreNotConfigured
		}
		if err := trackingStore.Save(ctx, trackingWithError, now, nextPollAt); err != nil {
			return inport.ErrDependencyFailure
		}
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, inport.ErrPaymentReceiptTrackingStoreNotConfigured):
			return err
		case errors.Is(err, inport.ErrDependencyFailure):
			return err
		default:
			return inport.ErrDependencyFailure
		}
	}
	return nil
}

func (uc *runReceiptPollingCycleUseCase) saveTrackingAndMaybeEnqueueStatusChanged(
	ctx context.Context,
	tracking entities.PaymentReceiptTracking,
	previousStatus valueobjects.PaymentReceiptStatus,
	now time.Time,
	nextPollAt time.Time,
) error {
	statusChangedEvent, statusChanged, err := tracking.StatusChangedEvent(previousStatus, now)
	if err != nil {
		return inport.ErrInternalFailure
	}

	err = uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
		trackingStore := txScope.PaymentReceiptTracking
		if trackingStore == nil {
			return inport.ErrPaymentReceiptTrackingStoreNotConfigured
		}
		if err := trackingStore.Save(ctx, tracking, now, nextPollAt); err != nil {
			return inport.ErrDependencyFailure
		}
		if !statusChanged {
			return nil
		}
		notificationOutbox := txScope.PaymentReceiptStatusNotificationOutbox
		if notificationOutbox == nil {
			return inport.ErrPaymentReceiptStatusOutboxNotConfigured
		}
		if err := notificationOutbox.EnqueueStatusChanged(ctx, statusChangedEvent); err != nil {
			return inport.ErrDependencyFailure
		}
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, inport.ErrPaymentReceiptTrackingStoreNotConfigured):
			return err
		case errors.Is(err, inport.ErrPaymentReceiptStatusOutboxNotConfigured):
			return err
		case errors.Is(err, inport.ErrDependencyFailure):
			return err
		default:
			return inport.ErrDependencyFailure
		}
	}
	return nil
}

func (uc *runReceiptPollingCycleUseCase) fetchLatestBlockHeightForTracking(
	ctx context.Context,
	tracking entities.PaymentReceiptTracking,
	cache map[receiptPollingScopeKey]int64,
	errCache map[receiptPollingScopeKey]error,
) (int64, error) {
	scopeKey := receiptPollingScopeKey{
		chain:   tracking.Chain,
		network: tracking.Network,
	}

	if latestBlockHeight, found := cache[scopeKey]; found {
		return latestBlockHeight, nil
	}
	if err, found := errCache[scopeKey]; found {
		return 0, err
	}

	latestBlockHeight, err := uc.observer.FetchLatestBlockHeight(ctx, tracking.Chain, tracking.Network)
	if err != nil {
		errCache[scopeKey] = err
		return 0, err
	}

	cache[scopeKey] = latestBlockHeight
	return latestBlockHeight, nil
}
