package entities

import (
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

	issuancePolicy := AddressIssuancePolicy{
		AddressPolicy: AddressPolicy{
			AddressPolicyID: "policy-a",
			Chain:           valueobjects.SupportedChainBitcoin,
			Network:         valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:          string(valueobjects.BitcoinAddressSchemeNativeSegwit),
		},
		DerivationConfig: valueobjects.AddressDerivationConfig{
			DerivationPathPrefix: "m/84'/0'/0'",
		},
	}

	issued, err := allocation.MarkIssued(issuancePolicy, "bc1qexample", "0/42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issued.DerivationPath != "m/84'/0'/0'/0/42" {
		t.Fatalf("unexpected derivation path: got %q", issued.DerivationPath)
	}
	if issued.Address != "bc1qexample" {
		t.Fatalf("unexpected address: got %q", issued.Address)
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

	issuancePolicy := AddressIssuancePolicy{
		AddressPolicy: AddressPolicy{
			AddressPolicyID: "policy-b",
		},
		DerivationConfig: valueobjects.AddressDerivationConfig{
			DerivationPathPrefix: "m/84'/0'/0'",
		},
	}

	if _, err := allocation.MarkIssued(issuancePolicy, "bc1qexample", "0/42"); err == nil {
		t.Fatalf("expected policy mismatch error")
	}
}

func TestPaymentAddressAllocationMarkDerivationFailed(t *testing.T) {
	allocation, err := NewPaymentAddressAllocation(11, "policy-a", 42, 5000, "order-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	failed, err := allocation.MarkDerivationFailed("derive failed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if failed.Status != valueobjects.PaymentAddressAllocationStatusDerivationFailed {
		t.Fatalf("unexpected status: got %q", failed.Status)
	}
	if failed.FailureReason != "derive failed" {
		t.Fatalf("unexpected failure reason: got %q", failed.FailureReason)
	}
}

func TestPaymentAddressAllocationIssueReceiptTracking(t *testing.T) {
	allocation, err := NewPaymentAddressAllocation(11, "policy-a", 42, 5000, "order-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	issuancePolicy := AddressIssuancePolicy{
		AddressPolicy: AddressPolicy{
			AddressPolicyID: "policy-a",
			Chain:           valueobjects.SupportedChainBitcoin,
			Network:         valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			Scheme:          string(valueobjects.BitcoinAddressSchemeNativeSegwit),
		},
		DerivationConfig: valueobjects.AddressDerivationConfig{
			DerivationPathPrefix: "m/84'/1'/0'",
		},
	}
	issued, err := allocation.MarkIssued(issuancePolicy, "tb1qexample", "0/42")
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
