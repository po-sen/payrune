package policy

import (
	"context"
	"testing"

	"payrune/internal/domain/entities"
	"payrune/internal/domain/valueobjects"
)

func TestAddressPolicyReaderPreservesAddressSpaceRef(t *testing.T) {
	reader := NewAddressPolicyReader([]entities.AddressIssuancePolicy{
		{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: "policy-a",
				Chain:           valueobjects.SupportedChainBitcoin,
				Network:         valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
				Scheme:          string(valueobjects.BitcoinAddressSchemeLegacy),
			},
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef: "xpub-a",
			},
		},
		{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: "policy-b",
				Chain:           valueobjects.SupportedChainBitcoin,
				Network:         valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
				Scheme:          string(valueobjects.BitcoinAddressSchemeLegacy),
			},
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef: "xpub-b",
			},
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

	if policyA.IssuanceConfig.AddressSpaceRef != "xpub-a" {
		t.Fatalf("unexpected account public key for policy-a: got %q", policyA.IssuanceConfig.AddressSpaceRef)
	}
	if policyB.IssuanceConfig.AddressSpaceRef != "xpub-b" {
		t.Fatalf("unexpected account public key for policy-b: got %q", policyB.IssuanceConfig.AddressSpaceRef)
	}
	if policyA.IssuanceConfig.AddressSpaceRef == policyB.IssuanceConfig.AddressSpaceRef {
		t.Fatalf("expected different account public keys for different policies")
	}
}

func TestAddressPolicyReaderUsesConfiguredIssuanceRefPrefix(t *testing.T) {
	reader := NewAddressPolicyReader([]entities.AddressIssuancePolicy{
		{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: "native-mainnet",
				Chain:           valueobjects.SupportedChainBitcoin,
				Network:         valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
				Scheme:          string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			},
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef:   "xpub-a",
				IssuanceRefPrefix: "m/84'/0'/0'",
			},
		},
		{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: "taproot-testnet4",
				Chain:           valueobjects.SupportedChainBitcoin,
				Network:         valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
				Scheme:          string(valueobjects.BitcoinAddressSchemeTaproot),
			},
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef:   "xpub-b",
				IssuanceRefPrefix: "m/86'/1'/0'",
			},
		},
	})

	nativeMainnet, ok, err := reader.FindIssuanceByID(context.Background(), "native-mainnet")
	if err != nil {
		t.Fatalf("FindIssuanceByID returned error for native-mainnet: %v", err)
	}
	if !ok {
		t.Fatalf("expected native-mainnet exists")
	}
	if nativeMainnet.IssuanceConfig.IssuanceRefPrefix != "m/84'/0'/0'" {
		t.Fatalf(
			"unexpected derivation path prefix for native-mainnet: got %q",
			nativeMainnet.IssuanceConfig.IssuanceRefPrefix,
		)
	}

	taprootTestnet4, ok, err := reader.FindIssuanceByID(context.Background(), "taproot-testnet4")
	if err != nil {
		t.Fatalf("FindIssuanceByID returned error for taproot-testnet4: %v", err)
	}
	if !ok {
		t.Fatalf("expected taproot-testnet4 exists")
	}
	if taprootTestnet4.IssuanceConfig.IssuanceRefPrefix != "m/86'/1'/0'" {
		t.Fatalf(
			"unexpected derivation path prefix for taproot-testnet4: got %q",
			taprootTestnet4.IssuanceConfig.IssuanceRefPrefix,
		)
	}
}

func TestAddressPolicyReaderPreservesEthereumCreate2Config(t *testing.T) {
	reader := NewAddressPolicyReader([]entities.AddressIssuancePolicy{
		{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: "ethereum-mainnet-create2",
				Chain:           valueobjects.SupportedChainEthereum,
				Network:         valueobjects.NetworkID("mainnet"),
				Scheme:          "create2",
				MinorUnit:       "wei",
				Decimals:        18,
			},
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef:   "create2.v1:factory=0x1111111111111111111111111111111111111111;collector=0x2222222222222222222222222222222222222222;init_code_hash=0x3333333333333333333333333333333333333333333333333333333333333333",
				IssuanceRefPrefix: "ethereum-mainnet-create2/",
			},
		},
		{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: "ethereum-sepolia-create2",
				Chain:           valueobjects.SupportedChainEthereum,
				Network:         valueobjects.NetworkID("sepolia"),
				Scheme:          "create2",
				MinorUnit:       "wei",
				Decimals:        18,
			},
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef:   "create2.v1:factory=0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa;collector=0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb;init_code_hash=0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
				IssuanceRefPrefix: "ethereum-sepolia-create2/",
			},
		},
	})

	policy, ok, err := reader.FindIssuanceByID(context.Background(), "ethereum-mainnet-create2")
	if err != nil {
		t.Fatalf("FindIssuanceByID returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected ethereum-mainnet-create2 exists")
	}
	if policy.AddressPolicy.Chain != valueobjects.SupportedChainEthereum {
		t.Fatalf("unexpected chain: got %q", policy.AddressPolicy.Chain)
	}
	if policy.AddressPolicy.MinorUnit != "wei" {
		t.Fatalf("unexpected minor unit: got %q", policy.AddressPolicy.MinorUnit)
	}
	if policy.AddressPolicy.Decimals != 18 {
		t.Fatalf("unexpected decimals: got %d", policy.AddressPolicy.Decimals)
	}
	if policy.IssuanceConfig.IssuanceRefPrefix != "ethereum-mainnet-create2" {
		t.Fatalf("unexpected address reference prefix: got %q", policy.IssuanceConfig.IssuanceRefPrefix)
	}
	if !policy.IsEnabled() {
		t.Fatal("expected ethereum policy enabled")
	}

	sepoliaPolicy, ok, err := reader.FindIssuanceByID(context.Background(), "ethereum-sepolia-create2")
	if err != nil {
		t.Fatalf("FindIssuanceByID returned error for sepolia policy: %v", err)
	}
	if !ok {
		t.Fatal("expected ethereum-sepolia-create2 exists")
	}
	if sepoliaPolicy.AddressPolicy.Network != valueobjects.NetworkID("sepolia") {
		t.Fatalf("unexpected sepolia network: got %q", sepoliaPolicy.AddressPolicy.Network)
	}
	if sepoliaPolicy.IssuanceConfig.IssuanceRefPrefix != "ethereum-sepolia-create2" {
		t.Fatalf("unexpected sepolia address reference prefix: got %q", sepoliaPolicy.IssuanceConfig.IssuanceRefPrefix)
	}
}
