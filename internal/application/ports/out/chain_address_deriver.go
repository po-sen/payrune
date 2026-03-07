package out

import (
	"context"
	"payrune/internal/domain/value_objects"
)

type DeriveChainAddressInput struct {
	Chain            value_objects.SupportedChain
	Network          value_objects.NetworkID
	Scheme           string
	AccountPublicKey string
	Index            uint32
}

type DeriveChainAddressOutput struct {
	Address                string
	RelativeDerivationPath string
}

type ChainAddressDeriver interface {
	SupportsChain(chain value_objects.SupportedChain) bool
	DeriveAddress(ctx context.Context, input DeriveChainAddressInput) (DeriveChainAddressOutput, error)
}
