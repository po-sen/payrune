package outbox

import (
	"errors"
	"time"
)

var (
	ErrPaymentReceiptStatusNotificationIDInvalid              = errors.New("notification id must be greater than zero")
	ErrPaymentReceiptStatusNotificationDeliveredAtRequired    = errors.New("delivered at is required")
	ErrPaymentReceiptStatusNotificationCurrentAttemptsInvalid = errors.New("current attempts must be greater than or equal to zero")
	ErrPaymentReceiptStatusNotificationMaxAttemptsInvalid     = errors.New("max attempts must be greater than zero")
	ErrPaymentReceiptStatusNotificationNowRequired            = errors.New("now is required")
	ErrPaymentReceiptStatusNotificationFailureReasonRequired  = errors.New("notification failure reason is required")
	ErrPaymentReceiptStatusNotificationRetryDelayInvalid      = errors.New("retry delay must be greater than zero")
)

type PaymentReceiptStatusNotificationDeliveryResult struct {
	NotificationID    int64
	Status            PaymentReceiptNotificationDeliveryStatus
	Attempts          int32
	LastFailureReason PaymentReceiptNotificationDeliveryFailureReason
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
		Status:         PaymentReceiptNotificationDeliveryStatusSent,
		DeliveredAt:    &deliveredAtUTC,
	}, nil
}

func ResolvePaymentReceiptStatusNotificationDeliveryFailure(
	notificationID int64,
	currentAttempts int32,
	maxAttempts int32,
	now time.Time,
	retryDelay time.Duration,
	failureReason PaymentReceiptNotificationDeliveryFailureReason,
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
			Status:            PaymentReceiptNotificationDeliveryStatusFailed,
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
		Status:            PaymentReceiptNotificationDeliveryStatusPending,
		Attempts:          attempts,
		LastFailureReason: failureReason,
		NextAttemptAt:     &nextAttemptAt,
	}, nil
}
