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
	ErrAddressPoolExhausted    = errors.New("address pool is exhausted")
	ErrInvalidExpectedAmount   = errors.New("expected amount is invalid")
	ErrIdempotencyKeyConflict  = errors.New("idempotency key conflicts with existing payment address")
	ErrPaymentAddressNotFound  = errors.New("payment address is not found")
)

type ListAddressPoliciesUseCase interface {
	Execute(ctx context.Context, chain value_objects.SupportedChain) (dto.ListAddressPoliciesResponse, error)
}

type GenerateAddressUseCase interface {
	Execute(ctx context.Context, input dto.GenerateAddressInput) (dto.GenerateAddressResponse, error)
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
