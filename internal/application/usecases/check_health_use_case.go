package usecases

import (
	"context"
	"errors"
	"time"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

type checkHealthUseCase struct {
	clock outport.Clock
}

func NewCheckHealthUseCase(clock outport.Clock) inport.CheckHealthUseCase {
	return &checkHealthUseCase{clock: clock}
}

func (uc *checkHealthUseCase) Execute(_ context.Context) (dto.HealthResponse, error) {
	if uc.clock == nil {
		return dto.HealthResponse{}, errors.New("clock is not configured")
	}
	now := uc.clock.NowUTC().Format(time.RFC3339)

	return dto.HealthResponse{
		Status:    string(valueobjects.ServiceStatusUp),
		Timestamp: now,
	}, nil
}
