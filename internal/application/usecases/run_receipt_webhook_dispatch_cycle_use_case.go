package usecases

import (
	"context"
	"errors"
	"strings"

	"payrune/internal/application/dto"
	applicationoutbox "payrune/internal/application/outbox"
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
	input dto.RunReceiptWebhookDispatchCycleInput,
) (dto.RunReceiptWebhookDispatchCycleOutput, error) {
	if uc.unitOfWork == nil {
		return dto.RunReceiptWebhookDispatchCycleOutput{}, inport.ErrUnitOfWorkNotConfigured
	}
	if uc.notifier == nil {
		return dto.RunReceiptWebhookDispatchCycleOutput{}, inport.ErrPaymentReceiptStatusNotifierNotConfigured
	}
	if uc.policyReader == nil {
		return dto.RunReceiptWebhookDispatchCycleOutput{}, inport.ErrAddressPolicyReaderNotConfigured
	}
	if uc.clock == nil {
		return dto.RunReceiptWebhookDispatchCycleOutput{}, inport.ErrClockNotConfigured
	}
	if input.BatchSize <= 0 {
		return dto.RunReceiptWebhookDispatchCycleOutput{}, inport.ErrBatchSizeMustBeGreaterThanZero
	}
	if input.MaxAttempts <= 0 {
		return dto.RunReceiptWebhookDispatchCycleOutput{}, inport.ErrMaxAttemptsMustBeGreaterThanZero
	}
	if input.RetryDelay <= 0 {
		return dto.RunReceiptWebhookDispatchCycleOutput{}, inport.ErrRetryDelayMustBeGreaterThanZero
	}

	claimNow := uc.clock.NowUTC()

	var claimedNotifications []applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage
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
			return dto.RunReceiptWebhookDispatchCycleOutput{}, err
		case errors.Is(err, inport.ErrDependencyFailure):
			return dto.RunReceiptWebhookDispatchCycleOutput{}, err
		default:
			return dto.RunReceiptWebhookDispatchCycleOutput{}, inport.ErrDependencyFailure
		}
	}

	output := dto.RunReceiptWebhookDispatchCycleOutput{ClaimedCount: len(claimedNotifications)}

	for _, notification := range claimedNotifications {
		outcome, err := uc.processNotification(ctx, notification, input)
		if err != nil {
			return output, err
		}
		switch outcome {
		case applicationoutbox.PaymentReceiptNotificationDeliveryStatusSent:
			output.SentCount++
		case applicationoutbox.PaymentReceiptNotificationDeliveryStatusPending:
			output.RetriedCount++
		case applicationoutbox.PaymentReceiptNotificationDeliveryStatusFailed:
			output.FailedCount++
		}
	}

	return output, nil
}

func (uc *runReceiptWebhookDispatchCycleUseCase) processNotification(
	ctx context.Context,
	notification applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage,
	input dto.RunReceiptWebhookDispatchCycleInput,
) (applicationoutbox.PaymentReceiptNotificationDeliveryStatus, error) {
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
		PreviousStatus:        string(notification.PreviousStatus),
		CurrentStatus:         string(notification.CurrentStatus),
		ObservedTotalMinor:    notification.ObservedTotalMinor,
		ConfirmedTotalMinor:   notification.ConfirmedTotalMinor,
		UnconfirmedTotalMinor: notification.UnconfirmedTotalMinor,
		StatusChangedAt:       notification.StatusChangedAt,
	})
	if err != nil {
		deliveryResult, resultErr := applicationoutbox.ResolvePaymentReceiptStatusNotificationDeliveryFailure(
			notification.NotificationID,
			notification.DeliveryAttempts,
			input.MaxAttempts,
			uc.clock.NowUTC(),
			input.RetryDelay,
			applicationoutbox.PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed,
		)
		if resultErr != nil {
			return "", inport.ErrInternalFailure
		}
		if err := uc.saveDeliveryResult(ctx, deliveryResult); err != nil {
			return "", err
		}
		return deliveryResult.Status, nil
	}

	deliveryResult, err := applicationoutbox.MarkPaymentReceiptStatusNotificationSent(
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
	deliveryResult applicationoutbox.PaymentReceiptStatusNotificationDeliveryResult,
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
