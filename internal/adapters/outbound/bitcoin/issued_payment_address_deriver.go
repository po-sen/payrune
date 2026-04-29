package bitcoin

import (
	"context"
	"errors"

	outport "payrune/internal/application/ports/outbound"
)

type IssuedPaymentAddressDeriver struct {
	chainAddressDeriver *ChainAddressDeriver
}

var _ outport.IssuedPaymentAddressDeriver = (*IssuedPaymentAddressDeriver)(nil)

func NewIssuedPaymentAddressDeriver(chainAddressDeriver *ChainAddressDeriver) *IssuedPaymentAddressDeriver {
	return &IssuedPaymentAddressDeriver{chainAddressDeriver: chainAddressDeriver}
}

func (d *IssuedPaymentAddressDeriver) Chain() string {
	return outport.SupportedChainBitcoin
}

func (d *IssuedPaymentAddressDeriver) SupportsChain(chain string) bool {
	return chain == outport.SupportedChainBitcoin
}

func (d *IssuedPaymentAddressDeriver) DeriveIssuedAddress(
	ctx context.Context,
	input outport.DeriveIssuedPaymentAddressInput,
) (outport.DeriveIssuedPaymentAddressOutput, error) {
	if d == nil || d.chainAddressDeriver == nil {
		return outport.DeriveIssuedPaymentAddressOutput{}, outport.ErrIssuedPaymentAddressDeriverNotConfigured
	}

	output, err := d.chainAddressDeriver.DeriveAddress(ctx, outport.DeriveChainAddressInput{
		Chain:             input.Policy.Chain,
		Network:           input.Policy.Network,
		Scheme:            input.Policy.Scheme,
		AddressSpaceRef:   input.Policy.AddressSpaceRef,
		IssuanceRefPrefix: input.Policy.IssuanceRefPrefix,
		SlotIndex:         input.Allocation.SlotIndex,
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
		input.Policy.Chain,
		input.Policy.Network,
		output.Address,
		output.IssuanceRef,
		input.Policy.AddressSpaceRef,
		input.Policy.Scheme,
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
