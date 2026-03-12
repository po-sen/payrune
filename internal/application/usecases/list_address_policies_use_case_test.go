package usecases

import (
	"context"
	"testing"

	"payrune/internal/application/dto"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/valueobjects"
)

func TestListAddressPoliciesUseCaseSuccess(t *testing.T) {
	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		newAddressIssuancePolicy(
			"bitcoin-mainnet-legacy",
			valueobjects.SupportedChainBitcoin,
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeLegacy),
			"satoshi",
			8,
			"xpub-main",
			testPublicKeyFingerprintAlgo,
			"fingerprint-main-legacy",
			"m/44'/0'/0'",
		),
		newAddressIssuancePolicy(
			"bitcoin-testnet4-native-segwit",
			valueobjects.SupportedChainBitcoin,
			valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			"satoshi",
			8,
			"",
			"",
			"",
			"",
		),
	})
	useCase := NewListAddressPoliciesUseCase(catalog)

	response, err := useCase.Execute(context.Background(), valueobjects.SupportedChainBitcoin)
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
	if first != (dto.AddressPolicy{
		AddressPolicyID: "bitcoin-mainnet-legacy",
		Chain:           "bitcoin",
		Network:         "mainnet",
		Scheme:          "legacy",
		MinorUnit:       "satoshi",
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

	response, err := useCase.Execute(context.Background(), valueobjects.SupportedChain("eth"))
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if response.Chain != "eth" {
		t.Fatalf("unexpected chain: got %q", response.Chain)
	}
	if len(response.AddressPolicies) != 0 {
		t.Fatalf("unexpected policy count: got %d", len(response.AddressPolicies))
	}
}

func TestListAddressPoliciesUseCaseValidationMissingPolicyReader(t *testing.T) {
	useCase := &listAddressPoliciesUseCase{}

	_, err := useCase.Execute(context.Background(), valueobjects.SupportedChainBitcoin)
	if err == nil || err.Error() != "address policy reader is not configured" {
		t.Fatalf("unexpected error: got %v", err)
	}
}
