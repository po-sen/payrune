package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	inport "payrune/internal/application/ports/inbound"
)

const idempotencyKeyHeader = "Idempotency-Key"
const idempotencyReplayedHeader = "Idempotency-Replayed"

type allocatePaymentAddressRequest struct {
	AddressPolicyID     string  `json:"addressPolicyId"`
	ExpectedAmountMinor *int64  `json:"expectedAmountMinor"`
	CustomerReference   *string `json:"customerReference"`
}

type AllocatePaymentAddressController struct {
	allocatePaymentAddress inport.AllocatePaymentAddressUseCase
}

func NewAllocatePaymentAddressController(
	allocatePaymentAddress inport.AllocatePaymentAddressUseCase,
) *AllocatePaymentAddressController {
	return &AllocatePaymentAddressController{allocatePaymentAddress: allocatePaymentAddress}
}

func (c *AllocatePaymentAddressController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chain, ok := parseSupportedChainPathValue(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var request allocatePaymentAddressRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	addressPolicyID := strings.TrimSpace(request.AddressPolicyID)
	if addressPolicyID == "" {
		writeErrorJSON(w, http.StatusBadRequest, "addressPolicyId is required")
		return
	}
	if request.ExpectedAmountMinor == nil {
		writeErrorJSON(w, http.StatusBadRequest, "expectedAmountMinor is required")
		return
	}

	response, err := c.allocatePaymentAddress.Execute(r.Context(), inport.AllocatePaymentAddressInput{
		Chain:               chain,
		AddressPolicyID:     addressPolicyID,
		ExpectedAmountMinor: *request.ExpectedAmountMinor,
		CustomerReference:   trimOptionalString(request.CustomerReference),
		IdempotencyKey:      strings.TrimSpace(r.Header.Get(idempotencyKeyHeader)),
	})
	if err != nil {
		statusCode, message := mapAllocatePaymentAddressError(err)
		logMappedControllerError(r, statusCode, message, err)
		writeErrorJSON(w, statusCode, message)
		return
	}

	if response.IdempotencyReplayed {
		w.Header().Set(idempotencyReplayedHeader, "true")
	}
	writeJSON(w, http.StatusCreated, newAllocatePaymentAddressResponse(response))
}

func trimOptionalString(raw *string) string {
	if raw == nil {
		return ""
	}
	return strings.TrimSpace(*raw)
}

func mapAllocatePaymentAddressError(err error) (int, string) {
	switch {
	case errors.Is(err, inport.ErrChainNotSupported):
		return http.StatusNotFound, publicUnsupportedChainMessage
	case errors.Is(err, inport.ErrInvalidAddressPolicyID):
		return http.StatusBadRequest, "addressPolicyId is invalid"
	case errors.Is(err, inport.ErrAddressPolicyNotFound):
		return http.StatusNotFound, "address policy is not supported"
	case errors.Is(err, inport.ErrAddressPolicyNotEnabled):
		return http.StatusConflict, "address policy is not enabled"
	case errors.Is(err, inport.ErrAddressPoolExhausted):
		return http.StatusConflict, "address pool is exhausted"
	case errors.Is(err, inport.ErrIdempotencyKeyConflict):
		return http.StatusConflict, "idempotency key conflicts with existing payment address"
	case errors.Is(err, inport.ErrInvalidExpectedAmount):
		return http.StatusBadRequest, "expected amount is invalid"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
