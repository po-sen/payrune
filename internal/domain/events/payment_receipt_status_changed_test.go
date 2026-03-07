package events

import (
	"testing"
	"time"

	"payrune/internal/domain/value_objects"
)

func TestNewPaymentReceiptStatusChanged(t *testing.T) {
	now := time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC)

	event, err := NewPaymentReceiptStatusChanged(
		101,
		value_objects.PaymentReceiptStatusWatching,
		value_objects.PaymentReceiptStatusPaidConfirmed,
		1000,
		1000,
		0,
		0,
		now,
	)
	if err != nil {
		t.Fatalf("NewPaymentReceiptStatusChanged returned error: %v", err)
	}
	if event.PaymentAddressID != 101 {
		t.Fatalf("unexpected payment address id: got %d", event.PaymentAddressID)
	}
	if !event.StatusChangedAt.Equal(now) {
		t.Fatalf("unexpected status changed at: got %s want %s", event.StatusChangedAt, now)
	}
}

func TestNewPaymentReceiptStatusChangedValidation(t *testing.T) {
	_, err := NewPaymentReceiptStatusChanged(
		0,
		value_objects.PaymentReceiptStatusWatching,
		value_objects.PaymentReceiptStatusPaidConfirmed,
		0,
		0,
		0,
		0,
		time.Now().UTC(),
	)
	if err == nil {
		t.Fatal("expected validation error")
	}

	_, err = NewPaymentReceiptStatusChanged(
		1,
		value_objects.PaymentReceiptStatusWatching,
		value_objects.PaymentReceiptStatusWatching,
		0,
		0,
		0,
		0,
		time.Now().UTC(),
	)
	if err == nil {
		t.Fatal("expected status change validation error")
	}
}
