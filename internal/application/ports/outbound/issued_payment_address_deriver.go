package outbound

import (
	"context"

	"payrune/internal/domain/entities"
	"payrune/internal/domain/valueobjects"
)

type DeriveIssuedPaymentAddressInput struct {
	Policy     entities.AddressIssuancePolicy
	Allocation entities.PaymentAddressAllocation
}

type DeriveIssuedPaymentAddressOutput struct {
	Address          string
	AddressReference string
}

type IssuedPaymentAddressDeriver interface {
	SupportsChain(chain valueobjects.SupportedChain) bool
	DeriveIssuedAddress(
		ctx context.Context,
		input DeriveIssuedPaymentAddressInput,
	) (DeriveIssuedPaymentAddressOutput, error)
}
