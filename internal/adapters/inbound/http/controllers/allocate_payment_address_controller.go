package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
)

const idempotencyKeyHeader = "Idempotency-Key"
const idempotencyReplayedHeader = "Idempotency-Replayed"

type allocatePaymentAddressRequest struct {
	AddressPolicyID     string `json:"addressPolicyId"`
	ExpectedAmountMinor *int64 `json:"expectedAmountMinor"`
	CustomerReference   string `json:"customerReference,omitempty"`
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
		writeJSON(w, http.StatusMethodNotAllowed, dto.ErrorResponse{Error: "method not allowed"})
		return
	}

	var request allocatePaymentAddressRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request body"})
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request body"})
		return
	}

	addressPolicyID := strings.TrimSpace(request.AddressPolicyID)
	if addressPolicyID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "addressPolicyId is required"})
		return
	}
	if request.ExpectedAmountMinor == nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "expectedAmountMinor is required"})
		return
	}

	response, err := c.allocatePaymentAddress.Execute(r.Context(), dto.AllocatePaymentAddressInput{
		Chain:               chain,
		AddressPolicyID:     addressPolicyID,
		ExpectedAmountMinor: *request.ExpectedAmountMinor,
		CustomerReference:   strings.TrimSpace(request.CustomerReference),
		IdempotencyKey:      strings.TrimSpace(r.Header.Get(idempotencyKeyHeader)),
	})
	if err != nil {
		switch {
		case errors.Is(err, inport.ErrChainNotSupported):
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		case errors.Is(err, inport.ErrAddressPolicyNotFound):
			writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		case errors.Is(err, inport.ErrAddressPolicyNotEnabled):
			writeJSON(w, http.StatusNotImplemented, dto.ErrorResponse{Error: err.Error()})
		case errors.Is(err, inport.ErrAddressPoolExhausted):
			writeJSON(w, http.StatusConflict, dto.ErrorResponse{Error: err.Error()})
		case errors.Is(err, inport.ErrIdempotencyKeyConflict):
			writeJSON(w, http.StatusConflict, dto.ErrorResponse{Error: err.Error()})
		case errors.Is(err, inport.ErrInvalidExpectedAmount):
			writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	if response.IdempotencyReplayed {
		w.Header().Set(idempotencyReplayedHeader, "true")
	}
	writeJSON(w, http.StatusCreated, response)
}
