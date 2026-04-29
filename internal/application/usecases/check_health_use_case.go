package usecases

import (
	"context"

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

func (uc *checkHealthUseCase) Execute(_ context.Context) (inport.HealthResponse, error) {
	if uc.clock == nil {
		return inport.HealthResponse{}, inport.ErrClockNotConfigured
	}

	return inport.HealthResponse{
		Status:    healthStatusUp,
		Timestamp: uc.clock.NowUTC(),
	}, nil
}
