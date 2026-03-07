package policies

import (
	"testing"
	"time"

	"payrune/internal/domain/value_objects"
)

func TestMarkPaymentReceiptStatusNotificationSent(t *testing.T) {
	now := time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC)

	result, err := MarkPaymentReceiptStatusNotificationSent(99, now)
	if err != nil {
		t.Fatalf("MarkPaymentReceiptStatusNotificationSent returned error: %v", err)
	}
	if result.Status != value_objects.PaymentReceiptNotificationDeliveryStatusSent {
		t.Fatalf("unexpected status: got %q", result.Status)
	}
	if result.DeliveredAt == nil || !result.DeliveredAt.Equal(now) {
		t.Fatalf("unexpected delivered at: got %v", result.DeliveredAt)
	}
}

func TestResolvePaymentReceiptStatusNotificationDeliveryFailureRetry(t *testing.T) {
	now := time.Date(2026, 3, 7, 12, 5, 0, 0, time.UTC)

	result, err := ResolvePaymentReceiptStatusNotificationDeliveryFailure(
		99,
		1,
		5,
		now,
		2*time.Minute,
		"timeout",
	)
	if err != nil {
		t.Fatalf("ResolvePaymentReceiptStatusNotificationDeliveryFailure returned error: %v", err)
	}
	if result.Status != value_objects.PaymentReceiptNotificationDeliveryStatusPending {
		t.Fatalf("unexpected status: got %q", result.Status)
	}
	if result.Attempts != 2 {
		t.Fatalf("unexpected attempts: got %d", result.Attempts)
	}
	if result.NextAttemptAt == nil || !result.NextAttemptAt.Equal(now.Add(2*time.Minute)) {
		t.Fatalf("unexpected next attempt at: got %v", result.NextAttemptAt)
	}
}

func TestResolvePaymentReceiptStatusNotificationDeliveryFailureTerminal(t *testing.T) {
	now := time.Date(2026, 3, 7, 12, 10, 0, 0, time.UTC)

	result, err := ResolvePaymentReceiptStatusNotificationDeliveryFailure(
		99,
		2,
		3,
		now,
		2*time.Minute,
		"webhook returned 500",
	)
	if err != nil {
		t.Fatalf("ResolvePaymentReceiptStatusNotificationDeliveryFailure returned error: %v", err)
	}
	if result.Status != value_objects.PaymentReceiptNotificationDeliveryStatusFailed {
		t.Fatalf("unexpected status: got %q", result.Status)
	}
	if result.Attempts != 3 {
		t.Fatalf("unexpected attempts: got %d", result.Attempts)
	}
	if result.NextAttemptAt != nil {
		t.Fatalf("expected nil next attempt at, got %v", result.NextAttemptAt)
	}
}
