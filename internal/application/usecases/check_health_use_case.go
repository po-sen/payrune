package usecases

import (
	"context"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
	outport "payrune/internal/application/ports/outbound"
)

const healthStatusUp = "up"

type checkHealthUseCase struct {
	clock outport.Clock
}

func NewCheckHealthUseCase(clock outport.Clock) inport.CheckHealthUseCase {
	return &checkHealthUseCase{clock: clock}
}

func (uc *checkHealthUseCase) Execute(_ context.Context) (dto.HealthResponse, error) {
	if uc.clock == nil {
		return dto.HealthResponse{}, inport.ErrClockNotConfigured
	}

	return dto.HealthResponse{
		Status:    healthStatusUp,
		Timestamp: uc.clock.NowUTC(),
	}, nil
}
