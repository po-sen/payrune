package use_cases

import (
	"context"
	"errors"
	"time"

	"payrune/internal/application/dto"
	applicationoutbox "payrune/internal/application/outbox"
	inport "payrune/internal/application/ports/in"
	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/value_objects"
)

const defaultReceiptWebhookDispatchClaimTTL = 30 * time.Second

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
		return dto.RunReceiptWebhookDispatchCycleOutput{}, errors.New("unit of work is not configured")
	}
	if uc.notifier == nil {
		return dto.RunReceiptWebhookDispatchCycleOutput{}, errors.New("payment receipt status notifier is not configured")
	}
	if uc.clock == nil {
		return dto.RunReceiptWebhookDispatchCycleOutput{}, errors.New("clock is not configured")
	}
	if input.BatchSize <= 0 {
		return dto.RunReceiptWebhookDispatchCycleOutput{}, errors.New("batch size must be greater than zero")
	}
	if input.MaxAttempts <= 0 {
		return dto.RunReceiptWebhookDispatchCycleOutput{}, errors.New("max attempts must be greater than zero")
	}
	if input.RetryDelay <= 0 {
		return dto.RunReceiptWebhookDispatchCycleOutput{}, errors.New("retry delay must be greater than zero")
	}

	claimTTL := input.DispatchTTL
	if claimTTL <= 0 {
		claimTTL = defaultReceiptWebhookDispatchClaimTTL
	}

	claimNow := uc.clock.NowUTC()

	var claimedNotifications []applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage
	err := uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
		outbox := txScope.PaymentReceiptStatusNotificationOutbox
		if outbox == nil {
			return errors.New("payment receipt status notification outbox is not configured")
		}

		notifications, err := outbox.ClaimPending(ctx, outport.ClaimPaymentReceiptStatusNotificationsInput{
			Now:        claimNow,
			Limit:      input.BatchSize,
			ClaimUntil: claimNow.Add(claimTTL),
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
				return output, resultErr
			}
			saveErr := uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
				outbox := txScope.PaymentReceiptStatusNotificationOutbox
				if outbox == nil {
					return errors.New("payment receipt status notification outbox is not configured")
				}
				return outbox.SaveDeliveryResult(ctx, deliveryResult)
			})
			if saveErr != nil {
				return output, saveErr
			}
			if deliveryResult.Status == value_objects.PaymentReceiptNotificationDeliveryStatusFailed {
				output.FailedCount++
			} else {
				output.RetriedCount++
			}
			continue
		}

		deliveryResult, resultErr := policies.MarkPaymentReceiptStatusNotificationSent(
			notification.NotificationID,
			uc.clock.NowUTC(),
		)
		if resultErr != nil {
			return output, resultErr
		}

		if err := uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
			outbox := txScope.PaymentReceiptStatusNotificationOutbox
			if outbox == nil {
				return errors.New("payment receipt status notification outbox is not configured")
			}
			return outbox.SaveDeliveryResult(ctx, deliveryResult)
		}); err != nil {
			return output, err
		}
		output.SentCount++
	}

	return output, nil
}
