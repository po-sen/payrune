package policies

import (
	"errors"
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
		return PaymentReceiptStatusNotificationDeliveryResult{}, errors.New("notification id must be greater than zero")
	}
	if deliveredAt.IsZero() {
		return PaymentReceiptStatusNotificationDeliveryResult{}, errors.New("delivered at is required")
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
		return PaymentReceiptStatusNotificationDeliveryResult{}, errors.New("notification id must be greater than zero")
	}
	if currentAttempts < 0 {
		return PaymentReceiptStatusNotificationDeliveryResult{}, errors.New("current attempts must be greater than or equal to zero")
	}
	if maxAttempts <= 0 {
		return PaymentReceiptStatusNotificationDeliveryResult{}, errors.New("max attempts must be greater than zero")
	}
	if now.IsZero() {
		return PaymentReceiptStatusNotificationDeliveryResult{}, errors.New("now is required")
	}
	normalizedError := strings.TrimSpace(lastError)
	if normalizedError == "" {
		return PaymentReceiptStatusNotificationDeliveryResult{}, errors.New("last error is required")
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
		return PaymentReceiptStatusNotificationDeliveryResult{}, errors.New("retry delay must be greater than zero")
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
