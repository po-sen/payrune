package outbound

import (
	"context"
	"errors"
)

var (
	ErrIssuedPaymentAddressDeriverNotConfigured   = errors.New("issued payment address deriver is not configured")
	ErrIssuedPaymentAddressDerivationInputInvalid = errors.New("issued payment address derivation input is invalid")
	ErrIssuedPaymentAddressDerivationFailed       = errors.New("issued payment address derivation failed")
)

type DeriveIssuedPaymentAddressInput struct {
	Policy     AddressIssuancePolicyRecord
	Allocation PaymentAddressAllocationRecord
}

type DeriveIssuedPaymentAddressOutput struct {
	Address         string
	IssuanceRefKind string
	IssuanceRef     string
	SweepMaterial   string
}

type IssuedPaymentAddressDeriver interface {
	SupportsChain(chain string) bool
	DeriveIssuedAddress(
		ctx context.Context,
		input DeriveIssuedPaymentAddressInput,
	) (DeriveIssuedPaymentAddressOutput, error)
}
