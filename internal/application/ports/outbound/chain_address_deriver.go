package outbound

import (
	"context"
	"errors"
)

var (
	ErrChainAddressDeriverNotConfigured   = errors.New("chain address deriver is not configured")
	ErrChainAddressDerivationInputInvalid = errors.New("chain address derivation input is invalid")
	ErrChainAddressDerivationFailed       = errors.New("chain address derivation failed")
)

type DeriveChainAddressInput struct {
	Chain               string
	Network             string
	Scheme              string
	AddressSpaceRef     string
	IssuanceRefPrefix   string
	RelativeIssuanceRef string
	SlotIndex           uint32
}

type DeriveChainAddressOutput struct {
	Address             string
	RelativeIssuanceRef string
	IssuanceRefKind     string
	IssuanceRef         string
}

type ChainAddressDeriver interface {
	SupportsChain(chain string) bool
	DeriveAddress(ctx context.Context, input DeriveChainAddressInput) (DeriveChainAddressOutput, error)
}
