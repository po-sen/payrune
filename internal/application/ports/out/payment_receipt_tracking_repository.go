package out

import (
	"context"
	"time"

	"payrune/internal/domain/entities"
)

type ClaimPaymentReceiptTrackingsInput struct {
	Now        time.Time
	Limit      int
	ClaimUntil time.Time
	Chain      string
	Network    string
}

type PaymentReceiptTrackingRepository interface {
	RegisterIssuedAllocation(
		ctx context.Context,
		paymentAddressID int64,
		defaultRequiredConfirmations int32,
		expiresAt time.Time,
	) (bool, error)
	ClaimDue(
		ctx context.Context,
		input ClaimPaymentReceiptTrackingsInput,
	) ([]entities.PaymentReceiptTracking, error)
	SaveObservation(
		ctx context.Context,
		tracking entities.PaymentReceiptTracking,
		now time.Time,
		nextPollAt time.Time,
	) error
	SavePollingError(
		ctx context.Context,
		paymentAddressID int64,
		errorReason string,
		now time.Time,
		nextPollAt time.Time,
	) error
}
