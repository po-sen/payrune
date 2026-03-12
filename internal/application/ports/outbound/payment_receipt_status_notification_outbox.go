package outbound

import (
	"context"
	"time"

	applicationoutbox "payrune/internal/application/outbox"
	"payrune/internal/domain/events"
	"payrune/internal/domain/policies"
)

type PaymentReceiptStatusNotificationOutbox interface {
	EnqueueStatusChanged(
		ctx context.Context,
		event events.PaymentReceiptStatusChanged,
	) error
	ClaimPending(
		ctx context.Context,
		input ClaimPaymentReceiptStatusNotificationsInput,
	) ([]applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage, error)
	SaveDeliveryResult(
		ctx context.Context,
		result policies.PaymentReceiptStatusNotificationDeliveryResult,
	) error
}

type ClaimPaymentReceiptStatusNotificationsInput struct {
	Now        time.Time
	Limit      int
	ClaimUntil time.Time
}
