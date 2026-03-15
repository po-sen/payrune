package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"payrune/internal/application/dto"
)

type fakeCheckHealthUseCase struct {
	response dto.HealthResponse
	err      error
}

func (f *fakeCheckHealthUseCase) Execute(_ context.Context) (dto.HealthResponse, error) {
	return f.response, f.err
}

func newHealthTestMux(controller *HealthController) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", controller.HandleHealth)
	return mux
}

func TestHealthControllerGetHealth(t *testing.T) {
	controller := NewHealthController(&fakeCheckHealthUseCase{
		response: dto.HealthResponse{Status: "up", Timestamp: "2026-03-03T11:00:00Z"},
	})

	mux := newHealthTestMux(controller)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}

	var body dto.HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.Status != "up" {
		t.Fatalf("unexpected status body: got %s", body.Status)
	}
}

func TestHealthControllerRejectMethod(t *testing.T) {
	controller := NewHealthController(&fakeCheckHealthUseCase{})
	mux := newHealthTestMux(controller)

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}

	if allow := rr.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("unexpected Allow header: got %q", allow)
	}
}
