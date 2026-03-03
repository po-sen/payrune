package in

import (
	"context"
	"errors"

	"payrune/internal/application/dto"
	"payrune/internal/domain/value_objects"
)

var (
	ErrChainNotSupported       = errors.New("chain is not supported")
	ErrAddressPolicyNotFound   = errors.New("address policy is not supported")
	ErrAddressPolicyNotEnabled = errors.New("address policy is not enabled")
)

type ListAddressPoliciesUseCase interface {
	Execute(ctx context.Context, chain value_objects.Chain) (dto.ListAddressPoliciesResponse, error)
}

type GenerateAddressUseCase interface {
	Execute(ctx context.Context, input dto.GenerateAddressInput) (dto.GenerateAddressResponse, error)
}
