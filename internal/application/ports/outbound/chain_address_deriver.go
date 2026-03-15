package outbound

import (
	"context"
	"payrune/internal/domain/valueobjects"
)

type DeriveChainAddressInput struct {
	Chain                valueobjects.SupportedChain
	Network              valueobjects.NetworkID
	Scheme               string
	AccountPublicKey     string
	DerivationPathPrefix string
	Index                uint32
}

type DeriveChainAddressOutput struct {
	Address                string
	RelativeDerivationPath string
	DerivationPath         string
}

type ChainAddressDeriver interface {
	SupportsChain(chain valueobjects.SupportedChain) bool
	DeriveAddress(ctx context.Context, input DeriveChainAddressInput) (DeriveChainAddressOutput, error)
}
