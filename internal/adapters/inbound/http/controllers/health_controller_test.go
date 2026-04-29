package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	inport "payrune/internal/application/ports/inbound"
)

type fakeCheckHealthUseCase struct {
	response inport.HealthResponse
	err      error
}

func (f *fakeCheckHealthUseCase) Execute(_ context.Context) (inport.HealthResponse, error) {
	return f.response, f.err
}

func TestHealthControllerGetHealth(t *testing.T) {
	controller := NewHealthController(&fakeCheckHealthUseCase{
		response: inport.HealthResponse{
			Status:    "up",
			Timestamp: time.Date(2026, 3, 3, 11, 0, 0, 0, time.UTC),
		},
	})

	mux := http.NewServeMux()
	mux.Handle("/health", controller)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}

	var body healthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.Status != "up" {
		t.Fatalf("unexpected status body: got %s", body.Status)
	}
	if body.Timestamp != "2026-03-03T11:00:00Z" {
		t.Fatalf("unexpected timestamp body: got %s", body.Timestamp)
	}
}

func TestHealthControllerRejectMethod(t *testing.T) {
	controller := NewHealthController(&fakeCheckHealthUseCase{})
	mux := http.NewServeMux()
	mux.Handle("/health", controller)

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}

	if allow := rr.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("unexpected Allow header: got %q", allow)
	}

	assertErrorResponse(t, rr, http.StatusMethodNotAllowed, "method not allowed")
}

func TestHealthControllerInternalErrorUsesJSONError(t *testing.T) {
	controller := NewHealthController(&fakeCheckHealthUseCase{err: errors.New("boom")})
	mux := http.NewServeMux()
	mux.Handle("/health", controller)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusInternalServerError, "internal server error")
}
