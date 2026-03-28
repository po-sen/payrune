package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
)

const maxNonHardenedIndex = uint64(0x7fffffff)

type GenerateAddressController struct {
	generateAddress inport.GenerateAddressUseCase
}

func NewGenerateAddressController(
	generateAddress inport.GenerateAddressUseCase,
) *GenerateAddressController {
	return &GenerateAddressController{generateAddress: generateAddress}
}

func (c *GenerateAddressController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chain, ok := parseSupportedChainPathValue(w, r)
	if !ok {
		return
	}
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
		case errors.Is(err, inport.ErrAddressPreviewNotSupported):
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response)
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
