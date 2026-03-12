package outbound

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

type FindIssuedPaymentAddressAllocationByIDInput struct {
	PaymentAddressID int64
}

type PaymentAddressAllocationStore interface {
	FindIssuedByID(
		ctx context.Context,
		input FindIssuedPaymentAddressAllocationByIDInput,
	) (entities.PaymentAddressAllocation, bool, error)
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
