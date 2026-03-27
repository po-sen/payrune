package policies

import (
	"strings"
	"time"

	"payrune/internal/domain/valueobjects"
)

type PaymentReceiptStatusNotificationDeliveryResult struct {
	NotificationID int64
	Status         valueobjects.PaymentReceiptNotificationDeliveryStatus
	Attempts       int32
	LastError      string
	NextAttemptAt  *time.Time
	DeliveredAt    *time.Time
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
	lastError string,
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
	normalizedError := strings.TrimSpace(lastError)
	if normalizedError == "" {
		return PaymentReceiptStatusNotificationDeliveryResult{}, ErrPaymentReceiptStatusNotificationLastErrorRequired
	}

	attempts := currentAttempts + 1
	if attempts >= maxAttempts {
		return PaymentReceiptStatusNotificationDeliveryResult{
			NotificationID: notificationID,
			Status:         valueobjects.PaymentReceiptNotificationDeliveryStatusFailed,
			Attempts:       attempts,
			LastError:      normalizedError,
		}, nil
	}

	if retryDelay <= 0 {
		return PaymentReceiptStatusNotificationDeliveryResult{}, ErrPaymentReceiptStatusNotificationRetryDelayInvalid
	}

	nextAttemptAt := now.Add(retryDelay).UTC()
	return PaymentReceiptStatusNotificationDeliveryResult{
		NotificationID: notificationID,
		Status:         valueobjects.PaymentReceiptNotificationDeliveryStatusPending,
		Attempts:       attempts,
		LastError:      normalizedError,
		NextAttemptAt:  &nextAttemptAt,
	}, nil
}
