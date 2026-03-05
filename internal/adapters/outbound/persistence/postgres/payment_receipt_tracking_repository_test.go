package postgres

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"
)

type stubScanner struct {
	values []any
	err    error
}

func (s stubScanner) Scan(dest ...any) error {
	if s.err != nil {
		return s.err
	}
	if len(dest) != len(s.values) {
		return fmt.Errorf("unexpected scan arg count: got %d want %d", len(dest), len(s.values))
	}

	for i := range dest {
		destValue := reflect.ValueOf(dest[i])
		if destValue.Kind() != reflect.Ptr {
			return fmt.Errorf("dest %d must be pointer", i)
		}
		target := destValue.Elem()

		source := reflect.ValueOf(s.values[i])
		if !source.IsValid() {
			target.Set(reflect.Zero(target.Type()))
			continue
		}

		if source.Type().AssignableTo(target.Type()) {
			target.Set(source)
			continue
		}
		if source.Type().ConvertibleTo(target.Type()) {
			target.Set(source.Convert(target.Type()))
			continue
		}
		return fmt.Errorf("value %d type mismatch: %s -> %s", i, source.Type(), target.Type())
	}

	return nil
}

func TestScanPaymentReceiptTrackingSupportsGenericChainNetwork(t *testing.T) {
	now := time.Date(2026, 3, 5, 15, 0, 0, 0, time.UTC)
	tracking, err := scanPaymentReceiptTracking(stubScanner{
		values: []any{
			int64(1),   // id
			int64(2),   // payment_address_id
			"policy-1", // address_policy_id
			"ethereum", // chain
			"sepolia",  // network
			"0xabc",    // address
			sql.NullTime{Valid: true, Time: now},
			int64(100),   // expected_amount_minor
			int32(2),     // required_confirmations
			"watching",   // receipt_status
			int64(10),    // observed_total_minor
			int64(5),     // confirmed_total_minor
			int64(5),     // unconfirmed_total_minor
			int64(0),     // conflict_total_minor
			int64(12345), // last_observed_block_height
			sql.NullTime{Valid: true, Time: now},
			sql.NullTime{}, // paid_at
			sql.NullTime{}, // confirmed_at
			"",             // last_error
		},
	})
	if err != nil {
		t.Fatalf("scanPaymentReceiptTracking returned error: %v", err)
	}

	if tracking.Chain != "ethereum" {
		t.Fatalf("unexpected chain: got %q", tracking.Chain)
	}
	if tracking.Network != "sepolia" {
		t.Fatalf("unexpected network: got %q", tracking.Network)
	}
	if tracking.FirstObservedAt == nil || !tracking.FirstObservedAt.Equal(now) {
		t.Fatalf("unexpected first observed at: got %+v", tracking.FirstObservedAt)
	}
	if tracking.IssuedAt.IsZero() || !tracking.IssuedAt.Equal(now) {
		t.Fatalf("unexpected issued at: got %s", tracking.IssuedAt)
	}
}

func TestScanPaymentReceiptTrackingRejectsInvalidNetwork(t *testing.T) {
	_, err := scanPaymentReceiptTracking(stubScanner{
		values: []any{
			int64(1),
			int64(2),
			"policy-1",
			"bitcoin",
			"main/net",
			"tb1qabc",
			sql.NullTime{Valid: true, Time: time.Now().UTC()},
			int64(100),
			int32(1),
			"watching",
			int64(0),
			int64(0),
			int64(0),
			int64(0),
			int64(0),
			sql.NullTime{},
			sql.NullTime{},
			sql.NullTime{},
			"",
		},
	})
	if err == nil {
		t.Fatal("expected invalid network error")
	}
}
