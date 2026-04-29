package inbound

import (
	"context"
	"time"
)

type HealthResponse struct {
	Status    string
	Timestamp time.Time
}

type CheckHealthUseCase interface {
	Execute(ctx context.Context) (HealthResponse, error)
}
