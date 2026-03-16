package inbound

import (
	"context"

	"payrune/internal/application/dto"
)

type RegisterEVMFactoryUseCase interface {
	Execute(ctx context.Context, input dto.RegisterEVMFactoryInput) (dto.RegisterEVMFactoryResponse, error)
}
