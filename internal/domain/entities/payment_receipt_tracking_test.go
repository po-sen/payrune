package entities

import (
	"testing"
	"time"

	"payrune/internal/domain/value_objects"
)

func TestNewPaymentReceiptTrackingSuccess(t *testing.T) {
	tracking, err := NewPaymentReceiptTracking(
		11,
		"bitcoin-testnet4-native-segwit",
		value_objects.ChainIDBitcoin,
		value_objects.NetworkID("testnet4"),
		"tb1qexample",
		time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC),
		1200,
		2,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if tracking.Status != value_objects.PaymentReceiptStatusWatching {
		t.Fatalf("unexpected status: got %q", tracking.Status)
	}
	if tracking.RequiredConfirmations != 2 {
		t.Fatalf("unexpected required confirmations: got %d", tracking.RequiredConfirmations)
	}
}

func TestNewPaymentReceiptTrackingValidation(t *testing.T) {
	tests := []struct {
		name      string
		paymentID int64
		chain     value_objects.ChainID
		network   value_objects.NetworkID
		address   string
		issuedAt  time.Time
		expected  int64
		required  int32
	}{
		{name: "invalid payment id", paymentID: 0, chain: value_objects.ChainIDBitcoin, network: value_objects.NetworkID("testnet4"), address: "tb1q", issuedAt: time.Now().UTC(), expected: 1, required: 1},
		{name: "invalid chain identifier", paymentID: 1, chain: "eth/mainnet", network: value_objects.NetworkID("testnet4"), address: "tb1q", issuedAt: time.Now().UTC(), expected: 1, required: 1},
		{name: "missing network", paymentID: 1, chain: value_objects.ChainIDBitcoin, network: "", address: "tb1q", issuedAt: time.Now().UTC(), expected: 1, required: 1},
		{name: "missing address", paymentID: 1, chain: value_objects.ChainIDBitcoin, network: value_objects.NetworkID("testnet4"), address: "", issuedAt: time.Now().UTC(), expected: 1, required: 1},
		{name: "missing issued at", paymentID: 1, chain: value_objects.ChainIDBitcoin, network: value_objects.NetworkID("testnet4"), address: "tb1q", issuedAt: time.Time{}, expected: 1, required: 1},
		{name: "invalid expected", paymentID: 1, chain: value_objects.ChainIDBitcoin, network: value_objects.NetworkID("testnet4"), address: "tb1q", issuedAt: time.Now().UTC(), expected: 0, required: 1},
		{name: "invalid confirmations", paymentID: 1, chain: value_objects.ChainIDBitcoin, network: value_objects.NetworkID("testnet4"), address: "tb1q", issuedAt: time.Now().UTC(), expected: 1, required: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewPaymentReceiptTracking(
				tc.paymentID,
				"policy",
				tc.chain,
				tc.network,
				tc.address,
				tc.issuedAt,
				tc.expected,
				tc.required,
			)
			if err == nil {
				t.Fatal("expected error but got nil")
			}
		})
	}
}

func TestPaymentReceiptTrackingApplyObservationTransitions(t *testing.T) {
	tracking, err := NewPaymentReceiptTracking(
		15,
		"bitcoin-testnet4-native-segwit",
		value_objects.ChainIDBitcoin,
		value_objects.NetworkID("testnet4"),
		"tb1qexample",
		time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC),
		1000,
		1,
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	now := time.Date(2026, 3, 5, 13, 30, 0, 0, time.UTC)

	partial, err := tracking.ApplyObservation(value_objects.PaymentReceiptObservation{
		ObservedTotalMinor:    300,
		ConfirmedTotalMinor:   0,
		UnconfirmedTotalMinor: 300,
		ConflictTotalMinor:    0,
		LatestBlockHeight:     120,
	}, now)
	if err != nil {
		t.Fatalf("partial apply error: %v", err)
	}
	if partial.Status != value_objects.PaymentReceiptStatusPartiallyPaid {
		t.Fatalf("unexpected partial status: got %q", partial.Status)
	}
	if partial.FirstObservedAt == nil {
		t.Fatal("expected first observed at")
	}
	if partial.PaidAt != nil {
		t.Fatal("did not expect paid timestamp yet")
	}

	unconfirmed, err := partial.ApplyObservation(value_objects.PaymentReceiptObservation{
		ObservedTotalMinor:    1100,
		ConfirmedTotalMinor:   200,
		UnconfirmedTotalMinor: 900,
		ConflictTotalMinor:    0,
		LatestBlockHeight:     121,
	}, now.Add(1*time.Minute))
	if err != nil {
		t.Fatalf("paid unconfirmed apply error: %v", err)
	}
	if unconfirmed.Status != value_objects.PaymentReceiptStatusPaidUnconfirmed {
		t.Fatalf("unexpected paid unconfirmed status: got %q", unconfirmed.Status)
	}
	if unconfirmed.PaidAt == nil {
		t.Fatal("expected paid timestamp")
	}
	if unconfirmed.ConfirmedAt != nil {
		t.Fatal("did not expect confirmed timestamp yet")
	}

	confirmed, err := unconfirmed.ApplyObservation(value_objects.PaymentReceiptObservation{
		ObservedTotalMinor:    1100,
		ConfirmedTotalMinor:   1100,
		UnconfirmedTotalMinor: 0,
		ConflictTotalMinor:    0,
		LatestBlockHeight:     122,
	}, now.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("paid confirmed apply error: %v", err)
	}
	if confirmed.Status != value_objects.PaymentReceiptStatusPaidConfirmed {
		t.Fatalf("unexpected paid confirmed status: got %q", confirmed.Status)
	}
	if confirmed.ConfirmedAt == nil {
		t.Fatal("expected confirmed timestamp")
	}

	conflicted, err := confirmed.ApplyObservation(value_objects.PaymentReceiptObservation{
		ObservedTotalMinor:    1100,
		ConfirmedTotalMinor:   800,
		UnconfirmedTotalMinor: 300,
		ConflictTotalMinor:    200,
		LatestBlockHeight:     123,
	}, now.Add(3*time.Minute))
	if err != nil {
		t.Fatalf("conflict apply error: %v", err)
	}
	if conflicted.Status != value_objects.PaymentReceiptStatusDoubleSpendSuspected {
		t.Fatalf("unexpected conflict status: got %q", conflicted.Status)
	}
}

func TestPaymentReceiptTrackingApplyObservationValidation(t *testing.T) {
	tracking, err := NewPaymentReceiptTracking(
		21,
		"bitcoin-testnet4-native-segwit",
		value_objects.ChainIDBitcoin,
		value_objects.NetworkID("testnet4"),
		"tb1qexample",
		time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC),
		1000,
		1,
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err = tracking.ApplyObservation(value_objects.PaymentReceiptObservation{
		ObservedTotalMinor:    100,
		ConfirmedTotalMinor:   50,
		UnconfirmedTotalMinor: 30,
		ConflictTotalMinor:    0,
		LatestBlockHeight:     1,
	}, time.Now())
	if err == nil {
		t.Fatal("expected validation error but got nil")
	}
}

func TestPaymentReceiptTrackingMarkPollingError(t *testing.T) {
	tracking, err := NewPaymentReceiptTracking(
		22,
		"bitcoin-testnet4-native-segwit",
		value_objects.ChainIDBitcoin,
		value_objects.NetworkID("testnet4"),
		"tb1qexample",
		time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC),
		1000,
		1,
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	updated, err := tracking.MarkPollingError("observer timeout")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if updated.LastError != "observer timeout" {
		t.Fatalf("unexpected last error: got %q", updated.LastError)
	}

	if _, err := tracking.MarkPollingError("   "); err == nil {
		t.Fatal("expected validation error but got nil")
	}
}
