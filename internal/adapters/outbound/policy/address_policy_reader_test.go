package policy

import (
	"context"
	"testing"

	"payrune/internal/domain/value_objects"
)

func TestAddressPolicyReaderComputesXPubFingerprint(t *testing.T) {
	reader := NewAddressPolicyReader([]AddressPolicyConfig{
		{
			AddressPolicyID: "policy-a",
			Chain:           value_objects.ChainBitcoin,
			Network:         value_objects.BitcoinNetworkMainnet,
			Scheme:          value_objects.BitcoinAddressSchemeLegacy,
			XPub:            "xpub-a",
		},
		{
			AddressPolicyID: "policy-b",
			Chain:           value_objects.ChainBitcoin,
			Network:         value_objects.BitcoinNetworkMainnet,
			Scheme:          value_objects.BitcoinAddressSchemeLegacy,
			XPub:            "xpub-b",
		},
	})

	policyA, ok, err := reader.FindByID(context.Background(), "policy-a")
	if err != nil {
		t.Fatalf("FindByID returned error for policy-a: %v", err)
	}
	if !ok {
		t.Fatalf("expected policy-a exists")
	}

	policyB, ok, err := reader.FindByID(context.Background(), "policy-b")
	if err != nil {
		t.Fatalf("FindByID returned error for policy-b: %v", err)
	}
	if !ok {
		t.Fatalf("expected policy-b exists")
	}

	if policyA.XPubFingerprintAlgo != xpubFingerprintAlgorithmSHA256Trunc64HexV1 {
		t.Fatalf("unexpected fingerprint algorithm for policy-a: got %q", policyA.XPubFingerprintAlgo)
	}
	if policyB.XPubFingerprintAlgo != xpubFingerprintAlgorithmSHA256Trunc64HexV1 {
		t.Fatalf("unexpected fingerprint algorithm for policy-b: got %q", policyB.XPubFingerprintAlgo)
	}
	if policyA.XPubFingerprint == "" {
		t.Fatalf("expected non-empty fingerprint for policy-a")
	}
	if policyB.XPubFingerprint == "" {
		t.Fatalf("expected non-empty fingerprint for policy-b")
	}
	if policyA.XPubFingerprint == policyB.XPubFingerprint {
		t.Fatalf("expected different fingerprints for different xpub values")
	}
}

func TestAddressPolicyReaderUsesConfiguredDerivationPathPrefix(t *testing.T) {
	reader := NewAddressPolicyReader([]AddressPolicyConfig{
		{
			AddressPolicyID:      "native-mainnet",
			Chain:                value_objects.ChainBitcoin,
			Network:              value_objects.BitcoinNetworkMainnet,
			Scheme:               value_objects.BitcoinAddressSchemeNativeSegwit,
			XPub:                 "xpub-a",
			DerivationPathPrefix: "m/84'/0'/0'",
		},
		{
			AddressPolicyID:      "taproot-testnet4",
			Chain:                value_objects.ChainBitcoin,
			Network:              value_objects.BitcoinNetworkTestnet4,
			Scheme:               value_objects.BitcoinAddressSchemeTaproot,
			XPub:                 "xpub-b",
			DerivationPathPrefix: "m/86'/1'/0'",
		},
	})

	nativeMainnet, ok, err := reader.FindByID(context.Background(), "native-mainnet")
	if err != nil {
		t.Fatalf("FindByID returned error for native-mainnet: %v", err)
	}
	if !ok {
		t.Fatalf("expected native-mainnet exists")
	}
	if nativeMainnet.DerivationPathPrefix != "m/84'/0'/0'" {
		t.Fatalf(
			"unexpected derivation path prefix for native-mainnet: got %q",
			nativeMainnet.DerivationPathPrefix,
		)
	}

	taprootTestnet4, ok, err := reader.FindByID(context.Background(), "taproot-testnet4")
	if err != nil {
		t.Fatalf("FindByID returned error for taproot-testnet4: %v", err)
	}
	if !ok {
		t.Fatalf("expected taproot-testnet4 exists")
	}
	if taprootTestnet4.DerivationPathPrefix != "m/86'/1'/0'" {
		t.Fatalf(
			"unexpected derivation path prefix for taproot-testnet4: got %q",
			taprootTestnet4.DerivationPathPrefix,
		)
	}
}
