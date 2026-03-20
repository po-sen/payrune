package entities

import (
	"errors"
	"testing"

	"payrune/internal/domain/valueobjects"
)

func newTestAddressIssuancePolicy() AddressIssuancePolicy {
	return AddressIssuancePolicy{
		AddressPolicy: AddressPolicy{
			AddressPolicyID: "bitcoin-mainnet-native-segwit",
			Chain:           valueobjects.SupportedChainBitcoin,
			Network:         valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:          string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			MinorUnit:       "satoshi",
			Decimals:        8,
		},
		IssuanceConfig: valueobjects.AddressIssuanceConfig{
			AddressSourceRef:       " xpub-main ",
			AddressReferencePrefix: "m/84'/0'/0'",
		},
	}
}

func TestAddressIssuancePolicyNormalize(t *testing.T) {
	policy := newTestAddressIssuancePolicy()

	normalized := policy.Normalize()

	if normalized.AddressPolicy.AddressPolicyID != "bitcoin-mainnet-native-segwit" {
		t.Fatalf("unexpected address policy id: got %q", normalized.AddressPolicy.AddressPolicyID)
	}
	if normalized.IssuanceConfig.AddressSourceRef != "xpub-main" {
		t.Fatalf("unexpected account public key: got %q", normalized.IssuanceConfig.AddressSourceRef)
	}
	if !normalized.AddressPolicy.Enabled {
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
				AddressPolicy: AddressPolicy{
					AddressPolicyID: "bitcoin-mainnet-native-segwit",
					Chain:           valueobjects.SupportedChainBitcoin,
				},
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
