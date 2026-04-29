package httpadapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"payrune/internal/adapters/inbound/http/controllers"
	inport "payrune/internal/application/ports/inbound"
)

type fakeCheckHealthUseCase struct{}

func (f *fakeCheckHealthUseCase) Execute(context.Context) (inport.HealthResponse, error) {
	return inport.HealthResponse{Status: "ok"}, nil
}

type fakeListAddressPoliciesUseCase struct{}

func (f *fakeListAddressPoliciesUseCase) Execute(
	context.Context,
	string,
) (inport.ListAddressPoliciesResponse, error) {
	return inport.ListAddressPoliciesResponse{}, nil
}

type fakeAllocatePaymentAddressUseCase struct{}

func (f *fakeAllocatePaymentAddressUseCase) Execute(
	context.Context,
	inport.AllocatePaymentAddressInput,
) (inport.AllocatePaymentAddressResponse, error) {
	return inport.AllocatePaymentAddressResponse{}, nil
}

type fakeGetPaymentAddressStatusUseCase struct{}

func (f *fakeGetPaymentAddressStatusUseCase) Execute(
	context.Context,
	inport.GetPaymentAddressStatusInput,
) (inport.GetPaymentAddressStatusResponse, error) {
	return inport.GetPaymentAddressStatusResponse{}, nil
}

func TestNewPublicRouterRegistersRoutesAndAppliesCORS(t *testing.T) {
	router := NewPublicRouter(RouterControllers{
		Health:              controllers.NewHealthController(&fakeCheckHealthUseCase{}),
		ListAddressPolicies: controllers.NewListAddressPoliciesController(&fakeListAddressPoliciesUseCase{}),
	})

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	request.Header.Set("Origin", "http://localhost:8081")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:8081" {
		t.Fatalf("unexpected allowed origin header: %q", got)
	}
}

func TestRouterRegistersChainAddressRoutes(t *testing.T) {
	router := newRouter(RouterControllers{
		ListAddressPolicies:    controllers.NewListAddressPoliciesController(&fakeListAddressPoliciesUseCase{}),
		AllocatePaymentAddress: controllers.NewAllocatePaymentAddressController(&fakeAllocatePaymentAddressUseCase{}),
		GetPaymentAddressStatus: controllers.NewGetPaymentAddressStatusController(
			&fakeGetPaymentAddressStatusUseCase{},
		),
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/address-policies", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

var (
	_ inport.CheckHealthUseCase             = (*fakeCheckHealthUseCase)(nil)
	_ inport.ListAddressPoliciesUseCase     = (*fakeListAddressPoliciesUseCase)(nil)
	_ inport.AllocatePaymentAddressUseCase  = (*fakeAllocatePaymentAddressUseCase)(nil)
	_ inport.GetPaymentAddressStatusUseCase = (*fakeGetPaymentAddressStatusUseCase)(nil)
)
