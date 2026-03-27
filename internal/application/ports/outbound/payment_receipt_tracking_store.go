package outbound

import (
	"context"
	"errors"
	"time"

	"payrune/internal/domain/entities"
	"payrune/internal/domain/valueobjects"
)

var (
	ErrPaymentReceiptTrackingNextPollAtRequired    = errors.New("next poll at is required")
	ErrPaymentReceiptTrackingAlreadyExists         = errors.New("payment receipt tracking already exists")
	ErrPaymentReceiptTrackingClaimNowRequired      = errors.New("claim now is required")
	ErrPaymentReceiptTrackingClaimUntilRequired    = errors.New("claim until is required")
	ErrPaymentReceiptTrackingClaimLimitInvalid     = errors.New("claim limit must be greater than zero")
	ErrPaymentReceiptTrackingClaimStatusesRequired = errors.New("claim statuses are required")
	ErrPaymentReceiptTrackingClaimStatusRequired   = errors.New("claim status is required")
	ErrPaymentReceiptTrackingPolledAtRequired      = errors.New("polled at is required")
	ErrPaymentReceiptTrackingNotFound              = errors.New("payment receipt tracking is not found")
)

type ClaimPaymentReceiptTrackingsInput struct {
	Now        time.Time
	Limit      int
	ClaimUntil time.Time
	Chain      string
	Network    string
	Statuses   []valueobjects.PaymentReceiptStatus
}

type PaymentReceiptTrackingStore interface {
	Create(
		ctx context.Context,
		tracking entities.PaymentReceiptTracking,
		nextPollAt time.Time,
	) error
	ClaimDue(
		ctx context.Context,
		input ClaimPaymentReceiptTrackingsInput,
	) ([]entities.PaymentReceiptTracking, error)
	Save(
		ctx context.Context,
		tracking entities.PaymentReceiptTracking,
		polledAt time.Time,
		nextPollAt time.Time,
	) error
}
