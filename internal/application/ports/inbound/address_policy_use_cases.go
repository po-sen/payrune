package inbound

import (
	"context"

	"payrune/internal/application/dto"
	"payrune/internal/domain/valueobjects"
)

type ListAddressPoliciesUseCase interface {
	Execute(ctx context.Context, chain valueobjects.SupportedChain) (dto.ListAddressPoliciesResponse, error)
}

type AllocatePaymentAddressUseCase interface {
	Execute(
		ctx context.Context,
		input dto.AllocatePaymentAddressInput,
	) (dto.AllocatePaymentAddressResponse, error)
}

type GetPaymentAddressStatusUseCase interface {
	Execute(
		ctx context.Context,
		input dto.GetPaymentAddressStatusInput,
	) (dto.GetPaymentAddressStatusResponse, error)
}
