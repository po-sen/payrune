package httpadapter

import (
	"net/http"

	"payrune/internal/adapters/inbound/http/controllers"
	"payrune/internal/adapters/inbound/http/middleware"
)

type RouterControllers struct {
	Health       *controllers.HealthController
	ChainAddress *controllers.ChainAddressController
}

func newRouter(routeControllers RouterControllers) http.Handler {
	mux := http.NewServeMux()
	if routeControllers.Health != nil {
		mux.HandleFunc("/health", routeControllers.Health.HandleHealth)
	}
	if routeControllers.ChainAddress != nil {
		mux.HandleFunc("/v1/chains/{chain}/address-policies", routeControllers.ChainAddress.HandleListAddressPolicies)
		mux.HandleFunc("/v1/chains/{chain}/addresses", routeControllers.ChainAddress.HandleGenerateAddress)
		mux.HandleFunc("/v1/chains/{chain}/payment-addresses", routeControllers.ChainAddress.HandleAllocatePaymentAddress)
		mux.HandleFunc(
			"/v1/chains/{chain}/payment-addresses/{paymentAddressId}",
			routeControllers.ChainAddress.HandleGetPaymentAddressStatus,
		)
	}
	return mux
}

func NewPublicRouter(routeControllers RouterControllers) http.Handler {
	return middleware.CORS(newRouter(routeControllers))
}
