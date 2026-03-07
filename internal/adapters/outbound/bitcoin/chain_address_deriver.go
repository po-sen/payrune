package bitcoin

import (
	"context"
	"errors"
	"fmt"

	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/value_objects"
)

type addressDeriver interface {
	DeriveAddress(
		network value_objects.BitcoinNetwork,
		scheme value_objects.BitcoinAddressScheme,
		xpub string,
		index uint32,
	) (string, error)
	DerivationPath(xpub string, index uint32) (string, error)
}

type ChainAddressDeriver struct {
	deriver addressDeriver
}

func NewChainAddressDeriver(deriver addressDeriver) *ChainAddressDeriver {
	return &ChainAddressDeriver{deriver: deriver}
}

func (g *ChainAddressDeriver) Chain() value_objects.SupportedChain {
	return value_objects.SupportedChainBitcoin
}

func (g *ChainAddressDeriver) SupportsChain(chain value_objects.SupportedChain) bool {
	return chain == value_objects.SupportedChainBitcoin
}

func (g *ChainAddressDeriver) DeriveAddress(
	_ context.Context,
	input outport.DeriveChainAddressInput,
) (outport.DeriveChainAddressOutput, error) {
	if g.deriver == nil {
		return outport.DeriveChainAddressOutput{}, errors.New("bitcoin address deriver is not configured")
	}
	if input.Chain != value_objects.SupportedChainBitcoin {
		return outport.DeriveChainAddressOutput{}, fmt.Errorf("bitcoin address deriver does not support chain: %s", input.Chain)
	}
	network, ok := value_objects.ParseBitcoinNetwork(string(input.Network))
	if !ok {
		return outport.DeriveChainAddressOutput{}, fmt.Errorf("bitcoin network is invalid: %s", input.Network)
	}
	scheme, ok := value_objects.ParseBitcoinAddressScheme(input.Scheme)
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

	relativeDerivationPath, err := g.deriver.DerivationPath(input.AccountPublicKey, input.Index)
	if err != nil {
		return outport.DeriveChainAddressOutput{}, err
	}

	return outport.DeriveChainAddressOutput{
		Address:                address,
		RelativeDerivationPath: relativeDerivationPath,
	}, nil
}
