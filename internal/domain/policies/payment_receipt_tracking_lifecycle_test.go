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

	expiredTracking, expired, err := NewPaymentReceiptTrackingLifecyclePolicy().ExpireIfDue(
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

func TestPaymentReceiptTrackingLifecyclePolicyExpireIfDueSkipsFullyPaid(t *testing.T) {
	tracking := newPolicyTestTracking(t)
	expiresAt := time.Date(2026, 3, 7, 9, 0, 0, 0, time.UTC)
	tracking.ExpiresAt = &expiresAt
	paidAt := time.Date(2026, 3, 7, 8, 30, 0, 0, time.UTC)
	tracking.PaidAt = &paidAt

	updatedTracking, expired, err := NewPaymentReceiptTrackingLifecyclePolicy().ExpireIfDue(
		tracking,
		expiresAt.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("ExpireIfDue returned error: %v", err)
	}
	if expired {
		t.Fatal("did not expect fully paid tracking to expire")
	}
	if updatedTracking.Status != tracking.Status {
		t.Fatalf("unexpected status change: got %q", updatedTracking.Status)
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
