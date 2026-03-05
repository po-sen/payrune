package use_cases

import (
	"context"
	"errors"
	"testing"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/in"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/value_objects"
)

func TestListAddressPoliciesUseCaseSuccess(t *testing.T) {
	catalog := newInMemoryAddressPolicyReader([]entities.AddressPolicy{
		{
			AddressPolicyID: "bitcoin-mainnet-legacy",
			Chain:           value_objects.ChainBitcoin,
			Network:         value_objects.BitcoinNetworkMainnet,
			Scheme:          value_objects.BitcoinAddressSchemeLegacy,
			MinorUnit:       "satoshi",
			Decimals:        8,
			XPub:            "xpub-main",
		},
		{
			AddressPolicyID: "bitcoin-testnet4-native-segwit",
			Chain:           value_objects.ChainBitcoin,
			Network:         value_objects.BitcoinNetworkTestnet4,
			Scheme:          value_objects.BitcoinAddressSchemeNativeSegwit,
			MinorUnit:       "satoshi",
			Decimals:        8,
			XPub:            "",
		},
	})
	useCase := NewListAddressPoliciesUseCase(catalog)

	response, err := useCase.Execute(context.Background(), value_objects.ChainBitcoin)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if response.Chain != string(value_objects.ChainBitcoin) {
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

func TestListAddressPoliciesUseCaseRejectUnsupportedChain(t *testing.T) {
	useCase := NewListAddressPoliciesUseCase(newInMemoryAddressPolicyReader(nil))

	_, err := useCase.Execute(context.Background(), value_objects.Chain("eth"))
	if !errors.Is(err, inport.ErrChainNotSupported) {
		t.Fatalf("expected chain not supported error, got %v", err)
	}
}
