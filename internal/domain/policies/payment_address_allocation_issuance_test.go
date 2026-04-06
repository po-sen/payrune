package policies

import (
	"errors"
	"testing"
	"time"

	"payrune/internal/domain/valueobjects"
)

func TestPaymentAddressAllocationIssuancePolicyPlanUsesDefaults(t *testing.T) {
	policy := NewPaymentAddressAllocationIssuancePolicy(nil, nil)
	issuedAt := time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC)

	plan, err := policy.Plan(
		AddressIssuancePolicy{
			AddressPolicyID: "bitcoin-mainnet-native-segwit",
			Chain:           valueobjects.SupportedChainBitcoin,
			Network:         valueobjects.NetworkIDMainnet,
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef:   "xpub-main",
				IssuanceRefPrefix: "m/84'/0'/0'",
			},
		},
		valueobjects.SupportedChainBitcoin,
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
		map[PaymentReceiptTermsScope]int32{
			{
				Chain:   valueobjects.SupportedChainBitcoin,
				Network: valueobjects.NetworkIDMainnet,
			}: 6,
		},
		map[PaymentReceiptTermsScope]time.Duration{
			{
				Chain:   valueobjects.SupportedChainBitcoin,
				Network: valueobjects.NetworkIDMainnet,
			}: 48 * time.Hour,
		},
	)
	issuedAt := time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC)

	plan, err := policy.Plan(
		AddressIssuancePolicy{
			AddressPolicyID: "bitcoin-mainnet-native-segwit",
			Chain:           valueobjects.SupportedChainBitcoin,
			Network:         valueobjects.NetworkIDMainnet,
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef:   "xpub-main",
				IssuanceRefPrefix: "m/84'/0'/0'",
			},
		},
		valueobjects.SupportedChainBitcoin,
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

func TestPaymentAddressAllocationIssuancePolicyPlanScopesOverridesByChainAndNetwork(t *testing.T) {
	policy := NewPaymentAddressAllocationIssuancePolicy(
		map[PaymentReceiptTermsScope]int32{
			{
				Chain:   valueobjects.SupportedChainBitcoin,
				Network: valueobjects.NetworkIDMainnet,
			}: 6,
			{
				Chain:   valueobjects.SupportedChainEthereum,
				Network: valueobjects.NetworkIDMainnet,
			}: 12,
		},
		map[PaymentReceiptTermsScope]time.Duration{
			{
				Chain:   valueobjects.SupportedChainBitcoin,
				Network: valueobjects.NetworkIDMainnet,
			}: 48 * time.Hour,
			{
				Chain:   valueobjects.SupportedChainEthereum,
				Network: valueobjects.NetworkIDMainnet,
			}: 72 * time.Hour,
		},
	)
	issuedAt := time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC)

	plan, err := policy.Plan(
		AddressIssuancePolicy{
			AddressPolicyID: "ethereum-mainnet-create2",
			Chain:           valueobjects.SupportedChainEthereum,
			Network:         valueobjects.NetworkIDMainnet,
			Scheme:          "create2",
			Decimals:        18,
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef:   "create2.v1:factory=0x1111111111111111111111111111111111111111;collector=0x2222222222222222222222222222222222222222;init_code_hash=0x3333333333333333333333333333333333333333333333333333333333333333",
				IssuanceRefPrefix: "ethereum-mainnet-create2",
			},
		},
		valueobjects.SupportedChainEthereum,
		1200,
		"order-001",
		issuedAt,
	)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if plan.ReceiptTerms.RequiredConfirmations != 12 {
		t.Fatalf("unexpected required confirmations: got %d", plan.ReceiptTerms.RequiredConfirmations)
	}
	expectedExpiresAt := issuedAt.Add(72 * time.Hour)
	if !plan.ReceiptTerms.ExpiresAt.Equal(expectedExpiresAt) {
		t.Fatalf("unexpected expires at: got %s want %s", plan.ReceiptTerms.ExpiresAt, expectedExpiresAt)
	}
}

func TestPaymentAddressAllocationIssuancePolicyPlanRejectsDisabledPolicy(t *testing.T) {
	policy := NewPaymentAddressAllocationIssuancePolicy(nil, nil)

	_, err := policy.Plan(
		AddressIssuancePolicy{
			AddressPolicyID: "bitcoin-mainnet-native-segwit",
			Chain:           valueobjects.SupportedChainBitcoin,
			Network:         valueobjects.NetworkIDMainnet,
			IssuanceConfig:  valueobjects.AddressIssuanceConfig{},
		},
		valueobjects.SupportedChainBitcoin,
		1200,
		"order-001",
		time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC),
	)
	if !errors.Is(err, ErrAddressPolicyNotEnabled) {
		t.Fatalf("unexpected error: got %v", err)
	}
}

func TestPaymentAddressAllocationIssuancePolicyPlanRejectsMissingIssuedAt(t *testing.T) {
	policy := NewPaymentAddressAllocationIssuancePolicy(nil, nil)

	_, err := policy.Plan(
		AddressIssuancePolicy{
			AddressPolicyID: "bitcoin-mainnet-native-segwit",
			Chain:           valueobjects.SupportedChainBitcoin,
			Network:         valueobjects.NetworkIDMainnet,
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef:   "xpub-main",
				IssuanceRefPrefix: "m/84'/0'/0'",
			},
		},
		valueobjects.SupportedChainBitcoin,
		1200,
		"order-001",
		time.Time{},
	)
	if !errors.Is(err, ErrPaymentAddressAllocationIssuedAtRequired) {
		t.Fatalf("unexpected error: got %v", err)
	}
}
