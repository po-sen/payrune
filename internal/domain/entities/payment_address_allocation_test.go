package entities

import (
	"errors"
	"testing"
	"time"

	"payrune/internal/domain/valueobjects"
)

func TestNewPaymentAddressAllocation(t *testing.T) {
	allocation, err := NewPaymentAddressAllocation(1, "policy-a", 10, 1000, " order-1 ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allocation.PaymentAddressID != 1 {
		t.Fatalf("unexpected payment address id: got %d", allocation.PaymentAddressID)
	}
	if allocation.AddressPolicyID != "policy-a" {
		t.Fatalf("unexpected address policy id: got %q", allocation.AddressPolicyID)
	}
	if allocation.CustomerReference != "order-1" {
		t.Fatalf("unexpected customer reference: got %q", allocation.CustomerReference)
	}
	if allocation.Status != valueobjects.PaymentAddressAllocationStatusReserved {
		t.Fatalf("unexpected status: got %q", allocation.Status)
	}
}

func TestPaymentAddressAllocationMarkIssued(t *testing.T) {
	allocation, err := NewPaymentAddressAllocation(11, "policy-a", 42, 5000, "order-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	issued, err := allocation.MarkIssued(
		"policy-a",
		valueobjects.SupportedChainBitcoin,
		valueobjects.NetworkIDMainnet,
		valueobjects.AddressSchemeNativeSegwit,
		"bc1qexample",
		`{"material_type":"bitcoin_hd"}`,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issued.Address != "bc1qexample" {
		t.Fatalf("unexpected address: got %q", issued.Address)
	}
	if issued.SweepMaterialJSON != `{"material_type":"bitcoin_hd"}` {
		t.Fatalf("unexpected sweep material: got %q", issued.SweepMaterialJSON)
	}
	if issued.Status != valueobjects.PaymentAddressAllocationStatusIssued {
		t.Fatalf("unexpected status: got %q", issued.Status)
	}
}

func TestPaymentAddressAllocationMarkIssuedRejectPolicyMismatch(t *testing.T) {
	allocation, err := NewPaymentAddressAllocation(11, "policy-a", 42, 5000, "order-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := allocation.MarkIssued(
		"policy-b",
		valueobjects.SupportedChainBitcoin,
		valueobjects.NetworkIDMainnet,
		valueobjects.AddressSchemeNativeSegwit,
		"bc1qexample",
		`{"material_type":"bitcoin_hd"}`,
	); err == nil {
		t.Fatalf("expected policy mismatch error")
	} else if !errors.Is(err, ErrAddressPolicyMismatch) {
		t.Fatalf("unexpected error: got %v", err)
	}
}

func TestPaymentAddressAllocationMarkDerivationFailed(t *testing.T) {
	allocation, err := NewPaymentAddressAllocation(11, "policy-a", 42, 5000, "order-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	failed, err := allocation.MarkDerivationFailed(
		valueobjects.PaymentAddressAllocationDerivationFailureReasonDerivationFailed,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if failed.Status != valueobjects.PaymentAddressAllocationStatusDerivationFailed {
		t.Fatalf("unexpected status: got %q", failed.Status)
	}
	if failed.DerivationFailureReason != valueobjects.PaymentAddressAllocationDerivationFailureReasonDerivationFailed {
		t.Fatalf("unexpected failure reason: got %q", failed.DerivationFailureReason)
	}
	if failed.SweepMaterialJSON != "" {
		t.Fatalf("expected sweep material to be cleared, got %q", failed.SweepMaterialJSON)
	}

	if _, err := allocation.MarkDerivationFailed(""); !errors.Is(err, ErrDerivationFailureReasonRequired) {
		t.Fatalf("unexpected error: got %v", err)
	}
}

func TestPaymentAddressAllocationIssueReceiptTracking(t *testing.T) {
	allocation, err := NewPaymentAddressAllocation(11, "policy-a", 42, 5000, "order-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	issued, err := allocation.MarkIssued(
		"policy-a",
		valueobjects.SupportedChainBitcoin,
		valueobjects.NetworkIDTestnet4,
		valueobjects.AddressSchemeNativeSegwit,
		"tb1qexample",
		`{"material_type":"bitcoin_hd"}`,
	)
	if err != nil {
		t.Fatalf("MarkIssued returned error: %v", err)
	}

	issuedAt := time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC)
	expiresAt := issuedAt.Add(24 * time.Hour)

	tracking, err := issued.IssueReceiptTracking(issuedAt, 2, expiresAt)
	if err != nil {
		t.Fatalf("IssueReceiptTracking returned error: %v", err)
	}
	if tracking.PaymentAddressID != issued.PaymentAddressID {
		t.Fatalf("unexpected payment address id: got %d", tracking.PaymentAddressID)
	}
	if tracking.Status != valueobjects.PaymentReceiptStatusWatching {
		t.Fatalf("unexpected tracking status: got %q", tracking.Status)
	}
	if tracking.ExpiresAt == nil || !tracking.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("unexpected expires at: got %v", tracking.ExpiresAt)
	}
}

func TestPaymentAddressAllocationMarkIssuedRequiresSweepMaterial(t *testing.T) {
	allocation, err := NewPaymentAddressAllocation(11, "policy-a", 42, 5000, "order-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = allocation.MarkIssued(
		"policy-a",
		valueobjects.SupportedChainBitcoin,
		valueobjects.NetworkIDMainnet,
		valueobjects.AddressSchemeNativeSegwit,
		"bc1qexample",
		"   ",
	)
	if !errors.Is(err, ErrSweepMaterialRequired) {
		t.Fatalf("unexpected error: got %v", err)
	}
}
