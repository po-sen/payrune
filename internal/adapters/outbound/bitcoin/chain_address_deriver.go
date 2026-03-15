package bitcoin

import (
	"context"
	"errors"
	"fmt"

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
		return outport.DeriveChainAddressOutput{}, errors.New("bitcoin address deriver is not configured")
	}
	if input.Chain != valueobjects.SupportedChainBitcoin {
		return outport.DeriveChainAddressOutput{}, fmt.Errorf("bitcoin address deriver does not support chain: %s", input.Chain)
	}
	network, ok := valueobjects.ParseBitcoinNetwork(string(input.Network))
	if !ok {
		return outport.DeriveChainAddressOutput{}, fmt.Errorf("bitcoin network is invalid: %s", input.Network)
	}
	scheme, ok := valueobjects.ParseBitcoinAddressScheme(input.Scheme)
	if !ok {
		return outport.DeriveChainAddressOutput{}, fmt.Errorf("bitcoin address scheme is invalid: %s", input.Scheme)
	}

	address, err := g.deriver.DeriveAddress(
		network,
		scheme,
		input.AccountPublicKey,
		input.Index,
	)
	if err != nil {
		return outport.DeriveChainAddressOutput{}, err
	}

	absoluteDerivationPath, err := g.deriver.AbsoluteDerivationPath(
		input.AccountPublicKey,
		input.DerivationPathPrefix,
		input.Index,
	)
	if err != nil {
		return outport.DeriveChainAddressOutput{}, err
	}

	relativeDerivationPath, err := g.deriver.DerivationPath(input.AccountPublicKey, input.Index)
	if err != nil {
		return outport.DeriveChainAddressOutput{}, err
	}

	return outport.DeriveChainAddressOutput{
		Address:                address,
		RelativeDerivationPath: relativeDerivationPath,
		DerivationPath:         absoluteDerivationPath,
	}, nil
}
