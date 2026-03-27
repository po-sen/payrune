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
		return outport.DeriveIssuedPaymentAddressOutput{}, errors.New("bitcoin address deriver is not configured")
	}

	policy := input.Policy.Normalize()
	output, err := d.chainAddressDeriver.DeriveAddress(ctx, outport.DeriveChainAddressInput{
		Chain:                  policy.AddressPolicy.Chain,
		Network:                policy.AddressPolicy.Network,
		Scheme:                 policy.AddressPolicy.Scheme,
		AddressSourceRef:       policy.IssuanceConfig.AddressSourceRef,
		AddressReferencePrefix: policy.IssuanceConfig.AddressReferencePrefix,
		Index:                  input.Allocation.DerivationIndex,
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
