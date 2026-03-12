package inbound

import (
	"context"

	"payrune/internal/application/dto"
)

type RunReceiptPollingCycleUseCase interface {
	Execute(ctx context.Context, input dto.RunReceiptPollingCycleInput) (dto.RunReceiptPollingCycleOutput, error)
}
