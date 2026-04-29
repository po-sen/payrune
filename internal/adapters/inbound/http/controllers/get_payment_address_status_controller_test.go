package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	inport "payrune/internal/application/ports/inbound"
)

func TestChainAddressControllerGetPaymentStatusSuccess(t *testing.T) {
	issuedAt := time.Date(2026, 3, 8, 11, 0, 0, 0, time.UTC)
	firstObservedAt := issuedAt.Add(5 * time.Minute)
	getStatusUC := &fakeGetPaymentAddressStatusUseCase{
		response: inport.GetPaymentAddressStatusResponse{
			PaymentAddressID:        "101",
			AddressPolicyID:         "bitcoin-mainnet-native-segwit",
			ExpectedAmountMinor:     120000,
			Chain:                   "bitcoin",
			Network:                 "mainnet",
			Scheme:                  "nativeSegwit",
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

	mux := http.NewServeMux()
	mux.Handle(
		"/v1/chains/{chain}/payment-addresses/{paymentAddressId}",
		NewGetPaymentAddressStatusController(getStatusUC),
	)

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/payment-addresses/101", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
	if getStatusUC.lastInput.Chain != "bitcoin" {
		t.Fatalf("unexpected chain in input: got %q", getStatusUC.lastInput.Chain)
	}
	if getStatusUC.lastInput.PaymentAddressID != 101 {
		t.Fatalf("unexpected payment address id in input: got %d", getStatusUC.lastInput.PaymentAddressID)
	}

	var response paymentAddressStatusResponse
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
	mux := http.NewServeMux()
	mux.Handle(
		"/v1/chains/{chain}/payment-addresses/{paymentAddressId}",
		NewGetPaymentAddressStatusController(&fakeGetPaymentAddressStatusUseCase{}),
	)

	req := httptest.NewRequest(http.MethodPost, "/v1/chains/bitcoin/payment-addresses/101", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
	if allow := rr.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("unexpected Allow header: got %q", allow)
	}
	assertErrorResponse(t, rr, http.StatusMethodNotAllowed, "method not allowed")
}

func TestChainAddressControllerGetPaymentStatusRejectInvalidPaymentAddressID(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle(
		"/v1/chains/{chain}/payment-addresses/{paymentAddressId}",
		NewGetPaymentAddressStatusController(&fakeGetPaymentAddressStatusUseCase{}),
	)

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/payment-addresses/not-a-number", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusBadRequest, "invalid paymentAddressId")
}

func TestChainAddressControllerGetPaymentStatusErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
		message    string
	}{
		{name: "not found", err: inport.ErrPaymentAddressNotFound, statusCode: http.StatusNotFound, message: "payment address is not found"},
		{name: "internal", err: inport.ErrDependencyFailure, statusCode: http.StatusInternalServerError, message: "internal server error"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.Handle(
				"/v1/chains/{chain}/payment-addresses/{paymentAddressId}",
				NewGetPaymentAddressStatusController(&fakeGetPaymentAddressStatusUseCase{err: tc.err}),
			)

			req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/payment-addresses/101", nil)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			assertErrorResponse(t, rr, tc.statusCode, tc.message)
		})
	}
}
