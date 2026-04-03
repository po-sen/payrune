package outbound

import (
	"context"
	"errors"

	"payrune/internal/domain/entities"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
)

var (
	ErrIssuedPaymentAddressDeriverNotConfigured   = errors.New("issued payment address deriver is not configured")
	ErrIssuedPaymentAddressDerivationInputInvalid = errors.New("issued payment address derivation input is invalid")
	ErrIssuedPaymentAddressDerivationFailed       = errors.New("issued payment address derivation failed")
)

type DeriveIssuedPaymentAddressInput struct {
	Policy     policies.AddressIssuancePolicy
	Allocation entities.PaymentAddressAllocation
}

type DeriveIssuedPaymentAddressOutput struct {
	Address           string
	IssuanceRefKind   valueobjects.IssuanceRefKind
	IssuanceRef       string
	SweepMaterialJSON string
}

type IssuedPaymentAddressDeriver interface {
	SupportsChain(chain valueobjects.SupportedChain) bool
	DeriveIssuedAddress(
		ctx context.Context,
		input DeriveIssuedPaymentAddressInput,
	) (DeriveIssuedPaymentAddressOutput, error)
}
