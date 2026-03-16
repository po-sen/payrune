package policy

import (
	"context"
	"testing"

	"payrune/internal/domain/valueobjects"
)

func TestAddressPolicyReaderPreservesAccountPublicKey(t *testing.T) {
	reader := NewAddressPolicyReader([]AddressPolicyConfig{
		{
			AddressPolicyID:  "policy-a",
			Chain:            valueobjects.SupportedChainBitcoin,
			Network:          valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:           string(valueobjects.BitcoinAddressSchemeLegacy),
			AccountPublicKey: "xpub-a",
		},
		{
			AddressPolicyID:  "policy-b",
			Chain:            valueobjects.SupportedChainBitcoin,
			Network:          valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:           string(valueobjects.BitcoinAddressSchemeLegacy),
			AccountPublicKey: "xpub-b",
		},
	})

	policyA, ok, err := reader.FindIssuanceByID(context.Background(), "policy-a")
	if err != nil {
		t.Fatalf("FindIssuanceByID returned error for policy-a: %v", err)
	}
	if !ok {
		t.Fatalf("expected policy-a exists")
	}

	policyB, ok, err := reader.FindIssuanceByID(context.Background(), "policy-b")
	if err != nil {
		t.Fatalf("FindIssuanceByID returned error for policy-b: %v", err)
	}
	if !ok {
		t.Fatalf("expected policy-b exists")
	}

	if policyA.DerivationConfig.AccountPublicKey != "xpub-a" {
		t.Fatalf("unexpected account public key for policy-a: got %q", policyA.DerivationConfig.AccountPublicKey)
	}
	if policyB.DerivationConfig.AccountPublicKey != "xpub-b" {
		t.Fatalf("unexpected account public key for policy-b: got %q", policyB.DerivationConfig.AccountPublicKey)
	}
	if policyA.DerivationConfig.AccountPublicKey == policyB.DerivationConfig.AccountPublicKey {
		t.Fatalf("expected different account public keys for different policies")
	}
}

func TestAddressPolicyReaderUsesConfiguredDerivationPathPrefix(t *testing.T) {
	reader := NewAddressPolicyReader([]AddressPolicyConfig{
		{
			AddressPolicyID:      "native-mainnet",
			Chain:                valueobjects.SupportedChainBitcoin,
			Network:              valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:               string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			AccountPublicKey:     "xpub-a",
			DerivationPathPrefix: "m/84'/0'/0'",
		},
		{
			AddressPolicyID:      "taproot-testnet4",
			Chain:                valueobjects.SupportedChainBitcoin,
			Network:              valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			Scheme:               string(valueobjects.BitcoinAddressSchemeTaproot),
			AccountPublicKey:     "xpub-b",
			DerivationPathPrefix: "m/86'/1'/0'",
		},
	})

	nativeMainnet, ok, err := reader.FindIssuanceByID(context.Background(), "native-mainnet")
	if err != nil {
		t.Fatalf("FindIssuanceByID returned error for native-mainnet: %v", err)
	}
	if !ok {
		t.Fatalf("expected native-mainnet exists")
	}
	if nativeMainnet.DerivationConfig.DerivationPathPrefix != "m/84'/0'/0'" {
		t.Fatalf(
			"unexpected derivation path prefix for native-mainnet: got %q",
			nativeMainnet.DerivationConfig.DerivationPathPrefix,
		)
	}

	taprootTestnet4, ok, err := reader.FindIssuanceByID(context.Background(), "taproot-testnet4")
	if err != nil {
		t.Fatalf("FindIssuanceByID returned error for taproot-testnet4: %v", err)
	}
	if !ok {
		t.Fatalf("expected taproot-testnet4 exists")
	}
	if taprootTestnet4.DerivationConfig.DerivationPathPrefix != "m/86'/1'/0'" {
		t.Fatalf(
			"unexpected derivation path prefix for taproot-testnet4: got %q",
			taprootTestnet4.DerivationConfig.DerivationPathPrefix,
		)
	}
}
