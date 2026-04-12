package httpadapter

import (
	"net/http"
	"payrune/internal/adapters/inbound/http/middleware"
)

type RouterControllers struct {
	Health                  http.Handler
	ListAddressPolicies     http.Handler
	AllocatePaymentAddress  http.Handler
	GetPaymentAddressStatus http.Handler
}

func newRouter(routeControllers RouterControllers) http.Handler {
	mux := http.NewServeMux()
	if routeControllers.Health != nil {
		mux.Handle("/health", routeControllers.Health)
	}
	if routeControllers.ListAddressPolicies != nil {
		mux.Handle("/v1/chains/{chain}/address-policies", routeControllers.ListAddressPolicies)
	}
	if routeControllers.AllocatePaymentAddress != nil {
		mux.Handle("/v1/chains/{chain}/payment-addresses", routeControllers.AllocatePaymentAddress)
	}
	if routeControllers.GetPaymentAddressStatus != nil {
		mux.Handle(
			"/v1/chains/{chain}/payment-addresses/{paymentAddressId}",
			routeControllers.GetPaymentAddressStatus,
		)
	}
	return mux
}

func NewPublicRouter(routeControllers RouterControllers) http.Handler {
	return middleware.RequestLog(middleware.CORS(newRouter(routeControllers)))
}
