package httpadapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"payrune/internal/adapters/inbound/http/controllers"
	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
	"payrune/internal/domain/valueobjects"
)

type fakeCheckHealthUseCase struct{}

func (f *fakeCheckHealthUseCase) Execute(context.Context) (dto.HealthResponse, error) {
	return dto.HealthResponse{Status: "ok"}, nil
}

type fakeListAddressPoliciesUseCase struct{}

func (f *fakeListAddressPoliciesUseCase) Execute(
	context.Context,
	valueobjects.SupportedChain,
) (dto.ListAddressPoliciesResponse, error) {
	return dto.ListAddressPoliciesResponse{}, nil
}

type fakeGenerateAddressUseCase struct{}

func (f *fakeGenerateAddressUseCase) Execute(
	context.Context,
	dto.GenerateAddressInput,
) (dto.GenerateAddressResponse, error) {
	return dto.GenerateAddressResponse{}, nil
}

type fakeAllocatePaymentAddressUseCase struct{}

func (f *fakeAllocatePaymentAddressUseCase) Execute(
	context.Context,
	dto.AllocatePaymentAddressInput,
) (dto.AllocatePaymentAddressResponse, error) {
	return dto.AllocatePaymentAddressResponse{}, nil
}

type fakeGetPaymentAddressStatusUseCase struct{}

func (f *fakeGetPaymentAddressStatusUseCase) Execute(
	context.Context,
	dto.GetPaymentAddressStatusInput,
) (dto.GetPaymentAddressStatusResponse, error) {
	return dto.GetPaymentAddressStatusResponse{}, nil
}

func TestNewPublicRouterRegistersRoutesAndAppliesCORS(t *testing.T) {
	router := NewPublicRouter(RouterControllers{
		Health: controllers.NewHealthController(&fakeCheckHealthUseCase{}),
		ChainAddress: controllers.NewChainAddressController(
			&fakeListAddressPoliciesUseCase{},
			&fakeGenerateAddressUseCase{},
			&fakeAllocatePaymentAddressUseCase{},
			&fakeGetPaymentAddressStatusUseCase{},
		),
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
		ChainAddress: controllers.NewChainAddressController(
			&fakeListAddressPoliciesUseCase{},
			&fakeGenerateAddressUseCase{},
			&fakeAllocatePaymentAddressUseCase{},
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
	_ inport.GenerateAddressUseCase         = (*fakeGenerateAddressUseCase)(nil)
	_ inport.AllocatePaymentAddressUseCase  = (*fakeAllocatePaymentAddressUseCase)(nil)
	_ inport.GetPaymentAddressStatusUseCase = (*fakeGetPaymentAddressStatusUseCase)(nil)
)
