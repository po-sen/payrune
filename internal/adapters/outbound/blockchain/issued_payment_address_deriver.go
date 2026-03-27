package blockchain

import (
	"context"
	"errors"
	"strings"

	ethereumadapter "payrune/internal/adapters/outbound/ethereum"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

type IssuedPaymentAddressDeriver struct {
	chainAddressDeriver        outport.ChainAddressDeriver
	ethereumCreate2SaltDeriver *ethereumadapter.Create2SaltDeriver
}

var _ outport.IssuedPaymentAddressDeriver = (*IssuedPaymentAddressDeriver)(nil)

func NewIssuedPaymentAddressDeriver(
	chainAddressDeriver outport.ChainAddressDeriver,
	ethereumCreate2SaltDeriver *ethereumadapter.Create2SaltDeriver,
) *IssuedPaymentAddressDeriver {
	return &IssuedPaymentAddressDeriver{
		chainAddressDeriver:        chainAddressDeriver,
		ethereumCreate2SaltDeriver: ethereumCreate2SaltDeriver,
	}
}

func (d *IssuedPaymentAddressDeriver) SupportsChain(chain valueobjects.SupportedChain) bool {
	return d != nil && d.chainAddressDeriver != nil && d.chainAddressDeriver.SupportsChain(chain)
}

func (d *IssuedPaymentAddressDeriver) DeriveIssuedAddress(
	ctx context.Context,
	input outport.DeriveIssuedPaymentAddressInput,
) (outport.DeriveIssuedPaymentAddressOutput, error) {
	if d == nil || d.chainAddressDeriver == nil {
		return outport.DeriveIssuedPaymentAddressOutput{}, errors.New("chain address deriver is not configured")
	}

	policy := input.Policy.Normalize()
	relativeAddressReference, err := d.deriveRelativeAddressReference(ctx, input)
	if err != nil {
		return outport.DeriveIssuedPaymentAddressOutput{}, err
	}

	output, err := d.chainAddressDeriver.DeriveAddress(ctx, outport.DeriveChainAddressInput{
		Chain:                    policy.AddressPolicy.Chain,
		Network:                  policy.AddressPolicy.Network,
		Scheme:                   policy.AddressPolicy.Scheme,
		AddressSourceRef:         policy.IssuanceConfig.AddressSourceRef,
		AddressReferencePrefix:   policy.IssuanceConfig.AddressReferencePrefix,
		RelativeAddressReference: relativeAddressReference,
		Index:                    input.Allocation.DerivationIndex,
	})
	if err != nil {
		return outport.DeriveIssuedPaymentAddressOutput{}, err
	}

	addressReference := strings.TrimSpace(output.AddressReference)
	if addressReference == "" {
		addressReference = strings.TrimSpace(output.RelativeAddressReference)
	}

	return outport.DeriveIssuedPaymentAddressOutput{
		Address:          output.Address,
		AddressReference: addressReference,
	}, nil
}

func (d *IssuedPaymentAddressDeriver) deriveRelativeAddressReference(
	ctx context.Context,
	input outport.DeriveIssuedPaymentAddressInput,
) (string, error) {
	policy := input.Policy.Normalize()
	if policy.AddressPolicy.Chain != valueobjects.SupportedChainEthereum || policy.AddressPolicy.Scheme != "create2" {
		return "", nil
	}
	if d.ethereumCreate2SaltDeriver == nil {
		return "", errors.New("ethereum create2 salt deriver is not configured")
	}

	return d.ethereumCreate2SaltDeriver.DeriveAllocationSalt(ctx, ethereumadapter.DeriveCreate2AllocationSaltInput{
		Network:          policy.AddressPolicy.Network,
		AddressPolicyID:  policy.AddressPolicy.AddressPolicyID,
		PaymentAddressID: input.Allocation.PaymentAddressID,
		DerivationIndex:  input.Allocation.DerivationIndex,
	})
}
