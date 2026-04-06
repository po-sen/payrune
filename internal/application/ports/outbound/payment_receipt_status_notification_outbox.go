package outbound

import (
	"context"
	"errors"
	"time"

	applicationoutbox "payrune/internal/application/outbox"
	"payrune/internal/domain/events"
)

var (
	ErrPaymentReceiptStatusNotificationOutboxFailed                    = errors.New("payment receipt status notification outbox failed")
	ErrPaymentReceiptStatusNotificationClaimNowRequired                = errors.New("claim now is required")
	ErrPaymentReceiptStatusNotificationClaimUntilRequired              = errors.New("claim until is required")
	ErrPaymentReceiptStatusNotificationClaimLimitInvalid               = errors.New("claim limit must be greater than zero")
	ErrPaymentReceiptStatusNotificationDeliveredAtRequired             = errors.New("delivered at is required")
	ErrPaymentReceiptStatusNotificationNextAttemptRequired             = errors.New("next attempt at is required")
	ErrPaymentReceiptStatusNotificationDeliveryStatusInvalid           = errors.New("delivery result status is invalid")
	ErrPaymentReceiptStatusNotificationNotFound                        = errors.New("payment receipt status notification is not found")
	ErrPaymentReceiptStatusNotificationPersistedAddressPolicyIDInvalid = errors.New("persisted receipt notification address policy id is invalid")
	ErrPaymentReceiptStatusNotificationPersistedPreviousStatusInvalid  = errors.New("persisted previous receipt status is invalid")
	ErrPaymentReceiptStatusNotificationPersistedCurrentStatusInvalid   = errors.New("persisted current receipt status is invalid")
	ErrPaymentReceiptStatusNotificationPersistedDeliveryStatusInvalid  = errors.New("persisted receipt notification delivery status is invalid")
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
		result applicationoutbox.PaymentReceiptStatusNotificationDeliveryResult,
	) error
}

type ClaimPaymentReceiptStatusNotificationsInput struct {
	Now        time.Time
	Limit      int
	ClaimUntil time.Time
}
