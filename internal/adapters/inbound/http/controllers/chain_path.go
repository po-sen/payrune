package controllers

import (
	"net/http"
	"strings"

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
		writeErrorJSON(w, http.StatusNotFound, publicUnsupportedChainMessage)
		return "", false
	}
	return chain, true
}
