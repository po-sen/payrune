package policies

import (
	"time"

	"payrune/internal/domain/valueobjects"
)

type PaymentReceiptStatusNotificationDeliveryResult struct {
	NotificationID    int64
	Status            valueobjects.PaymentReceiptNotificationDeliveryStatus
	Attempts          int32
	LastFailureReason valueobjects.PaymentReceiptNotificationDeliveryFailureReason
	NextAttemptAt     *time.Time
	DeliveredAt       *time.Time
}

func MarkPaymentReceiptStatusNotificationSent(
	notificationID int64,
	deliveredAt time.Time,
) (PaymentReceiptStatusNotificationDeliveryResult, error) {
	if notificationID <= 0 {
		return PaymentReceiptStatusNotificationDeliveryResult{}, ErrPaymentReceiptStatusNotificationIDInvalid
	}
	if deliveredAt.IsZero() {
		return PaymentReceiptStatusNotificationDeliveryResult{}, ErrPaymentReceiptStatusNotificationDeliveredAtRequired
	}

	deliveredAtUTC := deliveredAt.UTC()
	return PaymentReceiptStatusNotificationDeliveryResult{
		NotificationID: notificationID,
		Status:         valueobjects.PaymentReceiptNotificationDeliveryStatusSent,
		DeliveredAt:    &deliveredAtUTC,
	}, nil
}

func ResolvePaymentReceiptStatusNotificationDeliveryFailure(
	notificationID int64,
	currentAttempts int32,
	maxAttempts int32,
	now time.Time,
	retryDelay time.Duration,
	failureReason valueobjects.PaymentReceiptNotificationDeliveryFailureReason,
) (PaymentReceiptStatusNotificationDeliveryResult, error) {
	if notificationID <= 0 {
		return PaymentReceiptStatusNotificationDeliveryResult{}, ErrPaymentReceiptStatusNotificationIDInvalid
	}
	if currentAttempts < 0 {
		return PaymentReceiptStatusNotificationDeliveryResult{}, ErrPaymentReceiptStatusNotificationCurrentAttemptsInvalid
	}
	if maxAttempts <= 0 {
		return PaymentReceiptStatusNotificationDeliveryResult{}, ErrPaymentReceiptStatusNotificationMaxAttemptsInvalid
	}
	if now.IsZero() {
		return PaymentReceiptStatusNotificationDeliveryResult{}, ErrPaymentReceiptStatusNotificationNowRequired
	}
	if failureReason.IsZero() {
		return PaymentReceiptStatusNotificationDeliveryResult{}, ErrPaymentReceiptStatusNotificationFailureReasonRequired
	}

	attempts := currentAttempts + 1
	if attempts >= maxAttempts {
		return PaymentReceiptStatusNotificationDeliveryResult{
			NotificationID:    notificationID,
			Status:            valueobjects.PaymentReceiptNotificationDeliveryStatusFailed,
			Attempts:          attempts,
			LastFailureReason: failureReason,
		}, nil
	}

	if retryDelay <= 0 {
		return PaymentReceiptStatusNotificationDeliveryResult{}, ErrPaymentReceiptStatusNotificationRetryDelayInvalid
	}

	nextAttemptAt := now.Add(retryDelay).UTC()
	return PaymentReceiptStatusNotificationDeliveryResult{
		NotificationID:    notificationID,
		Status:            valueobjects.PaymentReceiptNotificationDeliveryStatusPending,
		Attempts:          attempts,
		LastFailureReason: failureReason,
		NextAttemptAt:     &nextAttemptAt,
	}, nil
}
