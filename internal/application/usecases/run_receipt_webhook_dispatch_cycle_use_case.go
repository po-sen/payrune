package usecases

import (
	"context"

	"payrune/internal/application/dto"
	applicationoutbox "payrune/internal/application/outbox"
	inport "payrune/internal/application/ports/inbound"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
)

type runReceiptWebhookDispatchCycleUseCase struct {
	unitOfWork outport.UnitOfWork
	notifier   outport.PaymentReceiptStatusNotifier
	clock      outport.Clock
}

func NewRunReceiptWebhookDispatchCycleUseCase(
	unitOfWork outport.UnitOfWork,
	notifier outport.PaymentReceiptStatusNotifier,
	clock outport.Clock,
) inport.RunReceiptWebhookDispatchCycleUseCase {
	return &runReceiptWebhookDispatchCycleUseCase{
		unitOfWork: unitOfWork,
		notifier:   notifier,
		clock:      clock,
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
			return err
		}

		claimedNotifications = notifications
		return nil
	})
	if err != nil {
		return dto.RunReceiptWebhookDispatchCycleOutput{}, err
	}

	output := dto.RunReceiptWebhookDispatchCycleOutput{ClaimedCount: len(claimedNotifications)}

	for _, notification := range claimedNotifications {
		outcome, err := uc.processNotification(ctx, notification, input)
		if err != nil {
			return output, err
		}
		switch outcome {
		case valueobjects.PaymentReceiptNotificationDeliveryStatusSent:
			output.SentCount++
		case valueobjects.PaymentReceiptNotificationDeliveryStatusPending:
			output.RetriedCount++
		case valueobjects.PaymentReceiptNotificationDeliveryStatusFailed:
			output.FailedCount++
		}
	}

	return output, nil
}

func (uc *runReceiptWebhookDispatchCycleUseCase) processNotification(
	ctx context.Context,
	notification applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage,
	input dto.RunReceiptWebhookDispatchCycleInput,
) (valueobjects.PaymentReceiptNotificationDeliveryStatus, error) {
	err := uc.notifier.NotifyStatusChanged(ctx, outport.NotifyPaymentReceiptStatusChangedInput{
		NotificationID:        notification.NotificationID,
		PaymentAddressID:      notification.PaymentAddressID,
		CustomerReference:     notification.CustomerReference,
		PreviousStatus:        string(notification.PreviousStatus),
		CurrentStatus:         string(notification.CurrentStatus),
		ObservedTotalMinor:    notification.ObservedTotalMinor,
		ConfirmedTotalMinor:   notification.ConfirmedTotalMinor,
		UnconfirmedTotalMinor: notification.UnconfirmedTotalMinor,
		StatusChangedAt:       notification.StatusChangedAt,
	})
	if err != nil {
		deliveryResult, resultErr := policies.ResolvePaymentReceiptStatusNotificationDeliveryFailure(
			notification.NotificationID,
			notification.DeliveryAttempts,
			input.MaxAttempts,
			uc.clock.NowUTC(),
			input.RetryDelay,
			err.Error(),
		)
		if resultErr != nil {
			return "", resultErr
		}
		if err := uc.saveDeliveryResult(ctx, deliveryResult); err != nil {
			return "", err
		}
		return deliveryResult.Status, nil
	}

	deliveryResult, err := policies.MarkPaymentReceiptStatusNotificationSent(
		notification.NotificationID,
		uc.clock.NowUTC(),
	)
	if err != nil {
		return "", err
	}
	if err := uc.saveDeliveryResult(ctx, deliveryResult); err != nil {
		return "", err
	}
	return deliveryResult.Status, nil
}

func (uc *runReceiptWebhookDispatchCycleUseCase) saveDeliveryResult(
	ctx context.Context,
	deliveryResult policies.PaymentReceiptStatusNotificationDeliveryResult,
) error {
	return uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
		outbox := txScope.PaymentReceiptStatusNotificationOutbox
		if outbox == nil {
			return inport.ErrPaymentReceiptStatusOutboxNotConfigured
		}
		return outbox.SaveDeliveryResult(ctx, deliveryResult)
	})
}
