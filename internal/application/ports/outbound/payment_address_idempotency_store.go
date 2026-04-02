package outbound

import (
	"context"
	"errors"

	"payrune/internal/domain/valueobjects"
)

var (
	ErrPaymentAddressIdempotencyStoreFailed                     = errors.New("payment address idempotency store failed")
	ErrPaymentAddressIdempotencyPersistedAddressPolicyIDInvalid = errors.New("persisted idempotency address policy id is invalid")
	ErrPaymentAddressIdempotencyKeyExists                       = errors.New("payment address idempotency key already exists")
	ErrPaymentAddressIdempotencyPersistedChainInvalid           = errors.New("persisted idempotency chain is invalid")
	ErrPaymentAddressIdempotencyChainRequired                   = errors.New("chain is required")
	ErrPaymentAddressIdempotencyKeyRequired                     = errors.New("idempotency key is required")
	ErrPaymentAddressIdempotencyAddressPolicyIDRequired         = errors.New("address policy id is required")
	ErrPaymentAddressIdempotencyExpectedAmountInvalid           = errors.New("expected amount minor must be greater than zero")
	ErrPaymentAddressIdempotencyPaymentAddressIDInvalid         = errors.New("payment address id must be greater than zero")
	ErrPaymentAddressIdempotencyClaimNotFound                   = errors.New("payment address idempotency claim was not found")
)

type PaymentAddressIdempotencyRecord struct {
	Chain               valueobjects.SupportedChain
	IdempotencyKey      string
	AddressPolicyID     valueobjects.AddressPolicyID
	ExpectedAmountMinor int64
	CustomerReference   string
	PaymentAddressID    int64
}

type FindPaymentAddressIdempotencyInput struct {
	Chain          valueobjects.SupportedChain
	IdempotencyKey string
}

type ClaimPaymentAddressIdempotencyInput struct {
	Chain               valueobjects.SupportedChain
	IdempotencyKey      string
	AddressPolicyID     valueobjects.AddressPolicyID
	ExpectedAmountMinor int64
	CustomerReference   string
}

type CompletePaymentAddressIdempotencyInput struct {
	Chain            valueobjects.SupportedChain
	IdempotencyKey   string
	PaymentAddressID int64
}

type ReleasePaymentAddressIdempotencyInput struct {
	Chain          valueobjects.SupportedChain
	IdempotencyKey string
}

type PaymentAddressIdempotencyStore interface {
	FindByKey(
		ctx context.Context,
		input FindPaymentAddressIdempotencyInput,
	) (PaymentAddressIdempotencyRecord, bool, error)
	Claim(
		ctx context.Context,
		input ClaimPaymentAddressIdempotencyInput,
	) (PaymentAddressIdempotencyRecord, error)
	Complete(ctx context.Context, input CompletePaymentAddressIdempotencyInput) error
	Release(ctx context.Context, input ReleasePaymentAddressIdempotencyInput) error
}
