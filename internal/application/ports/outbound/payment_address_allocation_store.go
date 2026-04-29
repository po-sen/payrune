package outbound

import (
	"context"
	"errors"
	"time"
)

var ErrAddressIndexExhausted = errors.New("address index is exhausted")

var (
	ErrPaymentAddressAllocationStoreFailed                     = errors.New("payment address allocation store failed")
	ErrPaymentAddressAllocationNotReserved                     = errors.New("address allocation is not reserved")
	ErrPaymentAddressAllocationPersistedAddressPolicyIDInvalid = errors.New("persisted allocation address policy id is invalid")
	ErrPaymentAddressAllocationPersistedChainInvalid           = errors.New("persisted allocation chain is invalid")
	ErrPaymentAddressAllocationPersistedNetworkInvalid         = errors.New("persisted allocation network is invalid")
	ErrPaymentAddressAllocationPersistedAssetReferenceInvalid  = errors.New("persisted allocation asset reference is invalid")
	ErrPaymentAddressAllocationIssuedAtRequired                = errors.New("issued at is required")
)

type PaymentAddressAllocationRecord struct {
	PaymentAddressID        int64
	AddressPolicyID         string
	SlotIndex               uint32
	ExpectedAmountMinor     int64
	CustomerReference       string
	Status                  string
	Chain                   string
	Network                 string
	Scheme                  string
	AssetReference          string
	Address                 string
	DerivationFailureReason string
}

type ReservePaymentAddressAllocationInput struct {
	IssuancePolicy      AddressIssuancePolicyRecord
	ExpectedAmountMinor int64
	CustomerReference   string
}

type FindIssuedPaymentAddressAllocationByIDInput struct {
	PaymentAddressID int64
}

type CompletePaymentAddressAllocationInput struct {
	Allocation    PaymentAddressAllocationRecord
	SweepMaterial string
	IssuedAt      time.Time
}

type PaymentAddressAllocationStore interface {
	FindIssuedByID(
		ctx context.Context,
		input FindIssuedPaymentAddressAllocationByIDInput,
	) (PaymentAddressAllocationRecord, bool, error)
	ReopenFailedReservation(
		ctx context.Context,
		input ReservePaymentAddressAllocationInput,
	) (PaymentAddressAllocationRecord, bool, error)
	ReserveFresh(
		ctx context.Context,
		input ReservePaymentAddressAllocationInput,
	) (PaymentAddressAllocationRecord, error)
	Complete(ctx context.Context, input CompletePaymentAddressAllocationInput) error
	MarkDerivationFailed(ctx context.Context, allocation PaymentAddressAllocationRecord) error
}
