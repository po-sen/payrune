package bitcoin

import (
	"context"

	outport "payrune/internal/application/ports/outbound"
)

type addressDeriver interface {
	DeriveAddress(
		network network,
		scheme addressScheme,
		xpub string,
		index uint32,
	) (string, error)
	DerivationPath(xpub string, index uint32) (string, error)
	AbsoluteDerivationPath(xpub string, derivationPathPrefix string, index uint32) (string, error)
}

type ChainAddressDeriver struct {
	deriver addressDeriver
}

func NewChainAddressDeriver(deriver addressDeriver) *ChainAddressDeriver {
	return &ChainAddressDeriver{deriver: deriver}
}

func (g *ChainAddressDeriver) Chain() string {
	return outport.SupportedChainBitcoin
}

func (g *ChainAddressDeriver) SupportsChain(chain string) bool {
	return chain == outport.SupportedChainBitcoin
}

func (g *ChainAddressDeriver) DeriveAddress(
	_ context.Context,
	input outport.DeriveChainAddressInput,
) (outport.DeriveChainAddressOutput, error) {
	if g.deriver == nil {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDeriverNotConfigured
	}
	if input.Chain != outport.SupportedChainBitcoin {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationInputInvalid
	}
	network, ok := parseNetwork(input.Network)
	if !ok {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationInputInvalid
	}
	scheme, ok := parseAddressScheme(input.Scheme)
	if !ok {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationInputInvalid
	}

	address, err := g.deriver.DeriveAddress(
		network,
		scheme,
		input.AddressSpaceRef,
		input.SlotIndex,
	)
	if err != nil {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationFailed
	}

	absoluteDerivationPath, err := g.deriver.AbsoluteDerivationPath(
		input.AddressSpaceRef,
		input.IssuanceRefPrefix,
		input.SlotIndex,
	)
	if err != nil {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationFailed
	}

	relativeDerivationPath, err := g.deriver.DerivationPath(input.AddressSpaceRef, input.SlotIndex)
	if err != nil {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationFailed
	}

	return outport.DeriveChainAddressOutput{
		Address:             address,
		RelativeIssuanceRef: relativeDerivationPath,
		IssuanceRefKind:     outport.IssuanceRefKindHDPathAbsolute,
		IssuanceRef:         absoluteDerivationPath,
	}, nil
}
