package out

import (
	"context"
	"errors"
	"time"

	"payrune/internal/domain/entities"
)

var ErrAddressIndexExhausted = errors.New("address index is exhausted")

type ReservePaymentAddressAllocationInput struct {
	IssuancePolicy      entities.AddressIssuancePolicy
	ExpectedAmountMinor int64
	CustomerReference   string
}

type PaymentAddressAllocationStore interface {
	ReopenFailedReservation(
		ctx context.Context,
		input ReservePaymentAddressAllocationInput,
	) (entities.PaymentAddressAllocation, bool, error)
	ReserveFresh(
		ctx context.Context,
		input ReservePaymentAddressAllocationInput,
	) (entities.PaymentAddressAllocation, error)
	Complete(ctx context.Context, allocation entities.PaymentAddressAllocation, issuedAt time.Time) error
	MarkDerivationFailed(ctx context.Context, allocation entities.PaymentAddressAllocation) error
}
