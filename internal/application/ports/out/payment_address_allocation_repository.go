package out

import (
	"context"
	"errors"

	"payrune/internal/domain/entities"
)

var ErrAddressIndexExhausted = errors.New("address index is exhausted")

type ReservePaymentAddressAllocationInput struct {
	Policy              entities.AddressPolicy
	ExpectedAmountMinor int64
	CustomerReference   string
}

type PaymentAddressAllocationRepository interface {
	ReopenFailedReservation(
		ctx context.Context,
		input ReservePaymentAddressAllocationInput,
	) (entities.PaymentAddressAllocation, bool, error)
	ReserveFresh(
		ctx context.Context,
		input ReservePaymentAddressAllocationInput,
	) (entities.PaymentAddressAllocation, error)
	Complete(ctx context.Context, allocation entities.PaymentAddressAllocation) error
	MarkDerivationFailed(ctx context.Context, allocation entities.PaymentAddressAllocation) error
}
