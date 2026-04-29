package controllers

import (
	"net/http"
	"strings"
)

const publicUnsupportedChainMessage = "chain is not supported"

func parseSupportedChainPathValue(
	w http.ResponseWriter,
	r *http.Request,
) (string, bool) {
	chain := strings.ToLower(strings.TrimSpace(r.PathValue("chain")))
	switch chain {
	case "bitcoin", "ethereum":
		return chain, true
	default:
		writeErrorJSON(w, http.StatusNotFound, publicUnsupportedChainMessage)
		return "", false
	}
}
