package out

import (
	"context"

	"payrune/internal/domain/entities"
	"payrune/internal/domain/value_objects"
)

type AddressPolicyReader interface {
	ListByChain(ctx context.Context, chain value_objects.Chain) ([]entities.AddressPolicy, error)
	FindByID(ctx context.Context, addressPolicyID string) (entities.AddressPolicy, bool, error)
}
