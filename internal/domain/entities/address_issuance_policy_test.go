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
		DerivationConfig: valueobjects.AddressDerivationConfig{
			AccountPublicKey:         " xpub-main ",
			PublicKeyFingerprintAlgo: " hash160 ",
			PublicKeyFingerprint:     " fingerprint-main ",
			DerivationPathPrefix:     "m/84'/0'/0'",
		},
	}
}

func TestAddressIssuancePolicyNormalize(t *testing.T) {
	policy := newTestAddressIssuancePolicy()

	normalized := policy.Normalize()

	if normalized.AddressPolicy.AddressPolicyID != "bitcoin-mainnet-native-segwit" {
		t.Fatalf("unexpected address policy id: got %q", normalized.AddressPolicy.AddressPolicyID)
	}
	if normalized.DerivationConfig.AccountPublicKey != "xpub-main" {
		t.Fatalf("unexpected account public key: got %q", normalized.DerivationConfig.AccountPublicKey)
	}
	if normalized.DerivationConfig.PublicKeyFingerprintAlgo != "hash160" {
		t.Fatalf("unexpected fingerprint algorithm: got %q", normalized.DerivationConfig.PublicKeyFingerprintAlgo)
	}
	if normalized.DerivationConfig.PublicKeyFingerprint != "fingerprint-main" {
		t.Fatalf("unexpected fingerprint: got %q", normalized.DerivationConfig.PublicKeyFingerprint)
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
		{
			name: "fingerprint missing",
			policy: AddressIssuancePolicy{
				AddressPolicy: AddressPolicy{
					AddressPolicyID: "bitcoin-mainnet-native-segwit",
					Chain:           valueobjects.SupportedChainBitcoin,
				},
				DerivationConfig: valueobjects.AddressDerivationConfig{
					AccountPublicKey:     "xpub-main",
					DerivationPathPrefix: "m/84'/0'/0'",
				},
			},
			chain:   valueobjects.SupportedChainBitcoin,
			amount:  1000,
			wantErr: ErrAddressPolicyFingerprintNotConfigured,
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
