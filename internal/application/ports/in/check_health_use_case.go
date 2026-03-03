package in

import (
	"context"

	"payrune/internal/application/dto"
)

type CheckHealthUseCase interface {
	Execute(ctx context.Context) (dto.HealthResponse, error)
}
