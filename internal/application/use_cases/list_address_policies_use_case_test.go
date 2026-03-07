package use_cases

import (
	"context"
	"testing"

	"payrune/internal/application/dto"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/value_objects"
)

func TestListAddressPoliciesUseCaseSuccess(t *testing.T) {
	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		newAddressIssuancePolicy(
			"bitcoin-mainnet-legacy",
			value_objects.SupportedChainBitcoin,
			value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			string(value_objects.BitcoinAddressSchemeLegacy),
			"satoshi",
			8,
			"xpub-main",
			testPublicKeyFingerprintAlgo,
			"fingerprint-main-legacy",
			"m/44'/0'/0'",
		),
		newAddressIssuancePolicy(
			"bitcoin-testnet4-native-segwit",
			value_objects.SupportedChainBitcoin,
			value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4),
			string(value_objects.BitcoinAddressSchemeNativeSegwit),
			"satoshi",
			8,
			"",
			"",
			"",
			"",
		),
	})
	useCase := NewListAddressPoliciesUseCase(catalog)

	response, err := useCase.Execute(context.Background(), value_objects.SupportedChainBitcoin)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if response.Chain != string(value_objects.SupportedChainBitcoin) {
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

	response, err := useCase.Execute(context.Background(), value_objects.SupportedChain("eth"))
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

	_, err := useCase.Execute(context.Background(), value_objects.SupportedChainBitcoin)
	if err == nil || err.Error() != "address policy reader is not configured" {
		t.Fatalf("unexpected error: got %v", err)
	}
}
