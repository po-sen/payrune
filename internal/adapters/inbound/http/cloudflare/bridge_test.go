package cloudflare

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestHandleRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	response, err := HandleRequest(context.Background(), mux, Request{
		Method: http.MethodGet,
		Path:   "/health",
	})
	if err != nil {
		t.Fatalf("HandleRequest returned error: %v", err)
	}
	if response.Status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Status)
	}
	if got := response.Headers["Content-Type"]; got != "application/json" {
		t.Fatalf("expected JSON content type, got %q", got)
	}
	if got := response.Body; got == "" {
		t.Fatalf("expected non-empty body")
	}
}
