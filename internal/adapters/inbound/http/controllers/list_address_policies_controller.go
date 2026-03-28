package controllers

import (
	"errors"
	"net/http"

	"payrune/internal/application/dto"
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
