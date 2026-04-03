package controllers

import (
	"errors"
	"net/http"

	inport "payrune/internal/application/ports/inbound"
)

type ListAddressPoliciesController struct {
	listAddressPolicies inport.ListAddressPoliciesUseCase
}

func NewListAddressPoliciesController(
	listAddressPolicies inport.ListAddressPoliciesUseCase,
) *ListAddressPoliciesController {
	return &ListAddressPoliciesController{listAddressPolicies: listAddressPolicies}
}

func (c *ListAddressPoliciesController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chain, ok := parseSupportedChainPathValue(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	response, err := c.listAddressPolicies.Execute(r.Context(), chain)
	if err != nil {
		statusCode, message := mapListAddressPoliciesError(err)
		writeErrorJSON(w, statusCode, message)
		return
	}

	writeJSON(w, http.StatusOK, newListAddressPoliciesResponse(response))
}

func mapListAddressPoliciesError(err error) (int, string) {
	if errors.Is(err, inport.ErrChainNotSupported) {
		return http.StatusNotFound, publicUnsupportedChainMessage
	}

	return http.StatusInternalServerError, "internal server error"
}
