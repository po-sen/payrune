package httpadapter

import (
	"net/http"

	"payrune/internal/adapters/inbound/http/controllers"
	"payrune/internal/adapters/inbound/http/middleware"
)

type Dependencies struct {
	HealthController       *controllers.HealthController
	ChainAddressController *controllers.ChainAddressController
}

func NewHandler(deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	if deps.HealthController != nil {
		deps.HealthController.RegisterRoutes(mux)
	}
	if deps.ChainAddressController != nil {
		deps.ChainAddressController.RegisterRoutes(mux)
	}
	return mux
}

func NewPublicHandler(deps Dependencies) http.Handler {
	return middleware.CORS(NewHandler(deps))
}
