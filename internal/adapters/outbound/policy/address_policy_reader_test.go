package policy

import (
	"context"
	"testing"

	outport "payrune/internal/application/ports/outbound"
)

func TestAddressPolicyReaderPreservesAddressSpaceRef(t *testing.T) {
	reader := NewAddressPolicyReader([]outport.AddressIssuancePolicyRecord{
		{
			AddressPolicyID: "policy-a",
			Chain:           outport.SupportedChainBitcoin,
			Network:         outport.NetworkIDMainnet,
			Scheme:          outport.AddressSchemeLegacy,
			Enabled:         true,
			AddressSpaceRef: "xpub-a",
		},
		{
			AddressPolicyID: "policy-b",
			Chain:           outport.SupportedChainBitcoin,
			Network:         outport.NetworkIDMainnet,
			Scheme:          outport.AddressSchemeLegacy,
			Enabled:         true,
			AddressSpaceRef: "xpub-b",
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

	if policyA.AddressSpaceRef != "xpub-a" {
		t.Fatalf("unexpected account public key for policy-a: got %q", policyA.AddressSpaceRef)
	}
	if policyB.AddressSpaceRef != "xpub-b" {
		t.Fatalf("unexpected account public key for policy-b: got %q", policyB.AddressSpaceRef)
	}
	if policyA.AddressSpaceRef == policyB.AddressSpaceRef {
		t.Fatalf("expected different account public keys for different policies")
	}
}

func TestAddressPolicyReaderUsesConfiguredIssuanceRefPrefix(t *testing.T) {
	reader := NewAddressPolicyReader([]outport.AddressIssuancePolicyRecord{
		{
			AddressPolicyID:   "native-mainnet",
			Chain:             outport.SupportedChainBitcoin,
			Network:           outport.NetworkIDMainnet,
			Scheme:            outport.AddressSchemeNativeSegwit,
			Enabled:           true,
			AddressSpaceRef:   "xpub-a",
			IssuanceRefPrefix: "m/84'/0'/0'",
		},
		{
			AddressPolicyID:   "taproot-testnet4",
			Chain:             outport.SupportedChainBitcoin,
			Network:           outport.NetworkIDTestnet4,
			Scheme:            outport.AddressSchemeTaproot,
			Enabled:           true,
			AddressSpaceRef:   "xpub-b",
			IssuanceRefPrefix: "m/86'/1'/0'",
		},
	})

	nativeMainnet, ok, err := reader.FindIssuanceByID(context.Background(), "native-mainnet")
	if err != nil {
		t.Fatalf("FindIssuanceByID returned error for native-mainnet: %v", err)
	}
	if !ok {
		t.Fatalf("expected native-mainnet exists")
	}
	if nativeMainnet.IssuanceRefPrefix != "m/84'/0'/0'" {
		t.Fatalf(
			"unexpected derivation path prefix for native-mainnet: got %q",
			nativeMainnet.IssuanceRefPrefix,
		)
	}

	taprootTestnet4, ok, err := reader.FindIssuanceByID(context.Background(), "taproot-testnet4")
	if err != nil {
		t.Fatalf("FindIssuanceByID returned error for taproot-testnet4: %v", err)
	}
	if !ok {
		t.Fatalf("expected taproot-testnet4 exists")
	}
	if taprootTestnet4.IssuanceRefPrefix != "m/86'/1'/0'" {
		t.Fatalf(
			"unexpected derivation path prefix for taproot-testnet4: got %q",
			taprootTestnet4.IssuanceRefPrefix,
		)
	}
}

func TestAddressPolicyReaderPreservesEthereumCreate2Config(t *testing.T) {
	reader := NewAddressPolicyReader([]outport.AddressIssuancePolicyRecord{
		{
			AddressPolicyID:   "ethereum-mainnet-create2",
			Chain:             outport.SupportedChainEthereum,
			Network:           outport.NetworkIDMainnet,
			Scheme:            "create2",
			Decimals:          18,
			Enabled:           true,
			AddressSpaceRef:   "create2.v1:factory=0x1111111111111111111111111111111111111111;collector=0x2222222222222222222222222222222222222222;init_code_hash=0x3333333333333333333333333333333333333333333333333333333333333333",
			IssuanceRefPrefix: "ethereum-mainnet-create2/",
		},
		{
			AddressPolicyID:   "ethereum-sepolia-create2",
			Chain:             outport.SupportedChainEthereum,
			Network:           outport.NetworkIDSepolia,
			Scheme:            "create2",
			Decimals:          18,
			Enabled:           true,
			AddressSpaceRef:   "create2.v1:factory=0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa;collector=0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb;init_code_hash=0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
			IssuanceRefPrefix: "ethereum-sepolia-create2/",
		},
	})

	policy, ok, err := reader.FindIssuanceByID(context.Background(), "ethereum-mainnet-create2")
	if err != nil {
		t.Fatalf("FindIssuanceByID returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected ethereum-mainnet-create2 exists")
	}
	if policy.Chain != outport.SupportedChainEthereum {
		t.Fatalf("unexpected chain: got %q", policy.Chain)
	}
	if policy.Decimals != 18 {
		t.Fatalf("unexpected decimals: got %d", policy.Decimals)
	}
	if policy.IssuanceRefPrefix != "ethereum-mainnet-create2" {
		t.Fatalf("unexpected address reference prefix: got %q", policy.IssuanceRefPrefix)
	}
	if !policy.Enabled {
		t.Fatal("expected ethereum policy enabled")
	}

	sepoliaPolicy, ok, err := reader.FindIssuanceByID(context.Background(), "ethereum-sepolia-create2")
	if err != nil {
		t.Fatalf("FindIssuanceByID returned error for sepolia policy: %v", err)
	}
	if !ok {
		t.Fatal("expected ethereum-sepolia-create2 exists")
	}
	if sepoliaPolicy.Network != outport.NetworkIDSepolia {
		t.Fatalf("unexpected sepolia network: got %q", sepoliaPolicy.Network)
	}
	if sepoliaPolicy.IssuanceRefPrefix != "ethereum-sepolia-create2" {
		t.Fatalf("unexpected sepolia address reference prefix: got %q", sepoliaPolicy.IssuanceRefPrefix)
	}
}
