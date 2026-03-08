package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	var called bool
	handler := CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "http://localhost:8081")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatalf("expected downstream handler to be called")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:8081" {
		t.Fatalf("unexpected allow-origin header: got %q", got)
	}
	if !strings.Contains(rr.Header().Get("Access-Control-Allow-Headers"), "Idempotency-Key") {
		t.Fatalf("expected allow-headers to include Idempotency-Key, got %q", rr.Header().Get("Access-Control-Allow-Headers"))
	}
	if !strings.Contains(rr.Header().Get("Access-Control-Expose-Headers"), "Idempotency-Replayed") {
		t.Fatalf("expected expose-headers to include Idempotency-Replayed, got %q", rr.Header().Get("Access-Control-Expose-Headers"))
	}
}

func TestCORSPreflightReturnsNoContent(t *testing.T) {
	var called bool
	handler := CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "http://localhost:8081")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	req.Header.Set("Access-Control-Request-Headers", "Idempotency-Key")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if called {
		t.Fatalf("expected downstream handler not to be called for preflight")
	}
	if rr.Code != http.StatusNoContent {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:8081" {
		t.Fatalf("unexpected allow-origin header: got %q", got)
	}
	if !strings.Contains(rr.Header().Get("Access-Control-Allow-Headers"), "Idempotency-Key") {
		t.Fatalf("expected allow-headers to include Idempotency-Key, got %q", rr.Header().Get("Access-Control-Allow-Headers"))
	}
}

func TestCORSDisallowedOriginOmitted(t *testing.T) {
	handler := CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unexpected allow-origin header: got %q", got)
	}
	if got := rr.Header().Get("Access-Control-Expose-Headers"); got != "" {
		t.Fatalf("unexpected expose-headers: got %q", got)
	}
}
