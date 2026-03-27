package outbound

import (
	"context"
	"errors"
	"time"

	applicationoutbox "payrune/internal/application/outbox"
	"payrune/internal/domain/events"
	"payrune/internal/domain/policies"
)

var (
	ErrPaymentReceiptStatusNotificationClaimNowRequired      = errors.New("claim now is required")
	ErrPaymentReceiptStatusNotificationClaimUntilRequired    = errors.New("claim until is required")
	ErrPaymentReceiptStatusNotificationClaimLimitInvalid     = errors.New("claim limit must be greater than zero")
	ErrPaymentReceiptStatusNotificationDeliveredAtRequired   = errors.New("delivered at is required")
	ErrPaymentReceiptStatusNotificationNextAttemptRequired   = errors.New("next attempt at is required")
	ErrPaymentReceiptStatusNotificationDeliveryStatusInvalid = errors.New("delivery result status is invalid")
	ErrPaymentReceiptStatusNotificationNotFound              = errors.New("payment receipt status notification is not found")
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
