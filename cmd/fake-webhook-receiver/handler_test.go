package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFakeWebhookHandlerValidSignature(t *testing.T) {
	body := []byte(`{"eventType":"payment_receipt.status_changed","eventVersion":1,"notificationId":9}`)
	request := httptest.NewRequest(http.MethodPost, "/receipt-status", bytes.NewReader(body))
	request.Header.Set("X-Payrune-Event", "payment_receipt.status_changed")
	request.Header.Set("X-Payrune-Event-Version", "1")
	request.Header.Set("X-Payrune-Notification-ID", "9")
	request.Header.Set(headerSignature256, "sha256="+computeWebhookSignature([]byte("secret-key"), body))

	var logs bytes.Buffer
	handler := newFakeWebhookHandler(log.New(&logs, "", 0), "secret-key")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: got %d", recorder.Code)
	}

	logOutput := logs.String()
	if !strings.Contains(logOutput, "fake webhook request:\n") {
		t.Fatalf("expected request section in logs: %s", logOutput)
	}
	if !strings.Contains(logOutput, "\n  headers:\n") {
		t.Fatalf("expected headers section in logs: %s", logOutput)
	}
	if !strings.Contains(logOutput, `"X-Payrune-Notification-Id": [`) &&
		!strings.Contains(logOutput, `"X-Payrune-Notification-ID": [`) {
		t.Fatalf("expected notification header in logs: %s", logOutput)
	}
	if !strings.Contains(logOutput, "\n  raw_body:\n") {
		t.Fatalf("expected raw body section in logs: %s", logOutput)
	}
	if !strings.Contains(logOutput, `"eventType": "payment_receipt.status_changed"`) {
		t.Fatalf("expected raw body in logs: %s", logOutput)
	}
	if strings.Contains(logOutput, "fake webhook verification:\n") {
		t.Fatalf("did not expect verification section in success logs: %s", logOutput)
	}
	if strings.Contains(logOutput, "fake webhook payload:\n") {
		t.Fatalf("did not expect payload section in success logs: %s", logOutput)
	}
}

func TestFakeWebhookHandlerInvalidSignature(t *testing.T) {
	body := []byte(`{"eventType":"payment_receipt.status_changed","eventVersion":1,"notificationId":9}`)
	request := httptest.NewRequest(http.MethodPost, "/receipt-status", bytes.NewReader(body))
	request.Header.Set("X-Payrune-Event", "payment_receipt.status_changed")
	request.Header.Set("X-Payrune-Event-Version", "1")
	request.Header.Set("X-Payrune-Notification-ID", "9")
	request.Header.Set(headerSignature256, "sha256=invalid")

	var logs bytes.Buffer
	handler := newFakeWebhookHandler(log.New(&logs, "", 0), "secret-key")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: got %d", recorder.Code)
	}

	logOutput := logs.String()
	if !strings.Contains(logOutput, `"eventType": "payment_receipt.status_changed"`) {
		t.Fatalf("expected raw body in logs: %s", logOutput)
	}
	if !strings.Contains(logOutput, "fake webhook signature invalid:\n") {
		t.Fatalf("expected invalid signature section in logs: %s", logOutput)
	}
	if !strings.Contains(logOutput, `provided_signature="sha256=invalid"`) {
		t.Fatalf("expected provided signature in logs: %s", logOutput)
	}
}

func TestFakeWebhookHandlerSkipsVerificationWhenSecretMissing(t *testing.T) {
	body := []byte(`{"eventType":"payment_receipt.status_changed","eventVersion":1,"notificationId":9}`)
	request := httptest.NewRequest(http.MethodPost, "/receipt-status", bytes.NewReader(body))
	request.Header.Set("X-Payrune-Event", "payment_receipt.status_changed")
	request.Header.Set("X-Payrune-Event-Version", "1")
	request.Header.Set("X-Payrune-Notification-ID", "9")

	var logs bytes.Buffer
	handler := newFakeWebhookHandler(log.New(&logs, "", 0), "")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: got %d", recorder.Code)
	}
	if strings.Contains(logs.String(), "fake webhook verification:\n") {
		t.Fatalf("did not expect verification section when secret missing: %s", logs.String())
	}
}

func TestFakeWebhookHandlerRejectsInvalidJSON(t *testing.T) {
	body := []byte(`{"eventType":"payment_receipt.status_changed"`)
	request := httptest.NewRequest(http.MethodPost, "/receipt-status", bytes.NewReader(body))
	request.Header.Set("X-Payrune-Event", "payment_receipt.status_changed")
	request.Header.Set("X-Payrune-Event-Version", "1")
	request.Header.Set("X-Payrune-Notification-ID", "9")
	request.Header.Set(headerSignature256, "sha256="+computeWebhookSignature([]byte("secret-key"), body))

	var logs bytes.Buffer
	handler := newFakeWebhookHandler(log.New(&logs, "", 0), "secret-key")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: got %d", recorder.Code)
	}
	responseBody, err := io.ReadAll(recorder.Result().Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	if !strings.Contains(string(responseBody), "invalid json payload") {
		t.Fatalf("unexpected response body: %q", string(responseBody))
	}
}
