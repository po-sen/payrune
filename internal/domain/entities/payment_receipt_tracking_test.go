package entities

import (
	"testing"
	"time"

	"payrune/internal/domain/events"
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

func TestPaymentReceiptTrackingExpirationHelpers(t *testing.T) {
	tracking, err := NewPaymentReceiptTracking(
		23,
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

	expiredAt := time.Date(2026, 3, 5, 13, 0, 0, 0, time.UTC)
	tracking.ExpiresAt = &expiredAt

	if !tracking.IsExpired(expiredAt.Add(1 * time.Second)) {
		t.Fatal("expected tracking to be expired")
	}
	if tracking.IsExpired(time.Date(2026, 3, 5, 12, 59, 59, 0, time.UTC)) {
		t.Fatal("did not expect tracking to be expired before deadline")
	}

	expired, err := tracking.MarkExpired("payment window expired")
	if err != nil {
		t.Fatalf("MarkExpired returned error: %v", err)
	}
	if expired.Status != value_objects.PaymentReceiptStatusFailedExpired {
		t.Fatalf("unexpected status: got %q", expired.Status)
	}
	if expired.LastError != "payment window expired" {
		t.Fatalf("unexpected last error: got %q", expired.LastError)
	}

	if _, err := tracking.MarkExpired("   "); err == nil {
		t.Fatal("expected validation error for empty expired reason")
	}
}

func TestPaymentReceiptTrackingExtendExpiryOnTransitionToPaidUnconfirmed(t *testing.T) {
	base, err := NewPaymentReceiptTracking(
		24,
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

	now := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)
	initialExpiresAt := now.Add(30 * time.Minute)

	tests := []struct {
		name              string
		currentStatus     value_objects.PaymentReceiptStatus
		previousStatus    value_objects.PaymentReceiptStatus
		expiresAt         *time.Time
		extension         time.Duration
		expectUnchanged   bool
		expectedExpiresAt time.Time
	}{
		{
			name:              "extend on transition to paid unconfirmed",
			currentStatus:     value_objects.PaymentReceiptStatusPaidUnconfirmed,
			previousStatus:    value_objects.PaymentReceiptStatusPartiallyPaid,
			expiresAt:         &initialExpiresAt,
			extension:         6 * time.Hour,
			expectUnchanged:   false,
			expectedExpiresAt: now.Add(6 * time.Hour),
		},
		{
			name:            "do not extend when status unchanged",
			currentStatus:   value_objects.PaymentReceiptStatusPaidUnconfirmed,
			previousStatus:  value_objects.PaymentReceiptStatusPaidUnconfirmed,
			expiresAt:       &initialExpiresAt,
			extension:       6 * time.Hour,
			expectUnchanged: true,
		},
		{
			name:            "do not extend for non paid-unconfirmed status",
			currentStatus:   value_objects.PaymentReceiptStatusPartiallyPaid,
			previousStatus:  value_objects.PaymentReceiptStatusWatching,
			expiresAt:       &initialExpiresAt,
			extension:       6 * time.Hour,
			expectUnchanged: true,
		},
		{
			name:            "do not shorten existing later expiry",
			currentStatus:   value_objects.PaymentReceiptStatusPaidUnconfirmed,
			previousStatus:  value_objects.PaymentReceiptStatusPartiallyPaid,
			expiresAt:       func() *time.Time { v := now.Add(10 * time.Hour); return &v }(),
			extension:       6 * time.Hour,
			expectUnchanged: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tracking := base
			tracking.Status = tc.currentStatus
			tracking.ExpiresAt = tc.expiresAt

			updated := tracking.ExtendExpiryOnTransitionToPaidUnconfirmed(
				tc.previousStatus,
				now,
				tc.extension,
			)

			if tc.expectUnchanged {
				if updated.ExpiresAt == nil || tracking.ExpiresAt == nil || !updated.ExpiresAt.Equal(*tracking.ExpiresAt) {
					t.Fatalf("expected expiry unchanged, got %v", updated.ExpiresAt)
				}
				return
			}
			if updated.ExpiresAt == nil {
				t.Fatal("expected expiry to be set")
			}
			if !updated.ExpiresAt.Equal(tc.expectedExpiresAt) {
				t.Fatalf("unexpected expiry: got %s, want %s", updated.ExpiresAt, tc.expectedExpiresAt)
			}
		})
	}
}

func TestPaymentReceiptTrackingStatusChangedEvent(t *testing.T) {
	tracking, err := NewPaymentReceiptTracking(
		31,
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
	tracking.Status = value_objects.PaymentReceiptStatusPaidConfirmed
	tracking.ObservedTotalMinor = 1000
	tracking.ConfirmedTotalMinor = 1000

	event, changed, err := tracking.StatusChangedEvent(
		value_objects.PaymentReceiptStatusWatching,
		time.Date(2026, 3, 5, 14, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("StatusChangedEvent returned error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}
	if event.CurrentStatus != value_objects.PaymentReceiptStatusPaidConfirmed {
		t.Fatalf("unexpected current status: got %q", event.CurrentStatus)
	}

	zeroEvent, changed, err := tracking.StatusChangedEvent(
		value_objects.PaymentReceiptStatusPaidConfirmed,
		time.Date(2026, 3, 5, 14, 1, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("StatusChangedEvent returned error on unchanged status: %v", err)
	}
	if changed {
		t.Fatal("expected changed=false")
	}
	if zeroEvent != (events.PaymentReceiptStatusChanged{}) {
		t.Fatalf("expected zero event, got %+v", zeroEvent)
	}
}

func TestPollablePaymentReceiptStatuses(t *testing.T) {
	statuses := PollablePaymentReceiptStatuses()
	if len(statuses) != 4 {
		t.Fatalf("unexpected status count: got %d", len(statuses))
	}
	if statuses[0] != value_objects.PaymentReceiptStatusWatching {
		t.Fatalf("unexpected first status: got %q", statuses[0])
	}
	if statuses[3] != value_objects.PaymentReceiptStatusDoubleSpendSuspected {
		t.Fatalf("unexpected last status: got %q", statuses[3])
	}
}
