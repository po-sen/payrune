package outbound

import (
	"context"
	"errors"
	"time"

	"payrune/internal/domain/entities"
	"payrune/internal/domain/policies"
)

var ErrAddressIndexExhausted = errors.New("address index is exhausted")

var (
	ErrPaymentAddressAllocationStoreFailed                     = errors.New("payment address allocation store failed")
	ErrPaymentAddressAllocationNotReserved                     = errors.New("address allocation is not reserved")
	ErrPaymentAddressAllocationPersistedAddressPolicyIDInvalid = errors.New("persisted allocation address policy id is invalid")
	ErrPaymentAddressAllocationPersistedChainInvalid           = errors.New("persisted allocation chain is invalid")
	ErrPaymentAddressAllocationPersistedNetworkInvalid         = errors.New("persisted allocation network is invalid")
	ErrPaymentAddressAllocationIssuedAtRequired                = errors.New("issued at is required")
)

type ReservePaymentAddressAllocationInput struct {
	IssuancePolicy      policies.AddressIssuancePolicy
	ExpectedAmountMinor int64
	CustomerReference   string
}

type FindIssuedPaymentAddressAllocationByIDInput struct {
	PaymentAddressID int64
}

type CompletePaymentAddressAllocationInput struct {
	Allocation        entities.PaymentAddressAllocation
	SweepMaterialJSON string
	IssuedAt          time.Time
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
	Complete(ctx context.Context, input CompletePaymentAddressAllocationInput) error
	MarkDerivationFailed(ctx context.Context, allocation entities.PaymentAddressAllocation) error
}
