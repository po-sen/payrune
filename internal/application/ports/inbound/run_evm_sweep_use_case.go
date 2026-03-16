package inbound

import (
	"context"

	"payrune/internal/application/dto"
)

type RunEVMSweepUseCase interface {
	Execute(ctx context.Context, input dto.RunEVMSweepInput) (dto.RunEVMSweepOutput, error)
}
