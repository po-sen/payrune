package out

import (
	"context"

	"payrune/internal/domain/entities"
	"payrune/internal/domain/value_objects"
)

type AddressPolicyReader interface {
	ListByChain(ctx context.Context, chain value_objects.SupportedChain) ([]entities.AddressPolicy, error)
	FindIssuanceByID(ctx context.Context, addressPolicyID string) (entities.AddressIssuancePolicy, bool, error)
}
