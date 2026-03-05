package entities

import (
	"testing"

	"payrune/internal/domain/value_objects"
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
	if allocation.Status != value_objects.PaymentAddressAllocationStatusReserved {
		t.Fatalf("unexpected status: got %q", allocation.Status)
	}
}

func TestPaymentAddressAllocationMarkIssued(t *testing.T) {
	allocation, err := NewPaymentAddressAllocation(11, "policy-a", 42, 5000, "order-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	policy := AddressPolicy{
		AddressPolicyID:      "policy-a",
		Chain:                value_objects.ChainBitcoin,
		Network:              value_objects.BitcoinNetworkMainnet,
		Scheme:               value_objects.BitcoinAddressSchemeNativeSegwit,
		DerivationPathPrefix: "m/84'/0'/0'",
	}

	issued, err := allocation.MarkIssued(policy, "bc1qexample", "0/42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issued.DerivationPath != "m/84'/0'/0'/0/42" {
		t.Fatalf("unexpected derivation path: got %q", issued.DerivationPath)
	}
	if issued.Address != "bc1qexample" {
		t.Fatalf("unexpected address: got %q", issued.Address)
	}
	if issued.Status != value_objects.PaymentAddressAllocationStatusIssued {
		t.Fatalf("unexpected status: got %q", issued.Status)
	}
}

func TestPaymentAddressAllocationMarkIssuedRejectPolicyMismatch(t *testing.T) {
	allocation, err := NewPaymentAddressAllocation(11, "policy-a", 42, 5000, "order-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	policy := AddressPolicy{
		AddressPolicyID:      "policy-b",
		DerivationPathPrefix: "m/84'/0'/0'",
	}

	if _, err := allocation.MarkIssued(policy, "bc1qexample", "0/42"); err == nil {
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
	if failed.Status != value_objects.PaymentAddressAllocationStatusDerivationFailed {
		t.Fatalf("unexpected status: got %q", failed.Status)
	}
	if failed.FailureReason != "derive failed" {
		t.Fatalf("unexpected failure reason: got %q", failed.FailureReason)
	}
}
