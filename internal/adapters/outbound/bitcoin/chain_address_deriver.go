package bitcoin

import (
	"context"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

type addressDeriver interface {
	DeriveAddress(
		network valueobjects.BitcoinNetwork,
		scheme valueobjects.BitcoinAddressScheme,
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

func (g *ChainAddressDeriver) Chain() valueobjects.SupportedChain {
	return valueobjects.SupportedChainBitcoin
}

func (g *ChainAddressDeriver) SupportsChain(chain valueobjects.SupportedChain) bool {
	return chain == valueobjects.SupportedChainBitcoin
}

func (g *ChainAddressDeriver) DeriveAddress(
	_ context.Context,
	input outport.DeriveChainAddressInput,
) (outport.DeriveChainAddressOutput, error) {
	if g.deriver == nil {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDeriverNotConfigured
	}
	if input.Chain != valueobjects.SupportedChainBitcoin {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationInputInvalid
	}
	network, ok := valueobjects.ParseBitcoinNetwork(string(input.Network))
	if !ok {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationInputInvalid
	}
	scheme, ok := valueobjects.ParseBitcoinAddressScheme(input.Scheme)
	if !ok {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationInputInvalid
	}

	address, err := g.deriver.DeriveAddress(
		network,
		scheme,
		input.AddressSourceRef,
		input.Index,
	)
	if err != nil {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationFailed
	}

	absoluteDerivationPath, err := g.deriver.AbsoluteDerivationPath(
		input.AddressSourceRef,
		input.AddressReferencePrefix,
		input.Index,
	)
	if err != nil {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationFailed
	}

	relativeDerivationPath, err := g.deriver.DerivationPath(input.AddressSourceRef, input.Index)
	if err != nil {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationFailed
	}

	return outport.DeriveChainAddressOutput{
		Address:                  address,
		RelativeAddressReference: relativeDerivationPath,
		AddressReference:         absoluteDerivationPath,
	}, nil
}
