package use_cases

import (
	"context"
	"errors"
	"time"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/in"
	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/entities"
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

	now := uc.clock.NowUTC()

	var claimedNotifications []entities.PaymentReceiptStatusNotification
	err := uc.unitOfWork.WithinTransaction(ctx, func(txRepositories outport.TxRepositories) error {
		repository := txRepositories.PaymentReceiptStatusNotification
		if repository == nil {
			return errors.New("payment receipt status notification repository is not configured")
		}

		notifications, err := repository.ClaimPending(ctx, outport.ClaimPaymentReceiptStatusNotificationsInput{
			Now:        now,
			Limit:      input.BatchSize,
			ClaimUntil: now.Add(claimTTL),
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
			ConflictTotalMinor:    notification.ConflictTotalMinor,
			StatusChangedAt:       notification.StatusChangedAt,
		})
		if err != nil {
			attempts := notification.DeliveryAttempts + 1
			saveErr := uc.unitOfWork.WithinTransaction(ctx, func(txRepositories outport.TxRepositories) error {
				repository := txRepositories.PaymentReceiptStatusNotification
				if repository == nil {
					return errors.New("payment receipt status notification repository is not configured")
				}
				if attempts >= input.MaxAttempts {
					return repository.MarkFailed(ctx, outport.MarkPaymentReceiptStatusNotificationFailureInput{
						NotificationID: notification.NotificationID,
						Attempts:       attempts,
						LastError:      err.Error(),
					})
				}
				return repository.MarkRetryScheduled(ctx, outport.MarkPaymentReceiptStatusNotificationRetryInput{
					NotificationID: notification.NotificationID,
					Attempts:       attempts,
					LastError:      err.Error(),
					NextAttemptAt:  now.Add(input.RetryDelay),
				})
			})
			if saveErr != nil {
				return output, saveErr
			}
			if attempts >= input.MaxAttempts {
				output.FailedCount++
			} else {
				output.RetriedCount++
			}
			continue
		}

		if err := uc.unitOfWork.WithinTransaction(ctx, func(txRepositories outport.TxRepositories) error {
			repository := txRepositories.PaymentReceiptStatusNotification
			if repository == nil {
				return errors.New("payment receipt status notification repository is not configured")
			}
			return repository.MarkSent(ctx, notification.NotificationID, now)
		}); err != nil {
			return output, err
		}
		output.SentCount++
	}

	return output, nil
}
