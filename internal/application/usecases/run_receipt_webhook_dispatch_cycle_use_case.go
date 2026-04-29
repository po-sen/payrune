package usecases

import (
	"context"
	"errors"
	"strings"

	inport "payrune/internal/application/ports/inbound"
	outport "payrune/internal/application/ports/outbound"
)

type runReceiptWebhookDispatchCycleUseCase struct {
	unitOfWork   outport.UnitOfWork
	policyReader outport.AddressPolicyReader
	notifier     outport.PaymentReceiptStatusNotifier
	clock        outport.Clock
}

func NewRunReceiptWebhookDispatchCycleUseCase(
	unitOfWork outport.UnitOfWork,
	policyReader outport.AddressPolicyReader,
	notifier outport.PaymentReceiptStatusNotifier,
	clock outport.Clock,
) inport.RunReceiptWebhookDispatchCycleUseCase {
	return &runReceiptWebhookDispatchCycleUseCase{
		unitOfWork:   unitOfWork,
		policyReader: policyReader,
		notifier:     notifier,
		clock:        clock,
	}
}

func (uc *runReceiptWebhookDispatchCycleUseCase) Execute(
	ctx context.Context,
	input inport.RunReceiptWebhookDispatchCycleInput,
) (inport.RunReceiptWebhookDispatchCycleOutput, error) {
	if uc.unitOfWork == nil {
		return inport.RunReceiptWebhookDispatchCycleOutput{}, inport.ErrUnitOfWorkNotConfigured
	}
	if uc.notifier == nil {
		return inport.RunReceiptWebhookDispatchCycleOutput{}, inport.ErrPaymentReceiptStatusNotifierNotConfigured
	}
	if uc.policyReader == nil {
		return inport.RunReceiptWebhookDispatchCycleOutput{}, inport.ErrAddressPolicyReaderNotConfigured
	}
	if uc.clock == nil {
		return inport.RunReceiptWebhookDispatchCycleOutput{}, inport.ErrClockNotConfigured
	}
	if input.BatchSize <= 0 {
		return inport.RunReceiptWebhookDispatchCycleOutput{}, inport.ErrBatchSizeMustBeGreaterThanZero
	}
	if input.MaxAttempts <= 0 {
		return inport.RunReceiptWebhookDispatchCycleOutput{}, inport.ErrMaxAttemptsMustBeGreaterThanZero
	}
	if input.RetryDelay <= 0 {
		return inport.RunReceiptWebhookDispatchCycleOutput{}, inport.ErrRetryDelayMustBeGreaterThanZero
	}

	claimNow := uc.clock.NowUTC()

	var claimedNotifications []outport.PaymentReceiptStatusNotificationOutboxMessage
	err := uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
		outbox := txScope.PaymentReceiptStatusNotificationOutbox
		if outbox == nil {
			return inport.ErrPaymentReceiptStatusOutboxNotConfigured
		}

		notifications, err := outbox.ClaimPending(ctx, outport.ClaimPaymentReceiptStatusNotificationsInput{
			Now:        claimNow,
			Limit:      input.BatchSize,
			ClaimUntil: claimNow.Add(input.DispatchTTL),
		})
		if err != nil {
			return inport.ErrDependencyFailure
		}

		claimedNotifications = notifications
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, inport.ErrPaymentReceiptStatusOutboxNotConfigured):
			return inport.RunReceiptWebhookDispatchCycleOutput{}, err
		case errors.Is(err, inport.ErrDependencyFailure):
			return inport.RunReceiptWebhookDispatchCycleOutput{}, err
		default:
			return inport.RunReceiptWebhookDispatchCycleOutput{}, inport.ErrDependencyFailure
		}
	}

	output := inport.RunReceiptWebhookDispatchCycleOutput{ClaimedCount: len(claimedNotifications)}

	for _, notification := range claimedNotifications {
		outcome, err := uc.processNotification(ctx, notification, input)
		if err != nil {
			return output, err
		}
		switch outcome {
		case outport.PaymentReceiptNotificationDeliveryStatusSent:
			output.SentCount++
		case outport.PaymentReceiptNotificationDeliveryStatusPending:
			output.RetriedCount++
		case outport.PaymentReceiptNotificationDeliveryStatusFailed:
			output.FailedCount++
		}
	}

	return output, nil
}

func (uc *runReceiptWebhookDispatchCycleUseCase) processNotification(
	ctx context.Context,
	notification outport.PaymentReceiptStatusNotificationOutboxMessage,
	input inport.RunReceiptWebhookDispatchCycleInput,
) (string, error) {
	policy, ok, err := uc.policyReader.FindIssuanceByID(ctx, notification.AddressPolicyID)
	if err != nil {
		return "", inport.ErrDependencyFailure
	}
	if !ok {
		return "", inport.ErrPaymentAddressPolicyNotConfigured
	}

	err = uc.notifier.NotifyStatusChanged(ctx, outport.NotifyPaymentReceiptStatusChangedInput{
		NotificationID:        notification.NotificationID,
		PaymentAddressID:      notification.PaymentAddressID,
		CustomerReference:     notification.CustomerReference,
		AssetReference:        strings.TrimSpace(policy.AssetReference),
		PreviousStatus:        notification.PreviousStatus,
		CurrentStatus:         notification.CurrentStatus,
		ObservedTotalMinor:    notification.ObservedTotalMinor,
		ConfirmedTotalMinor:   notification.ConfirmedTotalMinor,
		UnconfirmedTotalMinor: notification.UnconfirmedTotalMinor,
		StatusChangedAt:       notification.StatusChangedAt,
	})
	if err != nil {
		deliveryResult, resultErr := outport.ResolvePaymentReceiptStatusNotificationDeliveryFailure(
			notification.NotificationID,
			notification.DeliveryAttempts,
			input.MaxAttempts,
			uc.clock.NowUTC(),
			input.RetryDelay,
			outport.PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed,
		)
		if resultErr != nil {
			return "", inport.ErrInternalFailure
		}
		if err := uc.saveDeliveryResult(ctx, deliveryResult); err != nil {
			return "", err
		}
		return deliveryResult.Status, nil
	}

	deliveryResult, err := outport.MarkPaymentReceiptStatusNotificationSent(
		notification.NotificationID,
		uc.clock.NowUTC(),
	)
	if err != nil {
		return "", inport.ErrInternalFailure
	}
	if err := uc.saveDeliveryResult(ctx, deliveryResult); err != nil {
		return "", err
	}
	return deliveryResult.Status, nil
}

func (uc *runReceiptWebhookDispatchCycleUseCase) saveDeliveryResult(
	ctx context.Context,
	deliveryResult outport.PaymentReceiptStatusNotificationDeliveryResult,
) error {
	err := uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
		outbox := txScope.PaymentReceiptStatusNotificationOutbox
		if outbox == nil {
			return inport.ErrPaymentReceiptStatusOutboxNotConfigured
		}
		if err := outbox.SaveDeliveryResult(ctx, deliveryResult); err != nil {
			return inport.ErrDependencyFailure
		}
		return nil
	})
	if err != nil {
		switch {
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
