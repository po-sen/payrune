package entities

import (
	"testing"
	"time"

	"payrune/internal/domain/events"
	"payrune/internal/domain/valueobjects"
)

func TestNewPaymentReceiptTrackingSuccess(t *testing.T) {
	tracking, err := NewPaymentReceiptTracking(
		11,
		"bitcoin-testnet4-native-segwit",
		valueobjects.ChainIDBitcoin,
		valueobjects.NetworkID("testnet4"),
		"tb1qexample",
		time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC),
		1200,
		2,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if tracking.Status != valueobjects.PaymentReceiptStatusWatching {
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
		chain     valueobjects.ChainID
		network   valueobjects.NetworkID
		address   string
		issuedAt  time.Time
		expected  int64
		required  int32
	}{
		{name: "invalid payment id", paymentID: 0, chain: valueobjects.ChainIDBitcoin, network: valueobjects.NetworkID("testnet4"), address: "tb1q", issuedAt: time.Now().UTC(), expected: 1, required: 1},
		{name: "invalid chain identifier", paymentID: 1, chain: "eth/mainnet", network: valueobjects.NetworkID("testnet4"), address: "tb1q", issuedAt: time.Now().UTC(), expected: 1, required: 1},
		{name: "missing network", paymentID: 1, chain: valueobjects.ChainIDBitcoin, network: "", address: "tb1q", issuedAt: time.Now().UTC(), expected: 1, required: 1},
		{name: "missing address", paymentID: 1, chain: valueobjects.ChainIDBitcoin, network: valueobjects.NetworkID("testnet4"), address: "", issuedAt: time.Now().UTC(), expected: 1, required: 1},
		{name: "missing issued at", paymentID: 1, chain: valueobjects.ChainIDBitcoin, network: valueobjects.NetworkID("testnet4"), address: "tb1q", issuedAt: time.Time{}, expected: 1, required: 1},
		{name: "invalid expected", paymentID: 1, chain: valueobjects.ChainIDBitcoin, network: valueobjects.NetworkID("testnet4"), address: "tb1q", issuedAt: time.Now().UTC(), expected: 0, required: 1},
		{name: "invalid confirmations", paymentID: 1, chain: valueobjects.ChainIDBitcoin, network: valueobjects.NetworkID("testnet4"), address: "tb1q", issuedAt: time.Now().UTC(), expected: 1, required: 0},
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
		valueobjects.ChainIDBitcoin,
		valueobjects.NetworkID("testnet4"),
		"tb1qexample",
		time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC),
		1000,
		1,
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	now := time.Date(2026, 3, 5, 13, 30, 0, 0, time.UTC)

	partial, err := tracking.ApplyObservation(valueobjects.PaymentReceiptObservation{
		ObservedTotalMinor:    300,
		ConfirmedTotalMinor:   0,
		UnconfirmedTotalMinor: 300,
		LatestBlockHeight:     120,
	}, now)
	if err != nil {
		t.Fatalf("partial apply error: %v", err)
	}
	if partial.Status != valueobjects.PaymentReceiptStatusPartiallyPaid {
		t.Fatalf("unexpected partial status: got %q", partial.Status)
	}
	if partial.FirstObservedAt == nil {
		t.Fatal("expected first observed at")
	}
	if partial.PaidAt != nil {
		t.Fatal("did not expect paid timestamp yet")
	}

	unconfirmed, err := partial.ApplyObservation(valueobjects.PaymentReceiptObservation{
		ObservedTotalMinor:    1100,
		ConfirmedTotalMinor:   200,
		UnconfirmedTotalMinor: 900,
		LatestBlockHeight:     121,
	}, now.Add(1*time.Minute))
	if err != nil {
		t.Fatalf("paid unconfirmed apply error: %v", err)
	}
	if unconfirmed.Status != valueobjects.PaymentReceiptStatusPaidUnconfirmed {
		t.Fatalf("unexpected paid unconfirmed status: got %q", unconfirmed.Status)
	}
	if unconfirmed.PaidAt == nil {
		t.Fatal("expected paid timestamp")
	}
	if unconfirmed.ConfirmedAt != nil {
		t.Fatal("did not expect confirmed timestamp yet")
	}

	reverted, err := unconfirmed.ApplyObservation(valueobjects.PaymentReceiptObservation{
		ObservedTotalMinor:    0,
		ConfirmedTotalMinor:   0,
		UnconfirmedTotalMinor: 0,
		LatestBlockHeight:     122,
	}, now.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("paid unconfirmed reverted apply error: %v", err)
	}
	if reverted.Status != valueobjects.PaymentReceiptStatusPaidUnconfirmedReverted {
		t.Fatalf("unexpected paid unconfirmed reverted status: got %q", reverted.Status)
	}
	if reverted.PaidAt == nil {
		t.Fatal("expected paid timestamp to stay set")
	}

	recovered, err := reverted.ApplyObservation(valueobjects.PaymentReceiptObservation{
		ObservedTotalMinor:    1100,
		ConfirmedTotalMinor:   200,
		UnconfirmedTotalMinor: 900,
		LatestBlockHeight:     123,
	}, now.Add(3*time.Minute))
	if err != nil {
		t.Fatalf("paid unconfirmed recovery apply error: %v", err)
	}
	if recovered.Status != valueobjects.PaymentReceiptStatusPaidUnconfirmed {
		t.Fatalf("unexpected recovered paid unconfirmed status: got %q", recovered.Status)
	}

	confirmed, err := recovered.ApplyObservation(valueobjects.PaymentReceiptObservation{
		ObservedTotalMinor:    1100,
		ConfirmedTotalMinor:   1100,
		UnconfirmedTotalMinor: 0,
		LatestBlockHeight:     124,
	}, now.Add(4*time.Minute))
	if err != nil {
		t.Fatalf("paid confirmed apply error: %v", err)
	}
	if confirmed.Status != valueobjects.PaymentReceiptStatusPaidConfirmed {
		t.Fatalf("unexpected paid confirmed status: got %q", confirmed.Status)
	}
	if confirmed.ConfirmedAt == nil {
		t.Fatal("expected confirmed timestamp")
	}

	stillConfirmed, err := confirmed.ApplyObservation(valueobjects.PaymentReceiptObservation{
		ObservedTotalMinor:    1100,
		ConfirmedTotalMinor:   1100,
		UnconfirmedTotalMinor: 0,
		LatestBlockHeight:     125,
	}, now.Add(5*time.Minute))
	if err != nil {
		t.Fatalf("conflict apply error: %v", err)
	}
	if stillConfirmed.Status != valueobjects.PaymentReceiptStatusPaidConfirmed {
		t.Fatalf("unexpected conflict status handling: got %q", stillConfirmed.Status)
	}
}

func TestPaymentReceiptTrackingApplyObservationValidation(t *testing.T) {
	tracking, err := NewPaymentReceiptTracking(
		21,
		"bitcoin-testnet4-native-segwit",
		valueobjects.ChainIDBitcoin,
		valueobjects.NetworkID("testnet4"),
		"tb1qexample",
		time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC),
		1000,
		1,
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err = tracking.ApplyObservation(valueobjects.PaymentReceiptObservation{
		ObservedTotalMinor:    100,
		ConfirmedTotalMinor:   50,
		UnconfirmedTotalMinor: 30,
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
		valueobjects.ChainIDBitcoin,
		valueobjects.NetworkID("testnet4"),
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
		valueobjects.ChainIDBitcoin,
		valueobjects.NetworkID("testnet4"),
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
	if !tracking.CanExpireByPaymentWindow() {
		t.Fatal("expected unpaid tracking to remain expiry-eligible")
	}

	paidAt := expiredAt.Add(-5 * time.Minute)
	tracking.PaidAt = &paidAt
	if tracking.CanExpireByPaymentWindow() {
		t.Fatal("did not expect fully paid tracking to remain expiry-eligible")
	}
	tracking.PaidAt = nil

	expired, err := tracking.MarkExpired("payment window expired")
	if err != nil {
		t.Fatalf("MarkExpired returned error: %v", err)
	}
	if expired.Status != valueobjects.PaymentReceiptStatusFailedExpired {
		t.Fatalf("unexpected status: got %q", expired.Status)
	}
	if expired.LastError != "payment window expired" {
		t.Fatalf("unexpected last error: got %q", expired.LastError)
	}

	if _, err := tracking.MarkExpired("   "); err == nil {
		t.Fatal("expected validation error for empty expired reason")
	}
}

func TestPaymentReceiptTrackingApplyObservationDoesNotRevertToUnpaidAfterPaidAt(t *testing.T) {
	base, err := NewPaymentReceiptTracking(
		24,
		"bitcoin-testnet4-native-segwit",
		valueobjects.ChainIDBitcoin,
		valueobjects.NetworkID("testnet4"),
		"tb1qexample",
		time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC),
		1000,
		1,
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	paidAt := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)
	base.PaidAt = &paidAt

	tests := []struct {
		name        string
		observation valueobjects.PaymentReceiptObservation
		want        valueobjects.PaymentReceiptStatus
	}{
		{
			name: "zero observed becomes reverted",
			observation: valueobjects.PaymentReceiptObservation{
				ObservedTotalMinor:    0,
				ConfirmedTotalMinor:   0,
				UnconfirmedTotalMinor: 0,
				LatestBlockHeight:     10,
			},
			want: valueobjects.PaymentReceiptStatusPaidUnconfirmedReverted,
		},
		{
			name: "partial observed becomes reverted",
			observation: valueobjects.PaymentReceiptObservation{
				ObservedTotalMinor:    500,
				ConfirmedTotalMinor:   0,
				UnconfirmedTotalMinor: 500,
				LatestBlockHeight:     10,
			},
			want: valueobjects.PaymentReceiptStatusPaidUnconfirmedReverted,
		},
		{
			name: "full unconfirmed becomes paid unconfirmed",
			observation: valueobjects.PaymentReceiptObservation{
				ObservedTotalMinor:    1000,
				ConfirmedTotalMinor:   0,
				UnconfirmedTotalMinor: 1000,
				LatestBlockHeight:     10,
			},
			want: valueobjects.PaymentReceiptStatusPaidUnconfirmed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			updated, err := base.ApplyObservation(tc.observation, paidAt.Add(time.Minute))
			if err != nil {
				t.Fatalf("ApplyObservation returned error: %v", err)
			}
			if updated.Status != tc.want {
				t.Fatalf("unexpected status: got %q want %q", updated.Status, tc.want)
			}
		})
	}
}

func TestPaymentReceiptTrackingStatusChangedEvent(t *testing.T) {
	tracking, err := NewPaymentReceiptTracking(
		31,
		"bitcoin-testnet4-native-segwit",
		valueobjects.ChainIDBitcoin,
		valueobjects.NetworkID("testnet4"),
		"tb1qexample",
		time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC),
		1000,
		1,
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	tracking.Status = valueobjects.PaymentReceiptStatusPaidConfirmed
	tracking.ObservedTotalMinor = 1000
	tracking.ConfirmedTotalMinor = 1000

	event, changed, err := tracking.StatusChangedEvent(
		valueobjects.PaymentReceiptStatusWatching,
		time.Date(2026, 3, 5, 14, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("StatusChangedEvent returned error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}
	if event.CurrentStatus != valueobjects.PaymentReceiptStatusPaidConfirmed {
		t.Fatalf("unexpected current status: got %q", event.CurrentStatus)
	}

	zeroEvent, changed, err := tracking.StatusChangedEvent(
		valueobjects.PaymentReceiptStatusPaidConfirmed,
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
	if statuses[0] != valueobjects.PaymentReceiptStatusWatching {
		t.Fatalf("unexpected first status: got %q", statuses[0])
	}
	if statuses[3] != valueobjects.PaymentReceiptStatusPaidUnconfirmedReverted {
		t.Fatalf("unexpected reverted status position: got %q", statuses[3])
	}
}
