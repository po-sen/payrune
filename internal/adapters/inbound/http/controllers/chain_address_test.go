package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
	"payrune/internal/domain/valueobjects"
)

type fakeListAddressPoliciesUseCase struct {
	response  dto.ListAddressPoliciesResponse
	err       error
	lastChain valueobjects.SupportedChain
}

func (f *fakeListAddressPoliciesUseCase) Execute(
	_ context.Context,
	chain valueobjects.SupportedChain,
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

type fakeAllocatePaymentAddressUseCase struct {
	response  dto.AllocatePaymentAddressResponse
	err       error
	lastInput dto.AllocatePaymentAddressInput
}

func (f *fakeAllocatePaymentAddressUseCase) Execute(
	_ context.Context,
	input dto.AllocatePaymentAddressInput,
) (dto.AllocatePaymentAddressResponse, error) {
	f.lastInput = input
	if f.err != nil {
		return dto.AllocatePaymentAddressResponse{}, f.err
	}
	return f.response, nil
}

type fakeGetPaymentAddressStatusUseCase struct {
	response  dto.GetPaymentAddressStatusResponse
	err       error
	lastInput dto.GetPaymentAddressStatusInput
}

func (f *fakeGetPaymentAddressStatusUseCase) Execute(
	_ context.Context,
	input dto.GetPaymentAddressStatusInput,
) (dto.GetPaymentAddressStatusResponse, error) {
	f.lastInput = input
	if f.err != nil {
		return dto.GetPaymentAddressStatusResponse{}, f.err
	}
	return f.response, nil
}

func TestChainAddressRoutesRejectInvalidPath(t *testing.T) {
	mux := http.NewServeMux()

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
}

func TestChainAddressRoutesRejectUnknownChain(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle(
		"/v1/chains/{chain}/address-policies",
		NewListAddressPoliciesController(&fakeListAddressPoliciesUseCase{}),
	)

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
