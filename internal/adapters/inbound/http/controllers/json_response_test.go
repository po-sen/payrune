package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSONSetsStatusHeadersAndPayload(t *testing.T) {
	recorder := httptest.NewRecorder()

	writeJSON(recorder, http.StatusAccepted, map[string]any{
		"ok":    true,
		"value": "done",
	})

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("unexpected status code: got %d want %d", recorder.Code, http.StatusAccepted)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("unexpected content type: got %q", contentType)
	}

	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if got, ok := body["ok"].(bool); !ok || !got {
		t.Fatalf("unexpected ok field: got %#v", body["ok"])
	}
	if got := body["value"]; got != "done" {
		t.Fatalf("unexpected value field: got %#v", got)
	}
}
