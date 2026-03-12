package outbound

import (
	"context"

	"payrune/internal/domain/entities"
	"payrune/internal/domain/valueobjects"
)

type AddressPolicyReader interface {
	ListByChain(ctx context.Context, chain valueobjects.SupportedChain) ([]entities.AddressPolicy, error)
	FindIssuanceByID(ctx context.Context, addressPolicyID string) (entities.AddressIssuancePolicy, bool, error)
}
