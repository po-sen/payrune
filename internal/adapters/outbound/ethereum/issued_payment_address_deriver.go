package ethereum

import (
	"context"
	"errors"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

type IssuedPaymentAddressDeriver struct {
	chainAddressDeriver *ChainAddressDeriver
	create2SaltDeriver  *Create2SaltDeriver
}

var _ outport.IssuedPaymentAddressDeriver = (*IssuedPaymentAddressDeriver)(nil)

func NewIssuedPaymentAddressDeriver(
	chainAddressDeriver *ChainAddressDeriver,
	create2SaltDeriver *Create2SaltDeriver,
) *IssuedPaymentAddressDeriver {
	return &IssuedPaymentAddressDeriver{
		chainAddressDeriver: chainAddressDeriver,
		create2SaltDeriver:  create2SaltDeriver,
	}
}

func (d *IssuedPaymentAddressDeriver) Chain() valueobjects.SupportedChain {
	return valueobjects.SupportedChainEthereum
}

func (d *IssuedPaymentAddressDeriver) SupportsChain(chain valueobjects.SupportedChain) bool {
	return chain == valueobjects.SupportedChainEthereum
}

func (d *IssuedPaymentAddressDeriver) DeriveIssuedAddress(
	ctx context.Context,
	input outport.DeriveIssuedPaymentAddressInput,
) (outport.DeriveIssuedPaymentAddressOutput, error) {
	if d == nil || d.chainAddressDeriver == nil {
		return outport.DeriveIssuedPaymentAddressOutput{}, outport.ErrIssuedPaymentAddressDeriverNotConfigured
	}

	policy := input.Policy.Normalize()
	relativeIssuanceRef, err := d.deriveRelativeIssuanceRef(ctx, input)
	if err != nil {
		return outport.DeriveIssuedPaymentAddressOutput{}, err
	}

	output, err := d.chainAddressDeriver.DeriveAddress(ctx, outport.DeriveChainAddressInput{
		Chain:               policy.Chain,
		Network:             policy.Network,
		Scheme:              policy.Scheme,
		AddressSpaceRef:     policy.IssuanceConfig.AddressSpaceRef,
		IssuanceRefPrefix:   policy.IssuanceConfig.IssuanceRefPrefix,
		RelativeIssuanceRef: relativeIssuanceRef,
		SlotIndex:           input.Allocation.SlotIndex,
	})
	if err != nil {
		switch {
		case errors.Is(err, outport.ErrChainAddressDeriverNotConfigured):
			return outport.DeriveIssuedPaymentAddressOutput{}, outport.ErrIssuedPaymentAddressDeriverNotConfigured
		case errors.Is(err, outport.ErrChainAddressDerivationInputInvalid):
			return outport.DeriveIssuedPaymentAddressOutput{}, outport.ErrIssuedPaymentAddressDerivationInputInvalid
		default:
			return outport.DeriveIssuedPaymentAddressOutput{}, outport.ErrIssuedPaymentAddressDerivationFailed
		}
	}

	sweepMaterialJSON, err := buildSweepMaterialJSON(
		policy,
		output.Address,
		output.IssuanceRef,
	)
	if err != nil {
		return outport.DeriveIssuedPaymentAddressOutput{}, outport.ErrIssuedPaymentAddressDerivationFailed
	}

	return outport.DeriveIssuedPaymentAddressOutput{
		Address:           output.Address,
		IssuanceRefKind:   output.IssuanceRefKind,
		IssuanceRef:       output.IssuanceRef,
		SweepMaterialJSON: sweepMaterialJSON,
	}, nil
}

func (d *IssuedPaymentAddressDeriver) deriveRelativeIssuanceRef(
	ctx context.Context,
	input outport.DeriveIssuedPaymentAddressInput,
) (string, error) {
	policy := input.Policy.Normalize()
	if !policy.Scheme.IsCreate2() {
		return "", nil
	}
	if d.create2SaltDeriver == nil {
		return "", outport.ErrIssuedPaymentAddressDeriverNotConfigured
	}

	relativeIssuanceRef, err := d.create2SaltDeriver.DeriveAllocationSalt(ctx, DeriveCreate2AllocationSaltInput{
		Network:          policy.Network,
		AddressPolicyID:  policy.AddressPolicyID,
		PaymentAddressID: input.Allocation.PaymentAddressID,
		SlotIndex:        input.Allocation.SlotIndex,
	})
	if err != nil {
		return "", outport.ErrIssuedPaymentAddressDerivationFailed
	}
	return relativeIssuanceRef, nil
}
