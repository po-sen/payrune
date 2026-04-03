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
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	addressPolicyID := strings.TrimSpace(r.URL.Query().Get("addressPolicyId"))
	if addressPolicyID == "" {
		writeErrorJSON(w, http.StatusBadRequest, "addressPolicyId is required")
		return
	}

	index, err := parseIndexQuery(r.URL.Query().Get("index"))
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid index")
		return
	}

	response, err := c.generateAddress.Execute(r.Context(), dto.GenerateAddressInput{
		Chain:           chain,
		AddressPolicyID: addressPolicyID,
		Index:           index,
	})
	if err != nil {
		statusCode, message := mapGenerateAddressError(err)
		writeErrorJSON(w, statusCode, message)
		return
	}

	writeJSON(w, http.StatusOK, newGenerateAddressResponse(response))
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

func mapGenerateAddressError(err error) (int, string) {
	switch {
	case errors.Is(err, inport.ErrChainNotSupported):
		return http.StatusNotFound, publicUnsupportedChainMessage
	case errors.Is(err, inport.ErrInvalidAddressPolicyID):
		return http.StatusBadRequest, "addressPolicyId is invalid"
	case errors.Is(err, inport.ErrAddressPolicyNotFound):
		return http.StatusBadRequest, "address policy is not supported"
	case errors.Is(err, inport.ErrAddressPolicyNotEnabled):
		return http.StatusNotImplemented, "address policy is not enabled"
	case errors.Is(err, inport.ErrAddressPreviewNotSupported):
		return http.StatusNotFound, "address preview is not supported for this address policy"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
