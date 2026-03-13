package webhook

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	outport "payrune/internal/application/ports/outbound"
)

type fakeCloudflarePaymentReceiptStatusWebhookBridge struct {
	input CloudflarePaymentReceiptStatusWebhookPostInput
	err   error
}

func (f *fakeCloudflarePaymentReceiptStatusWebhookBridge) PostJSON(
	_ context.Context,
	input CloudflarePaymentReceiptStatusWebhookPostInput,
) error {
	f.input = input
	return f.err
}

func TestNewCloudflarePaymentReceiptStatusWebhookNotifierRejectsInsecureSkipVerify(t *testing.T) {
	_, err := NewCloudflarePaymentReceiptStatusWebhookNotifier(PaymentReceiptWebhookNotifierConfig{
		CloudflareBinding:  "RECEIPT_WEBHOOK_MOCK",
		Secret:             "secret",
		InsecureSkipVerify: true,
	}, &fakeCloudflarePaymentReceiptStatusWebhookBridge{})
	if err == nil {
		t.Fatal("expected insecure skip verify error")
	}
}

func TestCloudflarePaymentReceiptStatusWebhookNotifierNotifyStatusChanged(t *testing.T) {
	bridge := &fakeCloudflarePaymentReceiptStatusWebhookBridge{}
	notifier, err := NewCloudflarePaymentReceiptStatusWebhookNotifier(PaymentReceiptWebhookNotifierConfig{
		CloudflareBinding: "RECEIPT_WEBHOOK_MOCK",
		CloudflarePath:    "/receipt-status",
		Secret:            "top-secret",
		Timeout:           15 * time.Second,
	}, bridge)
	if err != nil {
		t.Fatalf("NewCloudflarePaymentReceiptStatusWebhookNotifier returned error: %v", err)
	}

	err = notifier.NotifyStatusChanged(context.Background(), outport.NotifyPaymentReceiptStatusChangedInput{
		NotificationID:        42,
		PaymentAddressID:      100,
		CustomerReference:     "order-1",
		PreviousStatus:        "watching",
		CurrentStatus:         "paid_unconfirmed",
		ObservedTotalMinor:    2000,
		ConfirmedTotalMinor:   1000,
		UnconfirmedTotalMinor: 1000,
		StatusChangedAt:       time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NotifyStatusChanged returned error: %v", err)
	}

	if bridge.input.Binding != "RECEIPT_WEBHOOK_MOCK" {
		t.Fatalf("unexpected webhook binding: %s", bridge.input.Binding)
	}
	if bridge.input.Path != "/receipt-status" {
		t.Fatalf("unexpected webhook path: %s", bridge.input.Path)
	}
	if bridge.input.Timeout != 15*time.Second {
		t.Fatalf("unexpected timeout: %v", bridge.input.Timeout)
	}
	if got := bridge.input.Headers["Content-Type"]; got != "application/json" {
		t.Fatalf("unexpected content type header: %s", got)
	}
	if got := bridge.input.Headers["X-Payrune-Notification-ID"]; got != "42" {
		t.Fatalf("unexpected notification id header: %s", got)
	}
	if bridge.input.Headers["X-Payrune-Signature-256"] == "" {
		t.Fatal("expected signature header")
	}

	var payload paymentReceiptStatusChangedPayload
	if err := json.Unmarshal(bridge.input.Body, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.NotificationID != 42 || payload.PaymentAddressID != 100 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestCloudflarePaymentReceiptStatusWebhookNotifierNotifyStatusChangedWithBinding(t *testing.T) {
	bridge := &fakeCloudflarePaymentReceiptStatusWebhookBridge{}
	notifier, err := NewCloudflarePaymentReceiptStatusWebhookNotifier(PaymentReceiptWebhookNotifierConfig{
		CloudflareBinding: "RECEIPT_WEBHOOK_MOCK",
		CloudflarePath:    "/receipt-status",
		Secret:            "top-secret",
		Timeout:           15 * time.Second,
	}, bridge)
	if err != nil {
		t.Fatalf("NewCloudflarePaymentReceiptStatusWebhookNotifier returned error: %v", err)
	}

	err = notifier.NotifyStatusChanged(context.Background(), outport.NotifyPaymentReceiptStatusChangedInput{
		NotificationID:        42,
		PaymentAddressID:      100,
		CustomerReference:     "order-1",
		PreviousStatus:        "watching",
		CurrentStatus:         "paid_unconfirmed",
		ObservedTotalMinor:    2000,
		ConfirmedTotalMinor:   1000,
		UnconfirmedTotalMinor: 1000,
		StatusChangedAt:       time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NotifyStatusChanged returned error: %v", err)
	}

	if bridge.input.Binding != "RECEIPT_WEBHOOK_MOCK" {
		t.Fatalf("unexpected webhook binding: %s", bridge.input.Binding)
	}
	if bridge.input.Path != "/receipt-status" {
		t.Fatalf("unexpected webhook binding path: %s", bridge.input.Path)
	}
}
