package entities

import (
	"testing"

	"payrune/internal/domain/valueobjects"
)

func TestAddressPolicyNormalize(t *testing.T) {
	policy := AddressPolicy{
		AddressPolicyID: " bitcoin-mainnet-native-segwit ",
		Chain:           valueobjects.SupportedChainBitcoin,
		Network:         valueobjects.NetworkID(" MAINNET "),
		Scheme:          " native-segwit ",
		MinorUnit:       " satoshi ",
		Decimals:        8,
		Enabled:         true,
	}

	normalized := policy.Normalize()

	if normalized.AddressPolicyID != "bitcoin-mainnet-native-segwit" {
		t.Fatalf("unexpected address policy id: got %q", normalized.AddressPolicyID)
	}
	if normalized.Network != valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet) {
		t.Fatalf("unexpected network: got %q", normalized.Network)
	}
	if normalized.Scheme != "native-segwit" {
		t.Fatalf("unexpected scheme: got %q", normalized.Scheme)
	}
	if normalized.MinorUnit != "satoshi" {
		t.Fatalf("unexpected minor unit: got %q", normalized.MinorUnit)
	}
	if !normalized.IsEnabled() {
		t.Fatal("expected normalized policy enabled")
	}
}

func TestAddressPolicyIsEnabled(t *testing.T) {
	if (AddressPolicy{Enabled: true}).IsEnabled() != true {
		t.Fatal("expected enabled policy to report enabled")
	}
	if (AddressPolicy{Enabled: false}).IsEnabled() != false {
		t.Fatal("expected disabled policy to report disabled")
	}
}
