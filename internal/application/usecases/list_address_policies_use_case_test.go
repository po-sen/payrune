package usecases

import (
	"context"
	"errors"
	"testing"

	inport "payrune/internal/application/ports/inbound"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
)

func TestListAddressPoliciesUseCaseSuccess(t *testing.T) {
	catalog := newInMemoryAddressPolicyReader([]policies.AddressIssuancePolicy{
		newAddressIssuancePolicy(
			"bitcoin-mainnet-legacy",
			valueobjects.SupportedChainBitcoin,
			valueobjects.NetworkIDMainnet,
			string(valueobjects.AddressSchemeLegacy),
			8,
			"xpub-main",
			"m/44'/0'/0'",
		),
		{
			AddressPolicyID: "bitcoin-testnet4-native-segwit",
			Chain:           valueobjects.SupportedChainBitcoin,
			Network:         valueobjects.NetworkIDTestnet4,
			Scheme:          valueobjects.AddressSchemeNativeSegwit,
			Decimals:        8,
			Enabled:         false,
		},
	})
	useCase := NewListAddressPoliciesUseCase(catalog)

	response, err := useCase.Execute(context.Background(), string(valueobjects.SupportedChainBitcoin))
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if response.Chain != string(valueobjects.SupportedChainBitcoin) {
		t.Fatalf("unexpected chain: got %q", response.Chain)
	}
	if len(response.AddressPolicies) != 2 {
		t.Fatalf("unexpected policy count: got %d", len(response.AddressPolicies))
	}

	first := response.AddressPolicies[0]
	if first != (inport.AddressPolicy{
		AddressPolicyID: "bitcoin-mainnet-legacy",
		Chain:           "bitcoin",
		Network:         "mainnet",
		Scheme:          "legacy",
		Decimals:        8,
		Enabled:         true,
	}) {
		t.Fatalf("unexpected first policy: %+v", first)
	}

	if response.AddressPolicies[1].Enabled {
		t.Fatalf("expected second policy disabled")
	}
}

func TestListAddressPoliciesUseCaseReturnsEmptyResultForUnconfiguredChain(t *testing.T) {
	useCase := NewListAddressPoliciesUseCase(newInMemoryAddressPolicyReader(nil))

	response, err := useCase.Execute(context.Background(), string(valueobjects.SupportedChainEthereum))
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if response.Chain != "ethereum" {
		t.Fatalf("unexpected chain: got %q", response.Chain)
	}
	if len(response.AddressPolicies) != 0 {
		t.Fatalf("unexpected policy count: got %d", len(response.AddressPolicies))
	}
}

func TestListAddressPoliciesUseCaseValidationMissingPolicyReader(t *testing.T) {
	useCase := &listAddressPoliciesUseCase{}

	_, err := useCase.Execute(context.Background(), string(valueobjects.SupportedChainBitcoin))
	if !errors.Is(err, inport.ErrAddressPolicyReaderNotConfigured) {
		t.Fatalf("unexpected error: got %v", err)
	}
}

func TestListAddressPoliciesUseCaseMapsReaderFailure(t *testing.T) {
	reader := newInMemoryAddressPolicyReader(nil)
	reader.listErr = errors.New("query failed")
	useCase := NewListAddressPoliciesUseCase(reader)

	_, err := useCase.Execute(context.Background(), string(valueobjects.SupportedChainBitcoin))
	if !errors.Is(err, inport.ErrDependencyFailure) {
		t.Fatalf("expected ErrDependencyFailure, got %v", err)
	}
}
