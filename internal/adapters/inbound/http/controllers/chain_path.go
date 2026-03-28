package controllers

import (
	"net/http"
	"strings"

	"payrune/internal/application/dto"
	"payrune/internal/domain/valueobjects"
)

const publicUnsupportedChainMessage = "chain is not supported"

func parseSupportedChainPathValue(
	w http.ResponseWriter,
	r *http.Request,
) (valueobjects.SupportedChain, bool) {
	chainRaw := strings.TrimSpace(r.PathValue("chain"))
	chain, ok := valueobjects.ParseSupportedChain(chainRaw)
	if !ok {
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Error: publicUnsupportedChainMessage})
		return "", false
	}
	return chain, true
}
