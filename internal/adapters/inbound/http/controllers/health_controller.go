package controllers

import (
	"net/http"

	inport "payrune/internal/application/ports/inbound"
)

type HealthController struct {
	checkHealth inport.CheckHealthUseCase
}

func NewHealthController(checkHealth inport.CheckHealthUseCase) *HealthController {
	return &HealthController{checkHealth: checkHealth}
}

func (c *HealthController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	response, err := c.checkHealth.Execute(r.Context())
	if err != nil {
		logMappedControllerError(r, http.StatusInternalServerError, "internal server error", err)
		writeErrorJSON(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, newHealthResponse(response))
}
