package use_cases

import (
	"context"
	"time"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/in"
	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/value_objects"
)

type checkHealthUseCase struct {
	clock outport.Clock
}

func NewCheckHealthUseCase(clock outport.Clock) inport.CheckHealthUseCase {
	return &checkHealthUseCase{clock: clock}
}

func (uc *checkHealthUseCase) Execute(_ context.Context) (dto.HealthResponse, error) {
	now := uc.clock.NowUTC().Format(time.RFC3339)

	return dto.HealthResponse{
		Status:    string(value_objects.ServiceStatusUp),
		Timestamp: now,
	}, nil
}
