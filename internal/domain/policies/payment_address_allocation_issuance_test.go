package policies

import (
	"errors"
	"testing"
	"time"

	"payrune/internal/domain/entities"
	"payrune/internal/domain/value_objects"
)

func TestPaymentAddressAllocationIssuancePolicyPlanUsesDefaults(t *testing.T) {
	policy := NewPaymentAddressAllocationIssuancePolicy(nil, nil)
	issuedAt := time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC)

	plan, err := policy.Plan(
		entities.AddressIssuancePolicy{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: "bitcoin-mainnet-native-segwit",
				Chain:           value_objects.SupportedChainBitcoin,
				Network:         value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			},
			DerivationConfig: value_objects.AddressDerivationConfig{
				AccountPublicKey:         "xpub-main",
				PublicKeyFingerprintAlgo: "hash160",
				PublicKeyFingerprint:     "fingerprint-main",
				DerivationPathPrefix:     "m/84'/0'/0'",
			},
		},
		value_objects.SupportedChainBitcoin,
		1200,
		" order-001 ",
		issuedAt,
	)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if plan.ReceiptTerms.RequiredConfirmations != defaultPaymentReceiptRequiredConfirmations {
		t.Fatalf("unexpected required confirmations: got %d", plan.ReceiptTerms.RequiredConfirmations)
	}
	if !plan.ReceiptTerms.ExpiresAt.Equal(issuedAt.Add(defaultPaymentReceiptExpiresAfter)) {
		t.Fatalf("unexpected expires at: got %s", plan.ReceiptTerms.ExpiresAt)
	}
	if plan.Reservation.CustomerReference != "order-001" {
		t.Fatalf("unexpected customer reference: got %q", plan.Reservation.CustomerReference)
	}
	if len(plan.ReservationAttempts) != 2 {
		t.Fatalf("unexpected reservation attempts length: got %d", len(plan.ReservationAttempts))
	}
	if plan.ReservationAttempts[0] != PaymentAddressAllocationReservationAttemptReopenFailed {
		t.Fatalf("unexpected first reservation attempt: got %q", plan.ReservationAttempts[0])
	}
	if plan.ReservationAttempts[1] != PaymentAddressAllocationReservationAttemptReserveFresh {
		t.Fatalf("unexpected second reservation attempt: got %q", plan.ReservationAttempts[1])
	}
}

func TestPaymentAddressAllocationIssuancePolicyPlanUsesNetworkOverrides(t *testing.T) {
	policy := NewPaymentAddressAllocationIssuancePolicy(
		map[value_objects.NetworkID]int32{
			value_objects.NetworkID(value_objects.BitcoinNetworkMainnet): 6,
		},
		map[value_objects.NetworkID]time.Duration{
			value_objects.NetworkID(value_objects.BitcoinNetworkMainnet): 48 * time.Hour,
		},
	)
	issuedAt := time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC)

	plan, err := policy.Plan(
		entities.AddressIssuancePolicy{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: "bitcoin-mainnet-native-segwit",
				Chain:           value_objects.SupportedChainBitcoin,
				Network:         value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			},
			DerivationConfig: value_objects.AddressDerivationConfig{
				AccountPublicKey:         "xpub-main",
				PublicKeyFingerprintAlgo: "hash160",
				PublicKeyFingerprint:     "fingerprint-main",
				DerivationPathPrefix:     "m/84'/0'/0'",
			},
		},
		value_objects.SupportedChainBitcoin,
		1200,
		"order-001",
		issuedAt,
	)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if plan.ReceiptTerms.RequiredConfirmations != 6 {
		t.Fatalf("unexpected required confirmations: got %d", plan.ReceiptTerms.RequiredConfirmations)
	}
	expectedExpiresAt := issuedAt.Add(48 * time.Hour)
	if !plan.ReceiptTerms.ExpiresAt.Equal(expectedExpiresAt) {
		t.Fatalf("unexpected expires at: got %s want %s", plan.ReceiptTerms.ExpiresAt, expectedExpiresAt)
	}
}

func TestPaymentAddressAllocationIssuancePolicyPlanRejectsInvalidPolicy(t *testing.T) {
	policy := NewPaymentAddressAllocationIssuancePolicy(nil, nil)

	_, err := policy.Plan(
		entities.AddressIssuancePolicy{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: "bitcoin-mainnet-native-segwit",
				Chain:           value_objects.SupportedChainBitcoin,
				Network:         value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			},
			DerivationConfig: value_objects.AddressDerivationConfig{
				AccountPublicKey:     "xpub-main",
				DerivationPathPrefix: "m/84'/0'/0'",
			},
		},
		value_objects.SupportedChainBitcoin,
		1200,
		"order-001",
		time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC),
	)
	if !errors.Is(err, entities.ErrAddressPolicyFingerprintNotConfigured) {
		t.Fatalf("unexpected error: got %v", err)
	}
}
