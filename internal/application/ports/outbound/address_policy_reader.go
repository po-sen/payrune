package outbound

import (
	"context"

	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
)

type AddressPolicyRecord struct {
	AddressPolicyID valueobjects.AddressPolicyID
	Chain           valueobjects.SupportedChain
	Network         valueobjects.NetworkID
	Scheme          valueobjects.AddressScheme
	AssetReference  string
	Decimals        uint8
	Enabled         bool
}

type AddressPolicyReader interface {
	ListByChain(ctx context.Context, chain valueobjects.SupportedChain) ([]AddressPolicyRecord, error)
	FindIssuanceByID(ctx context.Context, addressPolicyID valueobjects.AddressPolicyID) (policies.AddressIssuancePolicy, bool, error)
}
