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
	if input.Network != "" && input.Chain == "" {
		return dto.RunReceiptPollingCycleOutput{}, errors.New("poll chain is required when poll network is set")
	}

	var trackings []entities.PaymentReceiptTracking
	err := uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
		trackingStore := txScope.PaymentReceiptTracking
		if trackingStore == nil {
			return errors.New("payment receipt tracking store is not configured")
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
			return err
		}

		trackings = claimedTrackings
		return nil
	})
	if err != nil {
		return dto.RunReceiptPollingCycleOutput{}, err
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
		return 0, err
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
		return err
	}

	return uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
		trackingStore := txScope.PaymentReceiptTracking
		if trackingStore == nil {
			return errors.New("payment receipt tracking store is not configured")
		}
		return trackingStore.Save(ctx, trackingWithError, now, nextPollAt)
	})
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
		return err
	}

	return uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
		trackingStore := txScope.PaymentReceiptTracking
		if trackingStore == nil {
			return errors.New("payment receipt tracking store is not configured")
		}
		if err := trackingStore.Save(ctx, tracking, now, nextPollAt); err != nil {
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
	})
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
