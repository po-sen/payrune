package cloudflarepostgres

import (
	"testing"
	"time"
)

func TestScanPaymentReceiptStatusNotificationOutboxMessageSupportsTimeColumns(t *testing.T) {
	statusChangedAt := time.Date(2026, 3, 13, 12, 34, 56, 0, time.UTC)
	nextAttemptAt := statusChangedAt.Add(5 * time.Minute)
	deliveredAt := nextAttemptAt.Add(2 * time.Minute)

	message, err := scanPaymentReceiptStatusNotificationOutboxMessage(valueRow{
		values: []any{
			int64(1),
			int64(2),
			"order-123",
			"watching",
			"paid_unconfirmed",
			int64(2000),
			int64(0),
			int64(2000),
			statusChangedAt.Format(time.RFC3339Nano),
			"pending",
			int32(1),
			nextAttemptAt.Format(time.RFC3339Nano),
			"",
			deliveredAt.Format(time.RFC3339Nano),
		},
	})
	if err != nil {
		t.Fatalf("scanPaymentReceiptStatusNotificationOutboxMessage returned error: %v", err)
	}

	if !message.StatusChangedAt.Equal(statusChangedAt) {
		t.Fatalf("unexpected statusChangedAt: got %s want %s", message.StatusChangedAt, statusChangedAt)
	}
	if !message.NextAttemptAt.Equal(nextAttemptAt) {
		t.Fatalf("unexpected nextAttemptAt: got %s want %s", message.NextAttemptAt, nextAttemptAt)
	}
	if message.DeliveredAt == nil || !message.DeliveredAt.Equal(deliveredAt) {
		t.Fatalf("unexpected deliveredAt: got %v want %s", message.DeliveredAt, deliveredAt)
	}
}
