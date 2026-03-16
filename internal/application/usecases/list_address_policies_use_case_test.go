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
		AssetCode:       "btc",
		AssetType:       "native",
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

func TestListAddressPoliciesUseCaseIncludesAssetMetadata(t *testing.T) {
	reader := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		func() entities.AddressIssuancePolicy {
			policy := entities.AddressIssuancePolicy{
				AddressPolicy: entities.AddressPolicy{
					AddressPolicyID: "ethereum-mainnet-usdt",
					Chain:           valueobjects.SupportedChainEthereum,
					Network:         valueobjects.NetworkID("mainnet"),
					Scheme:          "create2_forwarder",
					AssetCode:       "usdt",
					AssetType:       "erc20",
					TokenAddress:    "0xdAC17F958D2ee523a2206206994597C13D831ec7",
					MinorUnit:       "microUsdt",
					Decimals:        6,
				},
			}
			return policy.Normalize()
		}(),
	})
	useCase := NewListAddressPoliciesUseCase(reader)

	response, err := useCase.Execute(context.Background(), valueobjects.SupportedChainEthereum)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(response.AddressPolicies) != 1 {
		t.Fatalf("unexpected policy count: got %d", len(response.AddressPolicies))
	}
	if response.AddressPolicies[0].AssetCode != "usdt" {
		t.Fatalf("unexpected asset code: got %q", response.AddressPolicies[0].AssetCode)
	}
	if response.AddressPolicies[0].AssetType != "erc20" {
		t.Fatalf("unexpected asset type: got %q", response.AddressPolicies[0].AssetType)
	}
	if response.AddressPolicies[0].TokenAddress != "0xdAC17F958D2ee523a2206206994597C13D831ec7" {
		t.Fatalf("unexpected token address: got %q", response.AddressPolicies[0].TokenAddress)
	}
}
