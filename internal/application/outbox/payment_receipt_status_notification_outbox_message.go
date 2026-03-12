package outbox

import (
	"time"

	"payrune/internal/domain/valueobjects"
)

// PaymentReceiptStatusNotificationOutboxMessage is a claimed outbox row used by
// the webhook dispatch use case.
type PaymentReceiptStatusNotificationOutboxMessage struct {
	NotificationID        int64
	PaymentAddressID      int64
	CustomerReference     string
	PreviousStatus        valueobjects.PaymentReceiptStatus
	CurrentStatus         valueobjects.PaymentReceiptStatus
	ObservedTotalMinor    int64
	ConfirmedTotalMinor   int64
	UnconfirmedTotalMinor int64
	StatusChangedAt       time.Time
	DeliveryStatus        valueobjects.PaymentReceiptNotificationDeliveryStatus
	DeliveryAttempts      int32
	NextAttemptAt         time.Time
	LastError             string
	DeliveredAt           *time.Time
}
