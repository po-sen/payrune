package bitcoin

import (
	"context"
	"errors"
	"strings"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

type IssuedPaymentAddressDeriver struct {
	chainAddressDeriver *ChainAddressDeriver
}

var _ outport.IssuedPaymentAddressDeriver = (*IssuedPaymentAddressDeriver)(nil)

func NewIssuedPaymentAddressDeriver(chainAddressDeriver *ChainAddressDeriver) *IssuedPaymentAddressDeriver {
	return &IssuedPaymentAddressDeriver{chainAddressDeriver: chainAddressDeriver}
}

func (d *IssuedPaymentAddressDeriver) Chain() valueobjects.SupportedChain {
	return valueobjects.SupportedChainBitcoin
}

func (d *IssuedPaymentAddressDeriver) SupportsChain(chain valueobjects.SupportedChain) bool {
	return chain == valueobjects.SupportedChainBitcoin
}

func (d *IssuedPaymentAddressDeriver) DeriveIssuedAddress(
	ctx context.Context,
	input outport.DeriveIssuedPaymentAddressInput,
) (outport.DeriveIssuedPaymentAddressOutput, error) {
	if d == nil || d.chainAddressDeriver == nil {
		return outport.DeriveIssuedPaymentAddressOutput{}, outport.ErrIssuedPaymentAddressDeriverNotConfigured
	}

	policy := input.Policy.Normalize()
	output, err := d.chainAddressDeriver.DeriveAddress(ctx, outport.DeriveChainAddressInput{
		Chain:             policy.Chain,
		Network:           policy.Network,
		Scheme:            policy.Scheme,
		AddressSpaceRef:   policy.IssuanceConfig.AddressSpaceRef,
		IssuanceRefPrefix: policy.IssuanceConfig.IssuanceRefPrefix,
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
		string(policy.Chain),
		string(policy.Network),
		output.Address,
		output.IssuanceRef,
		strings.TrimSpace(policy.IssuanceConfig.AddressSpaceRef),
		string(policy.Scheme.Normalize()),
	)
	if err != nil {
		return outport.DeriveIssuedPaymentAddressOutput{}, outport.ErrIssuedPaymentAddressDerivationFailed
	}

	return outport.DeriveIssuedPaymentAddressOutput{
		Address:           output.Address,
		SweepMaterialJSON: sweepMaterialJSON,
		IssuanceRefKind:   output.IssuanceRefKind,
		IssuanceRef:       output.IssuanceRef,
	}, nil
}
