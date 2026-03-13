package cloudflarepostgres

import (
	"testing"
	"time"
)

func TestScanValuesSupportsTimeValue(t *testing.T) {
	var statusChangedAt time.Time

	err := scanValues([]any{"2026-03-13T12:34:56Z"}, &statusChangedAt)
	if err != nil {
		t.Fatalf("scanValues returned error: %v", err)
	}

	expected := time.Date(2026, 3, 13, 12, 34, 56, 0, time.UTC)
	if !statusChangedAt.Equal(expected) {
		t.Fatalf("unexpected timestamp: got %s want %s", statusChangedAt, expected)
	}
}

func TestScanValuesSupportsTimePointer(t *testing.T) {
	var deliveredAt *time.Time

	err := scanValues([]any{"2026-03-13T12:34:56Z"}, &deliveredAt)
	if err != nil {
		t.Fatalf("scanValues returned error: %v", err)
	}
	if deliveredAt == nil {
		t.Fatal("expected deliveredAt to be set")
	}

	expected := time.Date(2026, 3, 13, 12, 34, 56, 0, time.UTC)
	if !deliveredAt.Equal(expected) {
		t.Fatalf("unexpected deliveredAt: got %s want %s", deliveredAt, expected)
	}
}

func TestScanValuesSupportsNilTimePointer(t *testing.T) {
	deliveredAt := time.Now().UTC()

	err := scanValues([]any{nil}, &deliveredAt)
	if err != nil {
		t.Fatalf("scanValues returned error: %v", err)
	}
	if !deliveredAt.IsZero() {
		t.Fatalf("expected zero time, got %s", deliveredAt)
	}

	var nextAttemptAt *time.Time
	err = scanValues([]any{nil}, &nextAttemptAt)
	if err != nil {
		t.Fatalf("scanValues returned error: %v", err)
	}
	if nextAttemptAt != nil {
		t.Fatalf("expected nil time pointer, got %v", nextAttemptAt)
	}
}
