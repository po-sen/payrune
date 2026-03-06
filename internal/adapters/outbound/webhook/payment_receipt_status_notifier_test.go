package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	outport "payrune/internal/application/ports/out"
)

func TestNewPaymentReceiptStatusWebhookNotifierValidation(t *testing.T) {
	_, err := NewPaymentReceiptStatusWebhookNotifier(PaymentReceiptWebhookNotifierConfig{
		URL:    "http://example.com/webhook",
		Secret: "secret",
	})
	if err == nil {
		t.Fatal("expected https validation error")
	}
}

func TestPaymentReceiptStatusWebhookNotifierNotifyStatusChangedSuccess(t *testing.T) {
	var (
		gotEventHeader     string
		gotVersionHeader   string
		gotNotificationID  string
		gotSignatureHeader string
		gotPayload         paymentReceiptStatusChangedPayload
	)

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotEventHeader = r.Header.Get("X-Payrune-Event")
		gotVersionHeader = r.Header.Get("X-Payrune-Event-Version")
		gotNotificationID = r.Header.Get("X-Payrune-Notification-ID")
		gotSignatureHeader = r.Header.Get("X-Payrune-Signature-256")

		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	notifier, err := NewPaymentReceiptStatusWebhookNotifier(PaymentReceiptWebhookNotifierConfig{
		URL:     server.URL,
		Secret:  "secret-key",
		Timeout: 5 * time.Second,
		Client:  server.Client(),
	})
	if err != nil {
		t.Fatalf("NewPaymentReceiptStatusWebhookNotifier returned error: %v", err)
	}

	input := outport.NotifyPaymentReceiptStatusChangedInput{
		NotificationID:        9,
		PaymentAddressID:      101,
		CustomerReference:     "order-9",
		PreviousStatus:        "watching",
		CurrentStatus:         "paid_confirmed",
		ObservedTotalMinor:    1000,
		ConfirmedTotalMinor:   1000,
		UnconfirmedTotalMinor: 0,
		ConflictTotalMinor:    0,
		StatusChangedAt:       time.Date(2026, 3, 6, 19, 0, 0, 0, time.UTC),
	}
	if err := notifier.NotifyStatusChanged(context.Background(), input); err != nil {
		t.Fatalf("NotifyStatusChanged returned error: %v", err)
	}

	if gotEventHeader != outport.PaymentReceiptStatusChangedEventType {
		t.Fatalf("unexpected event header: got %q", gotEventHeader)
	}
	if gotVersionHeader != "1" {
		t.Fatalf("unexpected version header: got %q", gotVersionHeader)
	}
	if gotNotificationID != "9" {
		t.Fatalf("unexpected notification id header: got %q", gotNotificationID)
	}
	if gotPayload.NotificationID != 9 || gotPayload.PaymentAddressID != 101 {
		t.Fatalf("unexpected payload: %+v", gotPayload)
	}

	payloadBytes, err := json.Marshal(gotPayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	mac := hmac.New(sha256.New, []byte("secret-key"))
	_, _ = mac.Write(payloadBytes)
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if gotSignatureHeader != expectedSignature {
		t.Fatalf("unexpected signature: got %q want %q", gotSignatureHeader, expectedSignature)
	}
}

func TestPaymentReceiptStatusWebhookNotifierNotifyStatusChangedNon2xx(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	notifier, err := NewPaymentReceiptStatusWebhookNotifier(PaymentReceiptWebhookNotifierConfig{
		URL:    server.URL,
		Secret: "secret-key",
		Client: server.Client(),
	})
	if err != nil {
		t.Fatalf("NewPaymentReceiptStatusWebhookNotifier returned error: %v", err)
	}

	err = notifier.NotifyStatusChanged(context.Background(), outport.NotifyPaymentReceiptStatusChangedInput{
		NotificationID:   1,
		PaymentAddressID: 2,
		PreviousStatus:   "watching",
		CurrentStatus:    "paid_confirmed",
		StatusChangedAt:  time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("expected non-2xx error")
	}
}

func TestPaymentReceiptStatusWebhookNotifierNotifyStatusChangedWithInsecureSkipVerify(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	notifier, err := NewPaymentReceiptStatusWebhookNotifier(PaymentReceiptWebhookNotifierConfig{
		URL:                server.URL,
		Secret:             "secret-key",
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatalf("NewPaymentReceiptStatusWebhookNotifier returned error: %v", err)
	}

	err = notifier.NotifyStatusChanged(context.Background(), outport.NotifyPaymentReceiptStatusChangedInput{
		NotificationID:   7,
		PaymentAddressID: 8,
		PreviousStatus:   "watching",
		CurrentStatus:    "paid_confirmed",
		StatusChangedAt:  time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("NotifyStatusChanged returned error: %v", err)
	}
}
