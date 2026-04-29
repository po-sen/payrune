package ethereum

import (
	"context"
	"errors"

	outport "payrune/internal/application/ports/outbound"
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

func (d *IssuedPaymentAddressDeriver) Chain() string {
	return outport.SupportedChainEthereum
}

func (d *IssuedPaymentAddressDeriver) SupportsChain(chain string) bool {
	return chain == outport.SupportedChainEthereum
}

func (d *IssuedPaymentAddressDeriver) DeriveIssuedAddress(
	ctx context.Context,
	input outport.DeriveIssuedPaymentAddressInput,
) (outport.DeriveIssuedPaymentAddressOutput, error) {
	if d == nil || d.chainAddressDeriver == nil {
		return outport.DeriveIssuedPaymentAddressOutput{}, outport.ErrIssuedPaymentAddressDeriverNotConfigured
	}

	relativeIssuanceRef, err := d.deriveRelativeIssuanceRef(ctx, input)
	if err != nil {
		return outport.DeriveIssuedPaymentAddressOutput{}, err
	}

	output, err := d.chainAddressDeriver.DeriveAddress(ctx, outport.DeriveChainAddressInput{
		Chain:               input.Policy.Chain,
		Network:             input.Policy.Network,
		Scheme:              input.Policy.Scheme,
		AddressSpaceRef:     input.Policy.AddressSpaceRef,
		IssuanceRefPrefix:   input.Policy.IssuanceRefPrefix,
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
		input.Policy,
		output.Address,
		output.IssuanceRef,
	)
	if err != nil {
		return outport.DeriveIssuedPaymentAddressOutput{}, outport.ErrIssuedPaymentAddressDerivationFailed
	}

	return outport.DeriveIssuedPaymentAddressOutput{
		Address:         output.Address,
		IssuanceRefKind: output.IssuanceRefKind,
		IssuanceRef:     output.IssuanceRef,
		SweepMaterial:   sweepMaterialJSON,
	}, nil
}

func (d *IssuedPaymentAddressDeriver) deriveRelativeIssuanceRef(
	ctx context.Context,
	input outport.DeriveIssuedPaymentAddressInput,
) (string, error) {
	if input.Policy.Scheme != outport.AddressSchemeCreate2 {
		return "", nil
	}
	if d.create2SaltDeriver == nil {
		return "", outport.ErrIssuedPaymentAddressDeriverNotConfigured
	}

	relativeIssuanceRef, err := d.create2SaltDeriver.DeriveAllocationSalt(ctx, DeriveCreate2AllocationSaltInput{
		Network:          input.Policy.Network,
		AddressPolicyID:  input.Policy.AddressPolicyID,
		PaymentAddressID: input.Allocation.PaymentAddressID,
		SlotIndex:        input.Allocation.SlotIndex,
	})
	if err != nil {
		return "", outport.ErrIssuedPaymentAddressDerivationFailed
	}
	return relativeIssuanceRef, nil
}
