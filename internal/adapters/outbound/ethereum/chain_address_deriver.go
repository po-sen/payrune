package ethereum

import (
	"context"
	"errors"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

var errDeterministicAddressGenerationNotImplemented = errors.New(
	"ethereum deterministic address generation is not implemented",
)

type ChainAddressDeriver struct{}

func NewChainAddressDeriver() *ChainAddressDeriver {
	return &ChainAddressDeriver{}
}

func (d *ChainAddressDeriver) Chain() valueobjects.SupportedChain {
	return valueobjects.SupportedChainEthereum
}

func (d *ChainAddressDeriver) DeriveAddress(
	_ context.Context,
	_ outport.DeriveChainAddressInput,
) (outport.DeriveChainAddressOutput, error) {
	return outport.DeriveChainAddressOutput{}, errDeterministicAddressGenerationNotImplemented
}
