package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/in"
	"payrune/internal/domain/value_objects"
)

const maxNonHardenedIndex = uint64(0x7fffffff)

type ChainAddressController struct {
	listAddressPolicies    inport.ListAddressPoliciesUseCase
	generateAddress        inport.GenerateAddressUseCase
	allocatePaymentAddress inport.AllocatePaymentAddressUseCase
}

func NewChainAddressController(
	listAddressPolicies inport.ListAddressPoliciesUseCase,
	generateAddress inport.GenerateAddressUseCase,
	allocatePaymentAddress inport.AllocatePaymentAddressUseCase,
) *ChainAddressController {
	return &ChainAddressController{
		listAddressPolicies:    listAddressPolicies,
		generateAddress:        generateAddress,
		allocatePaymentAddress: allocatePaymentAddress,
	}
}

func (c *ChainAddressController) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/v1/chains/", c.handleChainsV1)
}

func (c *ChainAddressController) handleChainsV1(w http.ResponseWriter, r *http.Request) {
	chainRaw, resource, ok := parseChainRoute(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	chain, ok := value_objects.ParseSupportedChain(chainRaw)
	if !ok {
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Error: inport.ErrChainNotSupported.Error()})
		return
	}

	switch resource {
	case "address-policies":
		c.handleListAddressPolicies(w, r, chain)
	case "addresses":
		c.handleGenerateAddress(w, r, chain)
	case "payment-addresses":
		c.handleAllocatePaymentAddress(w, r, chain)
	default:
		http.NotFound(w, r)
	}
}

func (c *ChainAddressController) handleListAddressPolicies(
	w http.ResponseWriter,
	r *http.Request,
	chain value_objects.SupportedChain,
) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, dto.ErrorResponse{Error: "method not allowed"})
		return
	}

	response, err := c.listAddressPolicies.Execute(r.Context(), chain)
	if err != nil {
		if errors.Is(err, inport.ErrChainNotSupported) {
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

type allocatePaymentAddressRequest struct {
	AddressPolicyID     string `json:"addressPolicyId"`
	ExpectedAmountMinor *int64 `json:"expectedAmountMinor"`
	CustomerReference   string `json:"customerReference,omitempty"`
}

func (c *ChainAddressController) handleAllocatePaymentAddress(
	w http.ResponseWriter,
	r *http.Request,
	chain value_objects.SupportedChain,
) {
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
		case errors.Is(err, inport.ErrInvalidExpectedAmount):
			writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func (c *ChainAddressController) handleGenerateAddress(
	w http.ResponseWriter,
	r *http.Request,
	chain value_objects.SupportedChain,
) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, dto.ErrorResponse{Error: "method not allowed"})
		return
	}

	addressPolicyID := strings.TrimSpace(r.URL.Query().Get("addressPolicyId"))
	if addressPolicyID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "addressPolicyId is required"})
		return
	}

	index, err := parseIndexQuery(r.URL.Query().Get("index"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "invalid index"})
		return
	}

	response, err := c.generateAddress.Execute(r.Context(), dto.GenerateAddressInput{
		Chain:           chain,
		AddressPolicyID: addressPolicyID,
		Index:           index,
	})
	if err != nil {
		switch {
		case errors.Is(err, inport.ErrChainNotSupported):
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		case errors.Is(err, inport.ErrAddressPolicyNotFound):
			writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		case errors.Is(err, inport.ErrAddressPolicyNotEnabled):
			writeJSON(w, http.StatusNotImplemented, dto.ErrorResponse{Error: err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func parseChainRoute(path string) (string, string, bool) {
	const prefix = "/v1/chains/"
	if !strings.HasPrefix(path, prefix) {
		return "", "", false
	}

	rest := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	parts := strings.Split(rest, "/")
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

func parseIndexQuery(raw string) (uint32, error) {
	if raw == "" {
		return 0, errors.New("index is required")
	}

	parsed, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return 0, err
	}
	if parsed > maxNonHardenedIndex {
		return 0, errors.New("index exceeds non-hardened range")
	}

	return uint32(parsed), nil
}
