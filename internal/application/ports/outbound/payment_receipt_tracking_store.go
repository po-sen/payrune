package outbound

import (
	"context"
	"time"

	"payrune/internal/domain/entities"
	"payrune/internal/domain/valueobjects"
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
