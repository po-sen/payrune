package ethereum

import (
	"context"
	"testing"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

func TestCreate2SaltDeriverDeriveAllocationSaltDeterministically(t *testing.T) {
	deriver := NewCreate2SaltDeriver(map[valueobjects.NetworkID]string{
		valueobjects.NetworkID("mainnet"): "0x1111111111111111111111111111111111111111111111111111111111111111",
	})

	input := outport.DeriveEthereumCreate2SaltInput{
		Network:          valueobjects.NetworkID("mainnet"),
		AddressPolicyID:  "ethereum-mainnet-create2",
		PaymentAddressID: 42,
		DerivationIndex:  7,
	}

	first, err := deriver.DeriveAllocationSalt(context.Background(), input)
	if err != nil {
		t.Fatalf("DeriveAllocationSalt returned error: %v", err)
	}
	second, err := deriver.DeriveAllocationSalt(context.Background(), input)
	if err != nil {
		t.Fatalf("DeriveAllocationSalt returned error on second call: %v", err)
	}

	if first != second {
		t.Fatalf("expected deterministic salt, got %q and %q", first, second)
	}
	if len(first) != 66 {
		t.Fatalf("expected 32-byte hex salt, got %q", first)
	}
}

func TestCreate2SaltDeriverDeriveAllocationSaltVariesByAllocationIdentity(t *testing.T) {
	deriver := NewCreate2SaltDeriver(map[valueobjects.NetworkID]string{
		valueobjects.NetworkID("mainnet"): "0x1111111111111111111111111111111111111111111111111111111111111111",
	})

	first, err := deriver.DeriveAllocationSalt(context.Background(), outport.DeriveEthereumCreate2SaltInput{
		Network:          valueobjects.NetworkID("mainnet"),
		AddressPolicyID:  "ethereum-mainnet-create2",
		PaymentAddressID: 42,
		DerivationIndex:  7,
	})
	if err != nil {
		t.Fatalf("DeriveAllocationSalt returned error: %v", err)
	}
	second, err := deriver.DeriveAllocationSalt(context.Background(), outport.DeriveEthereumCreate2SaltInput{
		Network:          valueobjects.NetworkID("mainnet"),
		AddressPolicyID:  "ethereum-mainnet-create2",
		PaymentAddressID: 43,
		DerivationIndex:  7,
	})
	if err != nil {
		t.Fatalf("DeriveAllocationSalt returned error: %v", err)
	}

	if first == second {
		t.Fatalf("expected different salts for different allocations, got %q", first)
	}
}

func TestCreate2SaltDeriverRequiresConfiguredNetwork(t *testing.T) {
	deriver := NewCreate2SaltDeriver(map[valueobjects.NetworkID]string{
		valueobjects.NetworkID("mainnet"): "0x1111111111111111111111111111111111111111111111111111111111111111",
	})

	_, err := deriver.DeriveAllocationSalt(context.Background(), outport.DeriveEthereumCreate2SaltInput{
		Network:          valueobjects.NetworkID("sepolia"),
		AddressPolicyID:  "ethereum-sepolia-create2",
		PaymentAddressID: 42,
		DerivationIndex:  7,
	})
	if err == nil {
		t.Fatal("expected missing-network error")
	}
}

func TestCreate2SaltDeriverFiltersInvalidSecrets(t *testing.T) {
	deriver := NewCreate2SaltDeriver(map[valueobjects.NetworkID]string{
		valueobjects.NetworkID("mainnet"): "not-hex",
	})

	if deriver.HasNetwork(valueobjects.NetworkID("mainnet")) {
		t.Fatal("expected invalid secret to be ignored")
	}
}
