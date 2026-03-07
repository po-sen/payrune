package policy

import (
	"context"
	"testing"

	"payrune/internal/domain/value_objects"
)

func TestAddressPolicyReaderComputesPublicKeyFingerprint(t *testing.T) {
	reader := NewAddressPolicyReader([]AddressPolicyConfig{
		{
			AddressPolicyID:  "policy-a",
			Chain:            value_objects.SupportedChainBitcoin,
			Network:          value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			Scheme:           string(value_objects.BitcoinAddressSchemeLegacy),
			AccountPublicKey: "xpub-a",
		},
		{
			AddressPolicyID:  "policy-b",
			Chain:            value_objects.SupportedChainBitcoin,
			Network:          value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			Scheme:           string(value_objects.BitcoinAddressSchemeLegacy),
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

	if policyA.DerivationConfig.PublicKeyFingerprintAlgo != accountPublicKeyFingerprintAlgorithmSHA256Trunc64HexV1 {
		t.Fatalf("unexpected fingerprint algorithm for policy-a: got %q", policyA.DerivationConfig.PublicKeyFingerprintAlgo)
	}
	if policyB.DerivationConfig.PublicKeyFingerprintAlgo != accountPublicKeyFingerprintAlgorithmSHA256Trunc64HexV1 {
		t.Fatalf("unexpected fingerprint algorithm for policy-b: got %q", policyB.DerivationConfig.PublicKeyFingerprintAlgo)
	}
	if policyA.DerivationConfig.PublicKeyFingerprint == "" {
		t.Fatalf("expected non-empty fingerprint for policy-a")
	}
	if policyB.DerivationConfig.PublicKeyFingerprint == "" {
		t.Fatalf("expected non-empty fingerprint for policy-b")
	}
	if policyA.DerivationConfig.PublicKeyFingerprint == policyB.DerivationConfig.PublicKeyFingerprint {
		t.Fatalf("expected different fingerprints for different account public keys")
	}
}

func TestAddressPolicyReaderUsesConfiguredDerivationPathPrefix(t *testing.T) {
	reader := NewAddressPolicyReader([]AddressPolicyConfig{
		{
			AddressPolicyID:      "native-mainnet",
			Chain:                value_objects.SupportedChainBitcoin,
			Network:              value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			Scheme:               string(value_objects.BitcoinAddressSchemeNativeSegwit),
			AccountPublicKey:     "xpub-a",
			DerivationPathPrefix: "m/84'/0'/0'",
		},
		{
			AddressPolicyID:      "taproot-testnet4",
			Chain:                value_objects.SupportedChainBitcoin,
			Network:              value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4),
			Scheme:               string(value_objects.BitcoinAddressSchemeTaproot),
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
