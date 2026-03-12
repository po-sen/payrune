package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
	controller := NewChainAddressController(
		listUC,
		&fakeGenerateAddressUseCase{},
		&fakeAllocatePaymentAddressUseCase{},
		&fakeGetPaymentAddressStatusUseCase{},
	)

	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/address-policies", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
	if listUC.lastChain != valueobjects.SupportedChainBitcoin {
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
	controller := NewChainAddressController(
		&fakeListAddressPoliciesUseCase{},
		generateUC,
		&fakeAllocatePaymentAddressUseCase{},
		&fakeGetPaymentAddressStatusUseCase{},
	)

	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)

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

func TestChainAddressControllerRejectMethod(t *testing.T) {
	controller := NewChainAddressController(
		&fakeListAddressPoliciesUseCase{},
		&fakeGenerateAddressUseCase{},
		&fakeAllocatePaymentAddressUseCase{},
		&fakeGetPaymentAddressStatusUseCase{},
	)
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
	controller := NewChainAddressController(
		&fakeListAddressPoliciesUseCase{},
		&fakeGenerateAddressUseCase{},
		&fakeAllocatePaymentAddressUseCase{},
		&fakeGetPaymentAddressStatusUseCase{},
	)
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
	controller := NewChainAddressController(
		&fakeListAddressPoliciesUseCase{},
		&fakeGenerateAddressUseCase{},
		&fakeAllocatePaymentAddressUseCase{},
		&fakeGetPaymentAddressStatusUseCase{},
	)
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
	controller := NewChainAddressController(
		&fakeListAddressPoliciesUseCase{},
		&fakeGenerateAddressUseCase{},
		&fakeAllocatePaymentAddressUseCase{},
		&fakeGetPaymentAddressStatusUseCase{},
	)
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
	controller := NewChainAddressController(
		&fakeListAddressPoliciesUseCase{},
		&fakeGenerateAddressUseCase{},
		&fakeAllocatePaymentAddressUseCase{},
		&fakeGetPaymentAddressStatusUseCase{},
	)
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
			controller := NewChainAddressController(
				&fakeListAddressPoliciesUseCase{},
				&fakeGenerateAddressUseCase{err: tc.err},
				&fakeAllocatePaymentAddressUseCase{},
				&fakeGetPaymentAddressStatusUseCase{},
			)
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
	controller := NewChainAddressController(
		&fakeListAddressPoliciesUseCase{},
		&fakeGenerateAddressUseCase{},
		allocateUC,
		&fakeGetPaymentAddressStatusUseCase{},
	)

	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)

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
		t.Fatalf(
			"unexpected expected amount minor in input: got %d",
			allocateUC.lastInput.ExpectedAmountMinor,
		)
	}
	if allocateUC.lastInput.CustomerReference != "order-20260304-001" {
		t.Fatalf("unexpected customer reference in input: got %q", allocateUC.lastInput.CustomerReference)
	}
	if allocateUC.lastInput.IdempotencyKey != "idem-101" {
		t.Fatalf("unexpected idempotency key in input: got %q", allocateUC.lastInput.IdempotencyKey)
	}

	var response dto.AllocatePaymentAddressResponse
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
	controller := NewChainAddressController(
		&fakeListAddressPoliciesUseCase{},
		&fakeGenerateAddressUseCase{},
		allocateUC,
		&fakeGetPaymentAddressStatusUseCase{},
	)

	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)

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
	controller := NewChainAddressController(
		&fakeListAddressPoliciesUseCase{},
		&fakeGenerateAddressUseCase{},
		&fakeAllocatePaymentAddressUseCase{},
		&fakeGetPaymentAddressStatusUseCase{},
	)
	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/payment-addresses", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
	if allow := rr.Header().Get("Allow"); allow != http.MethodPost {
		t.Fatalf("unexpected Allow header: got %q", allow)
	}
}

func TestChainAddressControllerAllocatePaymentAddressRejectInvalidBody(t *testing.T) {
	controller := NewChainAddressController(
		&fakeListAddressPoliciesUseCase{},
		&fakeGenerateAddressUseCase{},
		&fakeAllocatePaymentAddressUseCase{},
		&fakeGetPaymentAddressStatusUseCase{},
	)
	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/chains/bitcoin/payment-addresses",
		strings.NewReader(`{"addressPolicyId":"bitcoin-mainnet-legacy","unknown":"value"}`),
	)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
}

func TestChainAddressControllerAllocatePaymentAddressRejectMissingAddressPolicyID(t *testing.T) {
	controller := NewChainAddressController(
		&fakeListAddressPoliciesUseCase{},
		&fakeGenerateAddressUseCase{},
		&fakeAllocatePaymentAddressUseCase{},
		&fakeGetPaymentAddressStatusUseCase{},
	)
	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/chains/bitcoin/payment-addresses",
		strings.NewReader(`{"addressPolicyId":"   ","expectedAmountMinor":1}`),
	)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
}

func TestChainAddressControllerAllocatePaymentAddressRejectMissingExpectedAmountMinor(t *testing.T) {
	controller := NewChainAddressController(
		&fakeListAddressPoliciesUseCase{},
		&fakeGenerateAddressUseCase{},
		&fakeAllocatePaymentAddressUseCase{},
		&fakeGetPaymentAddressStatusUseCase{},
	)
	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/chains/bitcoin/payment-addresses",
		strings.NewReader(`{"addressPolicyId":"bitcoin-mainnet-legacy"}`),
	)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
}

func TestChainAddressControllerAllocatePaymentAddressErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
	}{
		{name: "policy not found", err: inport.ErrAddressPolicyNotFound, statusCode: http.StatusBadRequest},
		{name: "policy not enabled", err: inport.ErrAddressPolicyNotEnabled, statusCode: http.StatusNotImplemented},
		{name: "chain not supported", err: inport.ErrChainNotSupported, statusCode: http.StatusNotFound},
		{name: "pool exhausted", err: inport.ErrAddressPoolExhausted, statusCode: http.StatusConflict},
		{name: "idempotency key conflict", err: inport.ErrIdempotencyKeyConflict, statusCode: http.StatusConflict},
		{name: "invalid expected amount", err: inport.ErrInvalidExpectedAmount, statusCode: http.StatusBadRequest},
		{name: "internal", err: errors.New("boom"), statusCode: http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			controller := NewChainAddressController(
				&fakeListAddressPoliciesUseCase{},
				&fakeGenerateAddressUseCase{},
				&fakeAllocatePaymentAddressUseCase{err: tc.err},
				&fakeGetPaymentAddressStatusUseCase{},
			)
			mux := http.NewServeMux()
			controller.RegisterRoutes(mux)

			req := httptest.NewRequest(
				http.MethodPost,
				"/v1/chains/bitcoin/payment-addresses",
				strings.NewReader(`{"addressPolicyId":"bitcoin-mainnet-legacy","expectedAmountMinor":1}`),
			)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != tc.statusCode {
				t.Fatalf("unexpected status code: got %d, want %d", rr.Code, tc.statusCode)
			}
		})
	}
}

func TestChainAddressControllerListInternalError(t *testing.T) {
	controller := NewChainAddressController(
		&fakeListAddressPoliciesUseCase{err: errors.New("boom")},
		&fakeGenerateAddressUseCase{},
		&fakeAllocatePaymentAddressUseCase{},
		&fakeGetPaymentAddressStatusUseCase{},
	)
	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/address-policies", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
}

func TestChainAddressControllerGetPaymentStatusSuccess(t *testing.T) {
	issuedAt := time.Date(2026, 3, 8, 11, 0, 0, 0, time.UTC)
	firstObservedAt := issuedAt.Add(5 * time.Minute)
	getStatusUC := &fakeGetPaymentAddressStatusUseCase{
		response: dto.GetPaymentAddressStatusResponse{
			PaymentAddressID:        "101",
			AddressPolicyID:         "bitcoin-mainnet-native-segwit",
			ExpectedAmountMinor:     120000,
			Chain:                   "bitcoin",
			Network:                 "mainnet",
			Scheme:                  "nativeSegwit",
			MinorUnit:               "satoshi",
			Decimals:                8,
			Address:                 "bc1qstatus",
			CustomerReference:       "order-20260308-001",
			PaymentStatus:           "paid_unconfirmed_reverted",
			ObservedTotalMinor:      80000,
			ConfirmedTotalMinor:     40000,
			UnconfirmedTotalMinor:   40000,
			RequiredConfirmations:   1,
			LastObservedBlockHeight: 123,
			IssuedAt:                issuedAt,
			FirstObservedAt:         &firstObservedAt,
		},
	}
	controller := NewChainAddressController(
		&fakeListAddressPoliciesUseCase{},
		&fakeGenerateAddressUseCase{},
		&fakeAllocatePaymentAddressUseCase{},
		getStatusUC,
	)

	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/payment-addresses/101", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
	if getStatusUC.lastInput.Chain != valueobjects.SupportedChainBitcoin {
		t.Fatalf("unexpected chain in input: got %q", getStatusUC.lastInput.Chain)
	}
	if getStatusUC.lastInput.PaymentAddressID != 101 {
		t.Fatalf("unexpected payment address id in input: got %d", getStatusUC.lastInput.PaymentAddressID)
	}

	var response dto.GetPaymentAddressStatusResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.PaymentStatus != "paid_unconfirmed_reverted" {
		t.Fatalf("unexpected payment status: got %q", response.PaymentStatus)
	}
	if response.PaymentAddressID != "101" {
		t.Fatalf("unexpected payment address id: got %q", response.PaymentAddressID)
	}
}

func TestChainAddressControllerGetPaymentStatusRejectMethod(t *testing.T) {
	controller := NewChainAddressController(
		&fakeListAddressPoliciesUseCase{},
		&fakeGenerateAddressUseCase{},
		&fakeAllocatePaymentAddressUseCase{},
		&fakeGetPaymentAddressStatusUseCase{},
	)
	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/v1/chains/bitcoin/payment-addresses/101", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
	if allow := rr.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("unexpected Allow header: got %q", allow)
	}
}

func TestChainAddressControllerGetPaymentStatusRejectInvalidPaymentAddressID(t *testing.T) {
	controller := NewChainAddressController(
		&fakeListAddressPoliciesUseCase{},
		&fakeGenerateAddressUseCase{},
		&fakeAllocatePaymentAddressUseCase{},
		&fakeGetPaymentAddressStatusUseCase{},
	)
	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/payment-addresses/not-a-number", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
}

func TestChainAddressControllerGetPaymentStatusErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
	}{
		{name: "not found", err: inport.ErrPaymentAddressNotFound, statusCode: http.StatusNotFound},
		{name: "internal", err: errors.New("boom"), statusCode: http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			controller := NewChainAddressController(
				&fakeListAddressPoliciesUseCase{},
				&fakeGenerateAddressUseCase{},
				&fakeAllocatePaymentAddressUseCase{},
				&fakeGetPaymentAddressStatusUseCase{err: tc.err},
			)
			mux := http.NewServeMux()
			controller.RegisterRoutes(mux)

			req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/payment-addresses/101", nil)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != tc.statusCode {
				t.Fatalf("unexpected status code: got %d, want %d", rr.Code, tc.statusCode)
			}
		})
	}
}
