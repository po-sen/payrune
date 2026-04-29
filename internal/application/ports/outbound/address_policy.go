package outbound

import "context"

type AddressPolicyRecord struct {
	AddressPolicyID string
	Chain           string
	Network         string
	Scheme          string
	AssetReference  string
	Decimals        uint8
	Enabled         bool
}

type AddressIssuancePolicyRecord struct {
	AddressPolicyID   string
	Chain             string
	Network           string
	Scheme            string
	AssetReference    string
	Decimals          uint8
	Enabled           bool
	AddressSpaceRef   string
	IssuanceRefPrefix string
}

type AddressPolicyReader interface {
	ListByChain(ctx context.Context, chain string) ([]AddressPolicyRecord, error)
	FindIssuanceByID(ctx context.Context, addressPolicyID string) (AddressIssuancePolicyRecord, bool, error)
}
