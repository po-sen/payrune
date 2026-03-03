package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/in"
	"payrune/internal/domain/value_objects"
)

type fakeListAddressPoliciesUseCase struct {
	response  dto.ListAddressPoliciesResponse
	err       error
	lastChain value_objects.Chain
}

func (f *fakeListAddressPoliciesUseCase) Execute(
	_ context.Context,
	chain value_objects.Chain,
) (dto.ListAddressPoliciesResponse, error) {
	f.lastChain = chain
	if f.err != nil {
		return dto.ListAddressPoliciesResponse{}, f.err
	}
	return f.response, nil
}

type fakeGenerateAddressUseCase struct {
	response  dto.GenerateAddressResponse
	err       error
	lastInput dto.GenerateAddressInput
}

func (f *fakeGenerateAddressUseCase) Execute(
	_ context.Context,
	input dto.GenerateAddressInput,
) (dto.GenerateAddressResponse, error) {
	f.lastInput = input
	if f.err != nil {
		return dto.GenerateAddressResponse{}, f.err
	}
	return f.response, nil
}

func TestChainAddressControllerListSuccess(t *testing.T) {
	listUC := &fakeListAddressPoliciesUseCase{
		response: dto.ListAddressPoliciesResponse{
			Chain: "bitcoin",
			AddressPolicies: []dto.AddressPolicy{{
				AddressPolicyID: "bitcoin-mainnet-legacy",
				Chain:           "bitcoin",
				Network:         "mainnet",
				Scheme:          "legacy",
				MinorUnit:       "satoshi",
				Decimals:        8,
				Enabled:         true,
			}},
		},
	}
	controller := NewChainAddressController(listUC, &fakeGenerateAddressUseCase{})

	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/address-policies", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
	if listUC.lastChain != value_objects.ChainBitcoin {
		t.Fatalf("unexpected chain passed to use case: got %q", listUC.lastChain)
	}

	var body dto.ListAddressPoliciesResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(body.AddressPolicies) != 1 {
		t.Fatalf("unexpected policy count: got %d", len(body.AddressPolicies))
	}
	if body.AddressPolicies[0].MinorUnit != "satoshi" {
		t.Fatalf("unexpected minor unit: got %q", body.AddressPolicies[0].MinorUnit)
	}
	if body.AddressPolicies[0].Decimals != 8 {
		t.Fatalf("unexpected decimals: got %d", body.AddressPolicies[0].Decimals)
	}
}

func TestChainAddressControllerGenerateSuccess(t *testing.T) {
	generateUC := &fakeGenerateAddressUseCase{
		response: dto.GenerateAddressResponse{
			AddressPolicyID: "bitcoin-mainnet-legacy",
			Chain:           "bitcoin",
			Network:         "mainnet",
			Scheme:          "legacy",
			MinorUnit:       "satoshi",
			Decimals:        8,
			Index:           0,
			Address:         "1BitcoinAddressExample",
		},
	}
	controller := NewChainAddressController(&fakeListAddressPoliciesUseCase{}, generateUC)

	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/addresses?addressPolicyId=bitcoin-mainnet-legacy&index=0", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
	if generateUC.lastInput.Chain != value_objects.ChainBitcoin {
		t.Fatalf("unexpected chain in input: got %q", generateUC.lastInput.Chain)
	}
	if generateUC.lastInput.AddressPolicyID != "bitcoin-mainnet-legacy" {
		t.Fatalf("unexpected address policy id: got %q", generateUC.lastInput.AddressPolicyID)
	}
	if generateUC.lastInput.Index != 0 {
		t.Fatalf("unexpected index: got %d", generateUC.lastInput.Index)
	}
}

func TestChainAddressControllerRejectMethod(t *testing.T) {
	controller := NewChainAddressController(&fakeListAddressPoliciesUseCase{}, &fakeGenerateAddressUseCase{})
	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/v1/chains/bitcoin/address-policies", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
	if allow := rr.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("unexpected Allow header: got %q", allow)
	}
}

func TestChainAddressControllerRejectInvalidPath(t *testing.T) {
	controller := NewChainAddressController(&fakeListAddressPoliciesUseCase{}, &fakeGenerateAddressUseCase{})
	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
}

func TestChainAddressControllerRejectUnknownChain(t *testing.T) {
	controller := NewChainAddressController(&fakeListAddressPoliciesUseCase{}, &fakeGenerateAddressUseCase{})
	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/eth/address-policies", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}

	var body dto.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Error != inport.ErrChainNotSupported.Error() {
		t.Fatalf("unexpected error message: got %q", body.Error)
	}
}

func TestChainAddressControllerRejectMissingAddressPolicyID(t *testing.T) {
	controller := NewChainAddressController(&fakeListAddressPoliciesUseCase{}, &fakeGenerateAddressUseCase{})
	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/addresses?index=1", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
}

func TestChainAddressControllerRejectInvalidIndex(t *testing.T) {
	controller := NewChainAddressController(&fakeListAddressPoliciesUseCase{}, &fakeGenerateAddressUseCase{})
	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/addresses?addressPolicyId=bitcoin-mainnet-legacy&index=2147483648", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
}

func TestChainAddressControllerGenerateErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
	}{
		{name: "policy not found", err: inport.ErrAddressPolicyNotFound, statusCode: http.StatusBadRequest},
		{name: "policy not enabled", err: inport.ErrAddressPolicyNotEnabled, statusCode: http.StatusNotImplemented},
		{name: "chain not supported", err: inport.ErrChainNotSupported, statusCode: http.StatusNotFound},
		{name: "internal", err: errors.New("boom"), statusCode: http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			controller := NewChainAddressController(&fakeListAddressPoliciesUseCase{}, &fakeGenerateAddressUseCase{err: tc.err})
			mux := http.NewServeMux()
			controller.RegisterRoutes(mux)

			req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/addresses?addressPolicyId=bitcoin-mainnet-legacy&index=1", nil)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != tc.statusCode {
				t.Fatalf("unexpected status code: got %d, want %d", rr.Code, tc.statusCode)
			}
		})
	}
}

func TestChainAddressControllerListInternalError(t *testing.T) {
	controller := NewChainAddressController(&fakeListAddressPoliciesUseCase{err: errors.New("boom")}, &fakeGenerateAddressUseCase{})
	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/address-policies", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
}
