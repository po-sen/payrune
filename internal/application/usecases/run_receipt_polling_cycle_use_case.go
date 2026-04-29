package usecases

import (
	"context"
	"errors"
	"time"

	inport "payrune/internal/application/ports/inbound"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/valueobjects"
)

type runReceiptPollingCycleUseCase struct {
	unitOfWork outport.UnitOfWork
	observer   outport.BlockchainReceiptObserver
	clock      outport.Clock
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
) inport.RunReceiptPollingCycleUseCase {
	return &runReceiptPollingCycleUseCase{
		unitOfWork: unitOfWork,
		observer:   observer,
		clock:      clock,
	}
}

func pollablePaymentReceiptStatusStrings() []string {
	return []string{
		string(valueobjects.PaymentReceiptStatusWatching),
		string(valueobjects.PaymentReceiptStatusPartiallyPaid),
		string(valueobjects.PaymentReceiptStatusPaidUnconfirmed),
		string(valueobjects.PaymentReceiptStatusPaidUnconfirmedReverted),
	}
}

func (uc *runReceiptPollingCycleUseCase) Execute(
	ctx context.Context,
	input inport.RunReceiptPollingCycleInput,
) (inport.RunReceiptPollingCycleOutput, error) {
	if uc.unitOfWork == nil {
		return inport.RunReceiptPollingCycleOutput{}, inport.ErrUnitOfWorkNotConfigured
	}
	if uc.observer == nil {
		return inport.RunReceiptPollingCycleOutput{}, inport.ErrBlockchainReceiptObserverNotConfigured
	}
	if uc.clock == nil {
		return inport.RunReceiptPollingCycleOutput{}, inport.ErrClockNotConfigured
	}
	if input.BatchSize <= 0 {
		return inport.RunReceiptPollingCycleOutput{}, inport.ErrBatchSizeMustBeGreaterThanZero
	}

	now := uc.clock.NowUTC()
	if input.Network != "" && input.Chain == "" {
		return inport.RunReceiptPollingCycleOutput{}, inport.ErrPollChainRequiredWhenPollNetworkSet
	}
	if input.Chain != "" {
		chain, ok := valueobjects.ParseChainID(input.Chain)
		if !ok {
			return inport.RunReceiptPollingCycleOutput{}, inport.ErrChainNotSupported
		}
		input.Chain = string(chain)
	}
	if input.Network != "" {
		network, ok := valueobjects.ParseNetworkID(input.Network)
		if !ok {
			return inport.RunReceiptPollingCycleOutput{}, inport.ErrChainNotSupported
		}
		input.Network = string(network)
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
			Chain:      input.Chain,
			Network:    input.Network,
			Statuses:   pollablePaymentReceiptStatusStrings(),
		})
		if err != nil {
			return inport.ErrDependencyFailure
		}

		trackings = make([]entities.PaymentReceiptTracking, 0, len(claimedTrackings))
		for _, trackingRecord := range claimedTrackings {
			tracking, err := paymentReceiptTrackingFromRecord(trackingRecord)
			if err != nil {
				return inport.ErrInternalFailure
			}
			trackings = append(trackings, tracking)
		}
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, inport.ErrPaymentReceiptTrackingStoreNotConfigured):
			return inport.RunReceiptPollingCycleOutput{}, err
		case errors.Is(err, inport.ErrDependencyFailure):
			return inport.RunReceiptPollingCycleOutput{}, err
		case errors.Is(err, inport.ErrInternalFailure):
			return inport.RunReceiptPollingCycleOutput{}, err
		default:
			return inport.RunReceiptPollingCycleOutput{}, inport.ErrDependencyFailure
		}
	}

	output := inport.RunReceiptPollingCycleOutput{ClaimedCount: len(trackings)}
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
		if err := uc.savePollingFailure(
			ctx,
			tracking,
			valueobjects.PaymentReceiptTrackingFailureReasonTrackingInvalid,
			now,
			nextPollAt,
		); err != nil {
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
		if saveErr := uc.savePollingFailure(
			ctx,
			tracking,
			valueobjects.PaymentReceiptTrackingFailureReasonLatestBlockHeightUnavailable,
			now,
			nextPollAt,
		); saveErr != nil {
			return 0, saveErr
		}
		return receiptPollingTrackingProcessingError, nil
	}

	observation, err := uc.observer.ObserveAddress(ctx, outport.ObserveChainPaymentAddressInput{
		AssetReference:        tracking.AssetReference,
		Chain:                 string(tracking.Chain),
		Network:               string(tracking.Network),
		Address:               tracking.Address,
		IssuedAt:              tracking.IssuedAt,
		ObservedTotalMinor:    tracking.ObservedTotalMinor,
		ConfirmedTotalMinor:   tracking.ConfirmedTotalMinor,
		UnconfirmedTotalMinor: tracking.UnconfirmedTotalMinor,
		RequiredConfirmations: tracking.RequiredConfirmations,
		LatestBlockHeight:     latestBlockHeight,
		SinceBlockHeight:      tracking.LastObservedBlockHeight,
	})
	if err != nil {
		if saveErr := uc.savePollingFailure(
			ctx,
			tracking,
			valueobjects.PaymentReceiptTrackingFailureReasonObservationFailed,
			now,
			nextPollAt,
		); saveErr != nil {
			return 0, saveErr
		}
		return receiptPollingTrackingProcessingError, nil
	}

	updatedTracking, err := tracking.ApplyObservation(valueobjects.PaymentReceiptObservation{
		ObservedTotalMinor:    observation.ObservedTotalMinor,
		ConfirmedTotalMinor:   observation.ConfirmedTotalMinor,
		UnconfirmedTotalMinor: observation.UnconfirmedTotalMinor,
		LatestBlockHeight:     observation.LatestBlockHeight,
	}, now)
	if err != nil {
		if saveErr := uc.savePollingFailure(
			ctx,
			tracking,
			valueobjects.PaymentReceiptTrackingFailureReasonTrackingUpdateFailed,
			now,
			nextPollAt,
		); saveErr != nil {
			return 0, saveErr
		}
		return receiptPollingTrackingProcessingError, nil
	}

	finalTracking, expired, err := updatedTracking.ExpireIfDue(now)
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

func (uc *runReceiptPollingCycleUseCase) savePollingFailure(
	ctx context.Context,
	tracking entities.PaymentReceiptTracking,
	reason valueobjects.PaymentReceiptTrackingFailureReason,
	now time.Time,
	nextPollAt time.Time,
) error {
	trackingWithFailureReason, err := tracking.MarkPollingFailure(reason)
	if err != nil {
		return inport.ErrInternalFailure
	}

	err = uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
		trackingStore := txScope.PaymentReceiptTracking
		if trackingStore == nil {
			return inport.ErrPaymentReceiptTrackingStoreNotConfigured
		}
		if err := trackingStore.Save(ctx, paymentReceiptTrackingRecordFromDomain(trackingWithFailureReason), now, nextPollAt); err != nil {
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
		if err := trackingStore.Save(ctx, paymentReceiptTrackingRecordFromDomain(tracking), now, nextPollAt); err != nil {
			return inport.ErrDependencyFailure
		}
		if !statusChanged {
			return nil
		}
		notificationOutbox := txScope.PaymentReceiptStatusNotificationOutbox
		if notificationOutbox == nil {
			return inport.ErrPaymentReceiptStatusOutboxNotConfigured
		}
		if err := notificationOutbox.EnqueueStatusChanged(ctx, paymentReceiptStatusChangedRecordFromDomain(statusChangedEvent)); err != nil {
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

	latestBlockHeight, err := uc.observer.FetchLatestBlockHeight(ctx, string(tracking.Chain), string(tracking.Network))
	if err != nil {
		errCache[scopeKey] = err
		return 0, err
	}

	cache[scopeKey] = latestBlockHeight
	return latestBlockHeight, nil
}
