package outbox

import (
	"time"

	"payrune/internal/domain/value_objects"
)

// PaymentReceiptStatusNotificationOutboxMessage is a claimed outbox row used by
// the webhook dispatch use case.
type PaymentReceiptStatusNotificationOutboxMessage struct {
	NotificationID        int64
	PaymentAddressID      int64
	CustomerReference     string
	PreviousStatus        value_objects.PaymentReceiptStatus
	CurrentStatus         value_objects.PaymentReceiptStatus
	ObservedTotalMinor    int64
	ConfirmedTotalMinor   int64
	UnconfirmedTotalMinor int64
	ConflictTotalMinor    int64
	StatusChangedAt       time.Time
	DeliveryStatus        value_objects.PaymentReceiptNotificationDeliveryStatus
	DeliveryAttempts      int32
	NextAttemptAt         time.Time
	LastError             string
	DeliveredAt           *time.Time
}
