package usecases

import (
	"context"
	"errors"
	"testing"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/valueobjects"
)

func TestGenerateAddressUseCaseSuccess(t *testing.T) {
	deriver := newFakeChainAddressDeriver()
	deriver.output = dtoToDeriveOutput("1BitcoinAddressExample", "0/9")
	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		newAddressIssuancePolicy(
			"bitcoin-mainnet-legacy",
			valueobjects.SupportedChainBitcoin,
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeLegacy),
			"satoshi",
			8,
			"xpub-main",
			"m/44'/0'/0'",
		),
	})
	useCase := NewGenerateAddressUseCase(deriver, catalog)

	response, err := useCase.Execute(context.Background(), dto.GenerateAddressInput{
		Chain:           valueobjects.SupportedChainBitcoin,
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
	if deriver.lastInput.Network != valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet) {
		t.Fatalf("unexpected network: got %q", deriver.lastInput.Network)
	}
	if deriver.lastInput.Scheme != string(valueobjects.BitcoinAddressSchemeLegacy) {
		t.Fatalf("unexpected scheme: got %q", deriver.lastInput.Scheme)
	}
	if deriver.lastInput.AddressSpaceRef != "xpub-main" {
		t.Fatalf("unexpected address source ref: got %q", deriver.lastInput.AddressSpaceRef)
	}
	if deriver.lastInput.IssuanceRefPrefix != "m/44'/0'/0'" {
		t.Fatalf("unexpected address reference prefix: got %q", deriver.lastInput.IssuanceRefPrefix)
	}
	if deriver.lastInput.SlotIndex != 9 {
		t.Fatalf("unexpected index: got %d", deriver.lastInput.SlotIndex)
	}
	if deriver.lastInput.Chain != valueobjects.SupportedChainBitcoin {
		t.Fatalf("unexpected chain: got %q", deriver.lastInput.Chain)
	}
}

func TestGenerateAddressUseCaseRejectsEthereumCreate2Preview(t *testing.T) {
	deriver := newFakeChainAddressDeriver()
	deriver.supportedChains[valueobjects.SupportedChainEthereum] = true
	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		newEthereumCreate2IssuancePolicy(
			"ethereum-mainnet-create2",
			valueobjects.NetworkID("mainnet"),
			"create2.v1:factory=0x1111111111111111111111111111111111111111;collector=0x2222222222222222222222222222222222222222;init_code_hash=0x3333333333333333333333333333333333333333333333333333333333333333",
			"ethereum-mainnet-create2",
		),
	})
	useCase := NewGenerateAddressUseCase(deriver, catalog)

	response, err := useCase.Execute(context.Background(), dto.GenerateAddressInput{
		Chain:           valueobjects.SupportedChainEthereum,
		AddressPolicyID: "ethereum-mainnet-create2",
		Index:           9,
	})
	if err == nil {
		t.Fatalf("expected preview not supported error, got response %+v", response)
	}
	if !errors.Is(err, inport.ErrAddressPreviewNotSupported) {
		t.Fatalf("expected preview not supported error, got %v", err)
	}
	if deriver.calls != 0 {
		t.Fatalf("expected ethereum preview rejection before deriver call, got %d calls", deriver.calls)
	}
}

func TestGenerateAddressUseCaseRejectUnsupportedChain(t *testing.T) {
	useCase := NewGenerateAddressUseCase(
		newFakeChainAddressDeriver(),
		newInMemoryAddressPolicyReader(nil),
	)

	_, err := useCase.Execute(context.Background(), dto.GenerateAddressInput{
		Chain:           valueobjects.SupportedChain("eth"),
		AddressPolicyID: "eth-mainnet",
		Index:           0,
	})
	if !errors.Is(err, inport.ErrChainNotSupported) {
		t.Fatalf("expected chain not supported error, got %v", err)
	}
}

func TestGenerateAddressUseCaseRejectUnknownPolicy(t *testing.T) {
	useCase := NewGenerateAddressUseCase(
		newFakeChainAddressDeriver(),
		newInMemoryAddressPolicyReader(nil),
	)

	_, err := useCase.Execute(context.Background(), dto.GenerateAddressInput{
		Chain:           valueobjects.SupportedChainBitcoin,
		AddressPolicyID: "bitcoin-mainnet-legacy",
		Index:           0,
	})
	if !errors.Is(err, inport.ErrAddressPolicyNotFound) {
		t.Fatalf("expected address policy not found error, got %v", err)
	}
}

func TestGenerateAddressUseCaseRejectDisabledPolicy(t *testing.T) {
	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		newAddressIssuancePolicy(
			"bitcoin-mainnet-legacy",
			valueobjects.SupportedChainBitcoin,
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeLegacy),
			"satoshi",
			8,
			"",
			"",
		),
	})
	useCase := NewGenerateAddressUseCase(newFakeChainAddressDeriver(), catalog)

	_, err := useCase.Execute(context.Background(), dto.GenerateAddressInput{
		Chain:           valueobjects.SupportedChainBitcoin,
		AddressPolicyID: "bitcoin-mainnet-legacy",
		Index:           0,
	})
	if !errors.Is(err, inport.ErrAddressPolicyNotEnabled) {
		t.Fatalf("expected address policy not enabled error, got %v", err)
	}
}

func TestGenerateAddressUseCaseDerivationError(t *testing.T) {
	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		newAddressIssuancePolicy(
			"bitcoin-testnet4-native-segwit",
			valueobjects.SupportedChainBitcoin,
			valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			"satoshi",
			8,
			"tpub-testnet4",
			"m/84'/1'/0'",
		),
	})
	deriver := newFakeChainAddressDeriver()
	deriver.err = errors.New("derive failed")
	useCase := NewGenerateAddressUseCase(deriver, catalog)

	_, err := useCase.Execute(context.Background(), dto.GenerateAddressInput{
		Chain:           valueobjects.SupportedChainBitcoin,
		AddressPolicyID: "bitcoin-testnet4-native-segwit",
		Index:           3,
	})
	if !errors.Is(err, inport.ErrDependencyFailure) {
		t.Fatalf("expected ErrDependencyFailure, got %v", err)
	}
}

func TestGenerateAddressUseCaseMapsPolicyReaderFailure(t *testing.T) {
	reader := newInMemoryAddressPolicyReader(nil)
	reader.findErr = errors.New("query failed")
	useCase := NewGenerateAddressUseCase(newFakeChainAddressDeriver(), reader)

	_, err := useCase.Execute(context.Background(), dto.GenerateAddressInput{
		Chain:           valueobjects.SupportedChainBitcoin,
		AddressPolicyID: "bitcoin-mainnet-legacy",
		Index:           1,
	})
	if !errors.Is(err, inport.ErrDependencyFailure) {
		t.Fatalf("expected ErrDependencyFailure, got %v", err)
	}
}

func TestGenerateAddressUseCaseRoutesAllSchemes(t *testing.T) {
	tests := []struct {
		name            string
		addressPolicyID string
		scheme          string
	}{
		{name: "legacy", addressPolicyID: "bitcoin-mainnet-legacy", scheme: string(valueobjects.BitcoinAddressSchemeLegacy)},
		{name: "segwit", addressPolicyID: "bitcoin-mainnet-segwit", scheme: string(valueobjects.BitcoinAddressSchemeSegwit)},
		{name: "native segwit", addressPolicyID: "bitcoin-mainnet-native-segwit", scheme: string(valueobjects.BitcoinAddressSchemeNativeSegwit)},
		{name: "taproot", addressPolicyID: "bitcoin-mainnet-taproot", scheme: string(valueobjects.BitcoinAddressSchemeTaproot)},
	}

	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		newAddressIssuancePolicy("bitcoin-mainnet-legacy", valueobjects.SupportedChainBitcoin, valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet), string(valueobjects.BitcoinAddressSchemeLegacy), "satoshi", 8, "xpub-main", "m/44'/0'/0'"),
		newAddressIssuancePolicy("bitcoin-mainnet-segwit", valueobjects.SupportedChainBitcoin, valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet), string(valueobjects.BitcoinAddressSchemeSegwit), "satoshi", 8, "xpub-main", "m/49'/0'/0'"),
		newAddressIssuancePolicy("bitcoin-mainnet-native-segwit", valueobjects.SupportedChainBitcoin, valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet), string(valueobjects.BitcoinAddressSchemeNativeSegwit), "satoshi", 8, "xpub-main", "m/84'/0'/0'"),
		newAddressIssuancePolicy("bitcoin-mainnet-taproot", valueobjects.SupportedChainBitcoin, valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet), string(valueobjects.BitcoinAddressSchemeTaproot), "satoshi", 8, "xpub-main", "m/86'/0'/0'"),
	})

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			deriver := newFakeChainAddressDeriver()
			deriver.output = dtoToDeriveOutput("1BitcoinAddressExample", "0/1")
			useCase := NewGenerateAddressUseCase(deriver, catalog)

			_, err := useCase.Execute(context.Background(), dto.GenerateAddressInput{
				Chain:           valueobjects.SupportedChainBitcoin,
				AddressPolicyID: tc.addressPolicyID,
				Index:           1,
			})
			if err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}
			if deriver.lastInput.Scheme != tc.scheme {
				t.Fatalf("unexpected scheme routed to deriver: got %q, want %q", deriver.lastInput.Scheme, tc.scheme)
			}
		})
	}
}

func TestGenerateAddressUseCaseValidationMissingDependencies(t *testing.T) {
	input := dto.GenerateAddressInput{
		Chain:           valueobjects.SupportedChainBitcoin,
		AddressPolicyID: "bitcoin-mainnet-legacy",
		Index:           1,
	}

	tests := []struct {
		name    string
		useCase *generateAddressUseCase
		wantErr error
	}{
		{
			name:    "missing deriver",
			useCase: &generateAddressUseCase{policyReader: newInMemoryAddressPolicyReader(nil)},
			wantErr: inport.ErrChainAddressDeriverNotConfigured,
		},
		{
			name:    "missing policy reader",
			useCase: &generateAddressUseCase{deriver: newFakeChainAddressDeriver()},
			wantErr: inport.ErrAddressPolicyReaderNotConfigured,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.useCase.Execute(context.Background(), input)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("unexpected error: got %v want %v", err, tc.wantErr)
			}
		})
	}
}

func dtoToDeriveOutput(address string, path string) outport.DeriveChainAddressOutput {
	kind := valueobjects.IssuanceRefKindHDPathAbsolute
	if len(path) >= 2 && path[:2] == "0x" {
		kind = valueobjects.IssuanceRefKindCreate2Salt
	}
	return outport.DeriveChainAddressOutput{
		Address:             address,
		IssuanceRefKind:     kind,
		IssuanceRef:         path,
		RelativeIssuanceRef: path,
	}
}
