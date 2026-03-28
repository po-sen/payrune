package controllers

import (
	"net/http"
	"strings"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
	"payrune/internal/domain/valueobjects"
)

func parseSupportedChainPathValue(
	w http.ResponseWriter,
	r *http.Request,
) (valueobjects.SupportedChain, bool) {
	chainRaw := strings.TrimSpace(r.PathValue("chain"))
	chain, ok := valueobjects.ParseSupportedChain(chainRaw)
	if !ok {
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Error: inport.ErrChainNotSupported.Error()})
		return "", false
	}
	return chain, true
}
