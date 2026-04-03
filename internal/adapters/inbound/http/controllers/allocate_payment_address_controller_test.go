package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
	"payrune/internal/domain/valueobjects"
)

func TestChainAddressControllerAllocatePaymentAddressSuccess(t *testing.T) {
	allocateUC := &fakeAllocatePaymentAddressUseCase{
		response: dto.AllocatePaymentAddressResponse{
			PaymentAddressID:    "101",
			AddressPolicyID:     "bitcoin-mainnet-native-segwit",
			ExpectedAmountMinor: 120000,
			Chain:               "bitcoin",
			Network:             "mainnet",
			Scheme:              "nativeSegwit",
			MinorUnit:           "satoshi",
			Decimals:            8,
			Address:             "bc1qallocatedaddress",
			CustomerReference:   "order-20260304-001",
		},
	}

	mux := http.NewServeMux()
	mux.Handle("/v1/chains/{chain}/payment-addresses", NewAllocatePaymentAddressController(allocateUC))

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/chains/bitcoin/payment-addresses",
		strings.NewReader(`{"addressPolicyId":"bitcoin-mainnet-native-segwit","expectedAmountMinor":120000,"customerReference":" order-20260304-001 "}`),
	)
	req.Header.Set(idempotencyKeyHeader, " idem-101 ")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
	if got := rr.Header().Get(idempotencyReplayedHeader); got != "" {
		t.Fatalf("expected no idempotency replayed header on fresh success, got %q", got)
	}
	if allocateUC.lastInput.Chain != valueobjects.SupportedChainBitcoin {
		t.Fatalf("unexpected chain in input: got %q", allocateUC.lastInput.Chain)
	}
	if allocateUC.lastInput.AddressPolicyID != "bitcoin-mainnet-native-segwit" {
		t.Fatalf("unexpected address policy id in input: got %q", allocateUC.lastInput.AddressPolicyID)
	}
	if allocateUC.lastInput.ExpectedAmountMinor != 120000 {
		t.Fatalf("unexpected expected amount minor in input: got %d", allocateUC.lastInput.ExpectedAmountMinor)
	}
	if allocateUC.lastInput.CustomerReference != "order-20260304-001" {
		t.Fatalf("unexpected customer reference in input: got %q", allocateUC.lastInput.CustomerReference)
	}
	if allocateUC.lastInput.IdempotencyKey != "idem-101" {
		t.Fatalf("unexpected idempotency key in input: got %q", allocateUC.lastInput.IdempotencyKey)
	}

	var response allocatePaymentAddressResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Address == "" {
		t.Fatalf("expected non-empty address")
	}
	if response.PaymentAddressID != "101" {
		t.Fatalf("unexpected payment address id: got %q", response.PaymentAddressID)
	}
	if response.ExpectedAmountMinor != 120000 {
		t.Fatalf("unexpected expected amount minor: got %d", response.ExpectedAmountMinor)
	}
}

func TestChainAddressControllerAllocateEthereumPaymentAddressSuccess(t *testing.T) {
	allocateUC := &fakeAllocatePaymentAddressUseCase{
		response: dto.AllocatePaymentAddressResponse{
			PaymentAddressID:    "201",
			AddressPolicyID:     "ethereum-mainnet-create2",
			ExpectedAmountMinor: 15000000000000000,
			Chain:               "ethereum",
			Network:             "mainnet",
			Scheme:              "create2",
			MinorUnit:           "wei",
			Decimals:            18,
			Address:             "0x1234567890abcdef1234567890abcdef12345678",
		},
	}

	mux := http.NewServeMux()
	mux.Handle("/v1/chains/{chain}/payment-addresses", NewAllocatePaymentAddressController(allocateUC))

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/chains/ethereum/payment-addresses",
		strings.NewReader(`{"addressPolicyId":"ethereum-mainnet-create2","expectedAmountMinor":15000000000000000}`),
	)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
	if allocateUC.lastInput.Chain != valueobjects.SupportedChainEthereum {
		t.Fatalf("unexpected chain in input: got %q", allocateUC.lastInput.Chain)
	}
	if allocateUC.lastInput.AddressPolicyID != "ethereum-mainnet-create2" {
		t.Fatalf("unexpected address policy id in input: got %q", allocateUC.lastInput.AddressPolicyID)
	}
}

func TestChainAddressControllerAllocatePaymentAddressAcceptsNullCustomerReference(t *testing.T) {
	allocateUC := &fakeAllocatePaymentAddressUseCase{
		response: dto.AllocatePaymentAddressResponse{
			PaymentAddressID:    "202",
			AddressPolicyID:     "bitcoin-testnet4-native-segwit",
			ExpectedAmountMinor: 2000,
			Chain:               "bitcoin",
			Network:             "testnet4",
			Scheme:              "nativeSegwit",
			MinorUnit:           "satoshi",
			Decimals:            8,
			Address:             "tb1qallocatedaddress",
		},
	}

	mux := http.NewServeMux()
	mux.Handle("/v1/chains/{chain}/payment-addresses", NewAllocatePaymentAddressController(allocateUC))

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/chains/bitcoin/payment-addresses",
		strings.NewReader(`{"addressPolicyId":"bitcoin-testnet4-native-segwit","expectedAmountMinor":2000,"customerReference":null}`),
	)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
	if allocateUC.lastInput.CustomerReference != "" {
		t.Fatalf("expected empty customer reference for null input, got %q", allocateUC.lastInput.CustomerReference)
	}
}

func TestChainAddressControllerAllocatePaymentAddressReplayHeader(t *testing.T) {
	allocateUC := &fakeAllocatePaymentAddressUseCase{
		response: dto.AllocatePaymentAddressResponse{
			PaymentAddressID:    "101",
			AddressPolicyID:     "bitcoin-mainnet-native-segwit",
			ExpectedAmountMinor: 120000,
			Chain:               "bitcoin",
			Network:             "mainnet",
			Scheme:              "nativeSegwit",
			MinorUnit:           "satoshi",
			Decimals:            8,
			Address:             "bc1qallocatedaddress",
			CustomerReference:   "order-20260304-001",
			IdempotencyReplayed: true,
		},
	}

	mux := http.NewServeMux()
	mux.Handle("/v1/chains/{chain}/payment-addresses", NewAllocatePaymentAddressController(allocateUC))

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/chains/bitcoin/payment-addresses",
		strings.NewReader(`{"addressPolicyId":"bitcoin-mainnet-native-segwit","expectedAmountMinor":120000}`),
	)
	req.Header.Set(idempotencyKeyHeader, "idem-replay")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
	if got := rr.Header().Get(idempotencyReplayedHeader); got != "true" {
		t.Fatalf("unexpected idempotency replayed header: got %q", got)
	}
}

func TestChainAddressControllerAllocatePaymentAddressRejectMethod(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle(
		"/v1/chains/{chain}/payment-addresses",
		NewAllocatePaymentAddressController(&fakeAllocatePaymentAddressUseCase{}),
	)

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/payment-addresses", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
	if allow := rr.Header().Get("Allow"); allow != http.MethodPost {
		t.Fatalf("unexpected Allow header: got %q", allow)
	}
	assertErrorResponse(t, rr, http.StatusMethodNotAllowed, "method not allowed")
}

func TestChainAddressControllerAllocatePaymentAddressRejectInvalidBody(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle(
		"/v1/chains/{chain}/payment-addresses",
		NewAllocatePaymentAddressController(&fakeAllocatePaymentAddressUseCase{}),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/chains/bitcoin/payment-addresses",
		strings.NewReader(`{"addressPolicyId":"bitcoin-mainnet-legacy","unknown":"value"}`),
	)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusBadRequest, "invalid request body")
}

func TestChainAddressControllerAllocatePaymentAddressRejectMissingAddressPolicyID(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle(
		"/v1/chains/{chain}/payment-addresses",
		NewAllocatePaymentAddressController(&fakeAllocatePaymentAddressUseCase{}),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/chains/bitcoin/payment-addresses",
		strings.NewReader(`{"addressPolicyId":"   ","expectedAmountMinor":1}`),
	)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusBadRequest, "addressPolicyId is required")
}

func TestChainAddressControllerAllocatePaymentAddressRejectMissingExpectedAmountMinor(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle(
		"/v1/chains/{chain}/payment-addresses",
		NewAllocatePaymentAddressController(&fakeAllocatePaymentAddressUseCase{}),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/chains/bitcoin/payment-addresses",
		strings.NewReader(`{"addressPolicyId":"bitcoin-mainnet-legacy"}`),
	)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusBadRequest, "expectedAmountMinor is required")
}

func TestChainAddressControllerAllocatePaymentAddressErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
		message    string
	}{
		{name: "invalid address policy id", err: inport.ErrInvalidAddressPolicyID, statusCode: http.StatusBadRequest, message: "addressPolicyId is invalid"},
		{name: "policy not found", err: inport.ErrAddressPolicyNotFound, statusCode: http.StatusBadRequest, message: "address policy is not supported"},
		{name: "policy not enabled", err: inport.ErrAddressPolicyNotEnabled, statusCode: http.StatusNotImplemented, message: "address policy is not enabled"},
		{name: "chain not supported", err: inport.ErrChainNotSupported, statusCode: http.StatusNotFound, message: publicUnsupportedChainMessage},
		{name: "pool exhausted", err: inport.ErrAddressPoolExhausted, statusCode: http.StatusConflict, message: "address pool is exhausted"},
		{name: "idempotency key conflict", err: inport.ErrIdempotencyKeyConflict, statusCode: http.StatusConflict, message: "idempotency key conflicts with existing payment address"},
		{name: "invalid expected amount", err: inport.ErrInvalidExpectedAmount, statusCode: http.StatusBadRequest, message: "expected amount is invalid"},
		{name: "internal", err: inport.ErrDependencyFailure, statusCode: http.StatusInternalServerError, message: "internal server error"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.Handle(
				"/v1/chains/{chain}/payment-addresses",
				NewAllocatePaymentAddressController(&fakeAllocatePaymentAddressUseCase{err: tc.err}),
			)

			req := httptest.NewRequest(
				http.MethodPost,
				"/v1/chains/bitcoin/payment-addresses",
				strings.NewReader(`{"addressPolicyId":"bitcoin-mainnet-legacy","expectedAmountMinor":1}`),
			)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			assertErrorResponse(t, rr, tc.statusCode, tc.message)
		})
	}
}
