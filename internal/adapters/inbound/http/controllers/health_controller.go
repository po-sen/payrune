package controllers

import (
	"encoding/json"
	"net/http"

	inport "payrune/internal/application/ports/inbound"
)

type HealthController struct {
	checkHealth inport.CheckHealthUseCase
}

func NewHealthController(checkHealth inport.CheckHealthUseCase) *HealthController {
	return &HealthController{checkHealth: checkHealth}
}

func (c *HealthController) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", c.handleHealth)
}

func (c *HealthController) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response, err := c.checkHealth.Execute(r.Context())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
