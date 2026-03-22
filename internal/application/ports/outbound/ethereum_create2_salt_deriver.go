package outbound

import (
	"context"

	"payrune/internal/domain/valueobjects"
)

type DeriveEthereumCreate2SaltInput struct {
	Network          valueobjects.NetworkID
	AddressPolicyID  string
	PaymentAddressID int64
	DerivationIndex  uint32
}

type EthereumCreate2SaltDeriver interface {
	DeriveAllocationSalt(ctx context.Context, input DeriveEthereumCreate2SaltInput) (string, error)
}
