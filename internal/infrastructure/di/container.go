package di

import (
	httpcontroller "payrune/internal/adapters/inbound/http/controllers"
	"payrune/internal/adapters/outbound/system"
	"payrune/internal/application/use_cases"
)

type Container struct {
	HealthController *httpcontroller.HealthController
}

func NewContainer() *Container {
	clock := system.NewClock()
	healthUseCase := use_cases.NewCheckHealthUseCase(clock)
	healthController := httpcontroller.NewHealthController(healthUseCase)

	return &Container{
		HealthController: healthController,
	}
}
