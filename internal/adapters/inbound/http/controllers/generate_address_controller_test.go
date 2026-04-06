package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
	"payrune/internal/domain/valueobjects"
)

func TestChainAddressControllerGenerateSuccess(t *testing.T) {
	generateUC := &fakeGenerateAddressUseCase{
		response: dto.GenerateAddressResponse{
			AddressPolicyID: "bitcoin-mainnet-legacy",
			Chain:           "bitcoin",
			Network:         "mainnet",
			Scheme:          "legacy",
			Decimals:        8,
			Index:           0,
			Address:         "1BitcoinAddressExample",
		},
	}

	mux := http.NewServeMux()
	mux.Handle("/v1/chains/{chain}/addresses", NewGenerateAddressController(generateUC))

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/addresses?addressPolicyId=bitcoin-mainnet-legacy&index=0", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
	if generateUC.lastInput.Chain != valueobjects.SupportedChainBitcoin {
		t.Fatalf("unexpected chain in input: got %q", generateUC.lastInput.Chain)
	}
	if generateUC.lastInput.AddressPolicyID != "bitcoin-mainnet-legacy" {
		t.Fatalf("unexpected address policy id: got %q", generateUC.lastInput.AddressPolicyID)
	}
	if generateUC.lastInput.Index != 0 {
		t.Fatalf("unexpected index: got %d", generateUC.lastInput.Index)
	}
}

func TestChainAddressControllerRejectsEthereumPreview(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle(
		"/v1/chains/{chain}/addresses",
		NewGenerateAddressController(&fakeGenerateAddressUseCase{err: inport.ErrAddressPreviewNotSupported}),
	)

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/ethereum/addresses?addressPolicyId=ethereum-mainnet-create2&index=0", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusNotFound, "address preview is not supported for this address policy")
}

func TestChainAddressControllerRejectMissingAddressPolicyID(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/v1/chains/{chain}/addresses", NewGenerateAddressController(&fakeGenerateAddressUseCase{}))

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/addresses?index=1", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusBadRequest, "addressPolicyId is required")
}

func TestChainAddressControllerRejectInvalidIndex(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/v1/chains/{chain}/addresses", NewGenerateAddressController(&fakeGenerateAddressUseCase{}))

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/addresses?addressPolicyId=bitcoin-mainnet-legacy&index=2147483648", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusBadRequest, "invalid index")
}

func TestChainAddressControllerGenerateErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
		message    string
	}{
		{name: "invalid address policy id", err: inport.ErrInvalidAddressPolicyID, statusCode: http.StatusBadRequest, message: "addressPolicyId is invalid"},
		{name: "policy not found", err: inport.ErrAddressPolicyNotFound, statusCode: http.StatusBadRequest, message: "address policy is not supported"},
		{name: "policy not enabled", err: inport.ErrAddressPolicyNotEnabled, statusCode: http.StatusNotImplemented, message: "address policy is not enabled"},
		{name: "preview not supported", err: inport.ErrAddressPreviewNotSupported, statusCode: http.StatusNotFound, message: "address preview is not supported for this address policy"},
		{name: "chain not supported", err: inport.ErrChainNotSupported, statusCode: http.StatusNotFound, message: publicUnsupportedChainMessage},
		{name: "internal", err: inport.ErrDependencyFailure, statusCode: http.StatusInternalServerError, message: "internal server error"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.Handle(
				"/v1/chains/{chain}/addresses",
				NewGenerateAddressController(&fakeGenerateAddressUseCase{err: tc.err}),
			)

			req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/addresses?addressPolicyId=bitcoin-mainnet-legacy&index=1", nil)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			assertErrorResponse(t, rr, tc.statusCode, tc.message)
		})
	}
}
