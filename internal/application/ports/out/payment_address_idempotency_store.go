package out

import (
	"context"
	"errors"

	"payrune/internal/domain/value_objects"
)

var ErrPaymentAddressIdempotencyKeyExists = errors.New("payment address idempotency key already exists")

type PaymentAddressIdempotencyRecord struct {
	Chain               value_objects.SupportedChain
	IdempotencyKey      string
	AddressPolicyID     string
	ExpectedAmountMinor int64
	CustomerReference   string
	PaymentAddressID    int64
}

type FindPaymentAddressIdempotencyInput struct {
	Chain          value_objects.SupportedChain
	IdempotencyKey string
}

type ClaimPaymentAddressIdempotencyInput struct {
	Chain               value_objects.SupportedChain
	IdempotencyKey      string
	AddressPolicyID     string
	ExpectedAmountMinor int64
	CustomerReference   string
}

type CompletePaymentAddressIdempotencyInput struct {
	Chain            value_objects.SupportedChain
	IdempotencyKey   string
	PaymentAddressID int64
}

type ReleasePaymentAddressIdempotencyInput struct {
	Chain          value_objects.SupportedChain
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
