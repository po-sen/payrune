package policies

import (
	"errors"
	"testing"

	"payrune/internal/domain/valueobjects"
)

func newTestAddressIssuancePolicy() AddressIssuancePolicy {
	return AddressIssuancePolicy{
		AddressPolicyID: " bitcoin-mainnet-native-segwit ",
		Chain:           valueobjects.SupportedChainBitcoin,
		Network:         valueobjects.NetworkID(" MAINNET "),
		Scheme:          " native-segwit ",
		MinorUnit:       " satoshi ",
		Decimals:        8,
		IssuanceConfig: valueobjects.AddressIssuanceConfig{
			AddressSpaceRef:   " xpub-main ",
			IssuanceRefPrefix: "m/84'/0'/0'",
		},
	}
}

func TestAddressIssuancePolicyNormalize(t *testing.T) {
	policy := newTestAddressIssuancePolicy()

	normalized := policy.Normalize()

	if normalized.AddressPolicyID != "bitcoin-mainnet-native-segwit" {
		t.Fatalf("unexpected address policy id: got %q", normalized.AddressPolicyID)
	}
	if normalized.Network != valueobjects.NetworkIDMainnet {
		t.Fatalf("unexpected network: got %q", normalized.Network)
	}
	if normalized.Scheme != "native-segwit" {
		t.Fatalf("unexpected scheme: got %q", normalized.Scheme)
	}
	if normalized.MinorUnit != "satoshi" {
		t.Fatalf("unexpected minor unit: got %q", normalized.MinorUnit)
	}
	if normalized.IssuanceConfig.AddressSpaceRef != "xpub-main" {
		t.Fatalf("unexpected account public key: got %q", normalized.IssuanceConfig.AddressSpaceRef)
	}
	if !normalized.Enabled {
		t.Fatal("expected normalized address policy enabled")
	}
}

func TestAddressIssuancePolicyValidateForAllocationIssuance(t *testing.T) {
	policy := newTestAddressIssuancePolicy()

	validated, err := policy.ValidateForAllocationIssuance(valueobjects.SupportedChainBitcoin, 1000)
	if err != nil {
		t.Fatalf("ValidateForAllocationIssuance returned error: %v", err)
	}
	if !validated.IsEnabled() {
		t.Fatal("expected validated policy enabled")
	}
}

func TestAddressIssuancePolicyValidateForAllocationIssuanceRejectsInvalidInput(t *testing.T) {
	base := newTestAddressIssuancePolicy()

	tests := []struct {
		name    string
		policy  AddressIssuancePolicy
		chain   valueobjects.SupportedChain
		amount  int64
		wantErr error
	}{
		{
			name:    "chain mismatch",
			policy:  base,
			chain:   valueobjects.SupportedChain("eth"),
			amount:  1000,
			wantErr: ErrAddressPolicyChainMismatch,
		},
		{
			name: "not enabled",
			policy: AddressIssuancePolicy{
				AddressPolicyID: "bitcoin-mainnet-native-segwit",
				Chain:           valueobjects.SupportedChainBitcoin,
			},
			chain:   valueobjects.SupportedChainBitcoin,
			amount:  1000,
			wantErr: ErrAddressPolicyNotEnabled,
		},
		{
			name:    "invalid amount",
			policy:  base,
			chain:   valueobjects.SupportedChainBitcoin,
			amount:  0,
			wantErr: ErrExpectedAmountMinorInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.policy.ValidateForAllocationIssuance(tc.chain, tc.amount)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("unexpected error: got %v want %v", err, tc.wantErr)
			}
		})
	}
}

func TestAddressIssuancePolicyValidateForAddressPreview(t *testing.T) {
	policy := newTestAddressIssuancePolicy()

	validated, err := policy.ValidateForAddressPreview(valueobjects.SupportedChainBitcoin)
	if err != nil {
		t.Fatalf("ValidateForAddressPreview returned error: %v", err)
	}
	if !validated.SupportsAddressPreview() {
		t.Fatal("expected validated policy to support preview")
	}
}

func TestAddressIssuancePolicyValidateForAddressPreviewRejectsUnsupportedPolicy(t *testing.T) {
	policy := AddressIssuancePolicy{
		AddressPolicyID: "ethereum-mainnet-create2",
		Chain:           valueobjects.SupportedChainEthereum,
		Network:         valueobjects.NetworkIDMainnet,
		Scheme:          "create2",
		MinorUnit:       "wei",
		Decimals:        18,
		IssuanceConfig: valueobjects.AddressIssuanceConfig{
			AddressSpaceRef:   "configured",
			IssuanceRefPrefix: "ethereum-mainnet-create2",
		},
	}

	_, err := policy.ValidateForAddressPreview(valueobjects.SupportedChainEthereum)
	if !errors.Is(err, ErrAddressPolicyPreviewNotSupported) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrAddressPolicyPreviewNotSupported)
	}
	if policy.SupportsAddressPreview() {
		t.Fatal("expected create2 policy preview to be unsupported")
	}
}
