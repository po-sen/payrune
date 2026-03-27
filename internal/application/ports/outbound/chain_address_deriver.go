package outbound

import (
	"context"
	"errors"
	"payrune/internal/domain/valueobjects"
)

var (
	ErrChainAddressDeriverNotConfigured   = errors.New("chain address deriver is not configured")
	ErrChainAddressDerivationInputInvalid = errors.New("chain address derivation input is invalid")
	ErrChainAddressDerivationFailed       = errors.New("chain address derivation failed")
)

type DeriveChainAddressInput struct {
	Chain                    valueobjects.SupportedChain
	Network                  valueobjects.NetworkID
	Scheme                   string
	AddressSourceRef         string
	AddressReferencePrefix   string
	RelativeAddressReference string
	Index                    uint32
}

type DeriveChainAddressOutput struct {
	Address                  string
	RelativeAddressReference string
	AddressReference         string
}

type ChainAddressDeriver interface {
	SupportsChain(chain valueobjects.SupportedChain) bool
	DeriveAddress(ctx context.Context, input DeriveChainAddressInput) (DeriveChainAddressOutput, error)
}
