package policies

import (
	"testing"
	"time"

	"payrune/internal/domain/entities"
	"payrune/internal/domain/value_objects"
)

func TestPaymentReceiptTrackingLifecyclePolicyExpireIfDue(t *testing.T) {
	tracking := newPolicyTestTracking(t)
	expiresAt := time.Date(2026, 3, 7, 9, 0, 0, 0, time.UTC)
	tracking.ExpiresAt = &expiresAt

	expiredTracking, expired, err := NewPaymentReceiptTrackingLifecyclePolicy(0).ExpireIfDue(
		tracking,
		expiresAt.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("ExpireIfDue returned error: %v", err)
	}
	if !expired {
		t.Fatal("expected tracking to expire")
	}
	if expiredTracking.Status != value_objects.PaymentReceiptStatusFailedExpired {
		t.Fatalf("unexpected status: got %q", expiredTracking.Status)
	}
	if expiredTracking.LastError != defaultPaymentReceiptExpiredReason {
		t.Fatalf("unexpected expired reason: got %q", expiredTracking.LastError)
	}
}

func TestPaymentReceiptTrackingLifecyclePolicyApplyObservationUsesDefaultExtension(t *testing.T) {
	tracking := newPolicyTestTracking(t)
	expiresAt := time.Date(2026, 3, 7, 9, 0, 0, 0, time.UTC)
	tracking.ExpiresAt = &expiresAt
	now := time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC)

	updatedTracking, err := NewPaymentReceiptTrackingLifecyclePolicy(0).ApplyObservation(
		tracking,
		value_objects.PaymentReceiptObservation{
			ObservedTotalMinor:    1000,
			ConfirmedTotalMinor:   0,
			UnconfirmedTotalMinor: 1000,
			ConflictTotalMinor:    0,
			LatestBlockHeight:     10,
		},
		now,
	)
	if err != nil {
		t.Fatalf("ApplyObservation returned error: %v", err)
	}
	expectedExpiresAt := now.Add(defaultPaymentReceiptPaidUnconfirmedExpiryExtension)
	if updatedTracking.ExpiresAt == nil || !updatedTracking.ExpiresAt.Equal(expectedExpiresAt) {
		t.Fatalf("unexpected expires at: got %v want %s", updatedTracking.ExpiresAt, expectedExpiresAt)
	}
}

func TestPaymentReceiptTrackingLifecyclePolicyApplyObservationUsesConfiguredExtension(t *testing.T) {
	tracking := newPolicyTestTracking(t)
	expiresAt := time.Date(2026, 3, 7, 9, 0, 0, 0, time.UTC)
	tracking.ExpiresAt = &expiresAt
	now := time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC)

	updatedTracking, err := NewPaymentReceiptTrackingLifecyclePolicy(6*time.Hour).ApplyObservation(
		tracking,
		value_objects.PaymentReceiptObservation{
			ObservedTotalMinor:    1000,
			ConfirmedTotalMinor:   0,
			UnconfirmedTotalMinor: 1000,
			ConflictTotalMinor:    0,
			LatestBlockHeight:     10,
		},
		now,
	)
	if err != nil {
		t.Fatalf("ApplyObservation returned error: %v", err)
	}
	expectedExpiresAt := now.Add(6 * time.Hour)
	if updatedTracking.ExpiresAt == nil || !updatedTracking.ExpiresAt.Equal(expectedExpiresAt) {
		t.Fatalf("unexpected expires at: got %v want %s", updatedTracking.ExpiresAt, expectedExpiresAt)
	}
}

func newPolicyTestTracking(t *testing.T) entities.PaymentReceiptTracking {
	t.Helper()

	tracking, err := entities.NewPaymentReceiptTracking(
		1,
		"policy-a",
		value_objects.ChainIDBitcoin,
		value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4),
		"tb1qpolicytest",
		time.Date(2026, 3, 7, 8, 0, 0, 0, time.UTC),
		1000,
		1,
	)
	if err != nil {
		t.Fatalf("NewPaymentReceiptTracking returned error: %v", err)
	}
	return tracking
}
