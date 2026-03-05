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

func TestGenerateAddressUseCaseSuccess(t *testing.T) {
	deriver := &fakePolicyBitcoinAddressDeriver{address: "1BitcoinAddressExample"}
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
	})
	useCase := NewGenerateAddressUseCase(deriver, catalog)

	response, err := useCase.Execute(context.Background(), dto.GenerateAddressInput{
		Chain:           value_objects.ChainBitcoin,
		AddressPolicyID: "bitcoin-mainnet-legacy",
		Index:           9,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if response.AddressPolicyID != "bitcoin-mainnet-legacy" {
		t.Fatalf("unexpected address policy id: got %q", response.AddressPolicyID)
	}
	if response.Address != "1BitcoinAddressExample" {
		t.Fatalf("unexpected address: got %q", response.Address)
	}
	if response.MinorUnit != "satoshi" {
		t.Fatalf("unexpected minor unit: got %q", response.MinorUnit)
	}
	if response.Decimals != 8 {
		t.Fatalf("unexpected decimals: got %d", response.Decimals)
	}
	if deriver.lastNetwork != value_objects.BitcoinNetworkMainnet {
		t.Fatalf("unexpected network: got %q", deriver.lastNetwork)
	}
	if deriver.lastScheme != value_objects.BitcoinAddressSchemeLegacy {
		t.Fatalf("unexpected scheme: got %q", deriver.lastScheme)
	}
	if deriver.lastXPub != "xpub-main" {
		t.Fatalf("unexpected xpub: got %q", deriver.lastXPub)
	}
	if deriver.lastIndex != 9 {
		t.Fatalf("unexpected index: got %d", deriver.lastIndex)
	}
}

func TestGenerateAddressUseCaseRejectUnsupportedChain(t *testing.T) {
	useCase := NewGenerateAddressUseCase(
		&fakePolicyBitcoinAddressDeriver{},
		newInMemoryAddressPolicyReader(nil),
	)

	_, err := useCase.Execute(context.Background(), dto.GenerateAddressInput{
		Chain:           value_objects.Chain("eth"),
		AddressPolicyID: "eth-mainnet",
		Index:           0,
	})
	if !errors.Is(err, inport.ErrChainNotSupported) {
		t.Fatalf("expected chain not supported error, got %v", err)
	}
}

func TestGenerateAddressUseCaseRejectUnknownPolicy(t *testing.T) {
	useCase := NewGenerateAddressUseCase(
		&fakePolicyBitcoinAddressDeriver{},
		newInMemoryAddressPolicyReader(nil),
	)

	_, err := useCase.Execute(context.Background(), dto.GenerateAddressInput{
		Chain:           value_objects.ChainBitcoin,
		AddressPolicyID: "bitcoin-mainnet-legacy",
		Index:           0,
	})
	if !errors.Is(err, inport.ErrAddressPolicyNotFound) {
		t.Fatalf("expected address policy not found error, got %v", err)
	}
}

func TestGenerateAddressUseCaseRejectDisabledPolicy(t *testing.T) {
	catalog := newInMemoryAddressPolicyReader([]entities.AddressPolicy{
		{
			AddressPolicyID: "bitcoin-mainnet-legacy",
			Chain:           value_objects.ChainBitcoin,
			Network:         value_objects.BitcoinNetworkMainnet,
			Scheme:          value_objects.BitcoinAddressSchemeLegacy,
			MinorUnit:       "satoshi",
			Decimals:        8,
			XPub:            "",
		},
	})
	useCase := NewGenerateAddressUseCase(&fakePolicyBitcoinAddressDeriver{}, catalog)

	_, err := useCase.Execute(context.Background(), dto.GenerateAddressInput{
		Chain:           value_objects.ChainBitcoin,
		AddressPolicyID: "bitcoin-mainnet-legacy",
		Index:           0,
	})
	if !errors.Is(err, inport.ErrAddressPolicyNotEnabled) {
		t.Fatalf("expected address policy not enabled error, got %v", err)
	}
}

func TestGenerateAddressUseCaseDerivationError(t *testing.T) {
	expectedErr := errors.New("derive failed")
	catalog := newInMemoryAddressPolicyReader([]entities.AddressPolicy{
		{
			AddressPolicyID: "bitcoin-testnet4-native-segwit",
			Chain:           value_objects.ChainBitcoin,
			Network:         value_objects.BitcoinNetworkTestnet4,
			Scheme:          value_objects.BitcoinAddressSchemeNativeSegwit,
			MinorUnit:       "satoshi",
			Decimals:        8,
			XPub:            "tpub-testnet4",
		},
	})
	useCase := NewGenerateAddressUseCase(&fakePolicyBitcoinAddressDeriver{err: expectedErr}, catalog)

	_, err := useCase.Execute(context.Background(), dto.GenerateAddressInput{
		Chain:           value_objects.ChainBitcoin,
		AddressPolicyID: "bitcoin-testnet4-native-segwit",
		Index:           3,
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestGenerateAddressUseCaseRoutesAllSchemes(t *testing.T) {
	tests := []struct {
		name            string
		addressPolicyID string
		scheme          value_objects.BitcoinAddressScheme
	}{
		{name: "legacy", addressPolicyID: "bitcoin-mainnet-legacy", scheme: value_objects.BitcoinAddressSchemeLegacy},
		{name: "segwit", addressPolicyID: "bitcoin-mainnet-segwit", scheme: value_objects.BitcoinAddressSchemeSegwit},
		{name: "native segwit", addressPolicyID: "bitcoin-mainnet-native-segwit", scheme: value_objects.BitcoinAddressSchemeNativeSegwit},
		{name: "taproot", addressPolicyID: "bitcoin-mainnet-taproot", scheme: value_objects.BitcoinAddressSchemeTaproot},
	}

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
			AddressPolicyID: "bitcoin-mainnet-segwit",
			Chain:           value_objects.ChainBitcoin,
			Network:         value_objects.BitcoinNetworkMainnet,
			Scheme:          value_objects.BitcoinAddressSchemeSegwit,
			MinorUnit:       "satoshi",
			Decimals:        8,
			XPub:            "xpub-main",
		},
		{
			AddressPolicyID: "bitcoin-mainnet-native-segwit",
			Chain:           value_objects.ChainBitcoin,
			Network:         value_objects.BitcoinNetworkMainnet,
			Scheme:          value_objects.BitcoinAddressSchemeNativeSegwit,
			MinorUnit:       "satoshi",
			Decimals:        8,
			XPub:            "xpub-main",
		},
		{
			AddressPolicyID: "bitcoin-mainnet-taproot",
			Chain:           value_objects.ChainBitcoin,
			Network:         value_objects.BitcoinNetworkMainnet,
			Scheme:          value_objects.BitcoinAddressSchemeTaproot,
			MinorUnit:       "satoshi",
			Decimals:        8,
			XPub:            "xpub-main",
		},
	})

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			deriver := &fakePolicyBitcoinAddressDeriver{address: "1BitcoinAddressExample"}
			useCase := NewGenerateAddressUseCase(deriver, catalog)

			_, err := useCase.Execute(context.Background(), dto.GenerateAddressInput{
				Chain:           value_objects.ChainBitcoin,
				AddressPolicyID: tc.addressPolicyID,
				Index:           1,
			})
			if err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}
			if deriver.lastScheme != tc.scheme {
				t.Fatalf("unexpected scheme routed to deriver: got %q, want %q", deriver.lastScheme, tc.scheme)
			}
		})
	}
}
