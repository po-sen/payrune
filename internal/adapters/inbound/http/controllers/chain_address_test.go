package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"payrune/internal/application/dto"
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
	assertErrorResponse(t, rr, http.StatusNotFound, publicUnsupportedChainMessage)
}

func assertErrorResponse(
	t *testing.T,
	rr *httptest.ResponseRecorder,
	wantStatus int,
	wantError string,
) {
	t.Helper()

	if rr.Code != wantStatus {
		t.Fatalf("unexpected status code: got %d, want %d", rr.Code, wantStatus)
	}

	var body errorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Error != wantError {
		t.Fatalf("unexpected error message: got %q, want %q", body.Error, wantError)
	}
}

func captureControllerLogs(t *testing.T) *bytes.Buffer {
	t.Helper()

	var buffer bytes.Buffer
	previousWriter := log.Writer()
	previousFlags := log.Flags()
	log.SetOutput(&buffer)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(previousWriter)
		log.SetFlags(previousFlags)
	})
	return &buffer
}
