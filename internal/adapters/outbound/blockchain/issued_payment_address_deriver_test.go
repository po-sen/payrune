package blockchain

import (
	"context"
	"errors"
	"testing"

	ethereumadapter "payrune/internal/adapters/outbound/ethereum"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/valueobjects"
)

type fakeIssuedAddressChainDeriver struct {
	supportedChains map[valueobjects.SupportedChain]bool
	output          outport.DeriveChainAddressOutput
	err             error
	lastInput       outport.DeriveChainAddressInput
	calls           int
}

func (f *fakeIssuedAddressChainDeriver) SupportsChain(chain valueobjects.SupportedChain) bool {
	return f.supportedChains[chain]
}

func (f *fakeIssuedAddressChainDeriver) DeriveAddress(
	_ context.Context,
	input outport.DeriveChainAddressInput,
) (outport.DeriveChainAddressOutput, error) {
	f.calls++
	f.lastInput = input
	if f.err != nil {
		return outport.DeriveChainAddressOutput{}, f.err
	}
	return f.output, nil
}

func TestIssuedPaymentAddressDeriverDerivesBitcoinAddress(t *testing.T) {
	chainDeriver := &fakeIssuedAddressChainDeriver{
		supportedChains: map[valueobjects.SupportedChain]bool{
			valueobjects.SupportedChainBitcoin: true,
		},
		output: outport.DeriveChainAddressOutput{
			Address:          "bc1qallocated",
			AddressReference: "m/84'/0'/0'/0/5",
		},
	}
	deriver := NewIssuedPaymentAddressDeriver(chainDeriver, nil)

	output, err := deriver.DeriveIssuedAddress(context.Background(), outport.DeriveIssuedPaymentAddressInput{
		Policy: entities.AddressIssuancePolicy{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: "bitcoin-mainnet-native-segwit",
				Chain:           valueobjects.SupportedChainBitcoin,
				Network:         valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
				Scheme:          string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			},
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSourceRef:       "xpub-main",
				AddressReferencePrefix: "m/84'/0'/0'",
			},
		}.Normalize(),
		Allocation: entities.PaymentAddressAllocation{
			PaymentAddressID: 55,
			DerivationIndex:  5,
		},
	})
	if err != nil {
		t.Fatalf("DeriveIssuedAddress returned error: %v", err)
	}
	if output.Address != "bc1qallocated" {
		t.Fatalf("unexpected address: got %q", output.Address)
	}
	if output.AddressReference != "m/84'/0'/0'/0/5" {
		t.Fatalf("unexpected address reference: got %q", output.AddressReference)
	}
	if chainDeriver.lastInput.RelativeAddressReference != "" {
		t.Fatalf("expected empty relative reference for bitcoin, got %q", chainDeriver.lastInput.RelativeAddressReference)
	}
}

func TestIssuedPaymentAddressDeriverDerivesEthereumCreate2RelativeReference(t *testing.T) {
	chainDeriver := &fakeIssuedAddressChainDeriver{
		supportedChains: map[valueobjects.SupportedChain]bool{
			valueobjects.SupportedChainEthereum: true,
		},
		output: outport.DeriveChainAddressOutput{
			Address:                  "0x1234567890abcdef1234567890abcdef12345678",
			RelativeAddressReference: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
	}
	create2SaltDeriver := ethereumadapter.NewCreate2SaltDeriver(map[valueobjects.NetworkID]string{
		valueobjects.NetworkID("mainnet"): "0x1111111111111111111111111111111111111111111111111111111111111111",
	})
	deriver := NewIssuedPaymentAddressDeriver(chainDeriver, create2SaltDeriver)

	output, err := deriver.DeriveIssuedAddress(context.Background(), outport.DeriveIssuedPaymentAddressInput{
		Policy: entities.AddressIssuancePolicy{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: "ethereum-mainnet-create2",
				Chain:           valueobjects.SupportedChainEthereum,
				Network:         valueobjects.NetworkID("mainnet"),
				Scheme:          "create2",
			},
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSourceRef:       "create2.v1:factory=0x1111111111111111111111111111111111111111;collector=0x2222222222222222222222222222222222222222;init_code_hash=0x3333333333333333333333333333333333333333333333333333333333333333",
				AddressReferencePrefix: "ethereum-mainnet-create2",
			},
		}.Normalize(),
		Allocation: entities.PaymentAddressAllocation{
			PaymentAddressID: 145,
			DerivationIndex:  11,
		},
	})
	if err != nil {
		t.Fatalf("DeriveIssuedAddress returned error: %v", err)
	}
	if output.Address != "0x1234567890abcdef1234567890abcdef12345678" {
		t.Fatalf("unexpected address: got %q", output.Address)
	}
	if output.AddressReference != "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("unexpected address reference: got %q", output.AddressReference)
	}
	if chainDeriver.lastInput.RelativeAddressReference == "" {
		t.Fatal("expected create2 relative address reference to be passed to chain deriver")
	}
}

func TestIssuedPaymentAddressDeriverReturnsErrorBeforeChainDerivationWhenCreate2SaltFails(t *testing.T) {
	chainDeriver := &fakeIssuedAddressChainDeriver{
		supportedChains: map[valueobjects.SupportedChain]bool{
			valueobjects.SupportedChainEthereum: true,
		},
		output: outport.DeriveChainAddressOutput{Address: "ignored"},
	}
	deriver := NewIssuedPaymentAddressDeriver(chainDeriver, nil)

	_, err := deriver.DeriveIssuedAddress(context.Background(), outport.DeriveIssuedPaymentAddressInput{
		Policy: entities.AddressIssuancePolicy{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: "ethereum-mainnet-create2",
				Chain:           valueobjects.SupportedChainEthereum,
				Network:         valueobjects.NetworkID("mainnet"),
				Scheme:          "create2",
			},
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSourceRef:       "configured",
				AddressReferencePrefix: "ethereum-mainnet-create2",
			},
		}.Normalize(),
		Allocation: entities.PaymentAddressAllocation{
			PaymentAddressID: 145,
			DerivationIndex:  11,
		},
	})
	if err == nil {
		t.Fatal("expected create2 salt deriver error")
	}
	if chainDeriver.calls != 0 {
		t.Fatalf("expected chain deriver not to be called, got %d calls", chainDeriver.calls)
	}
}

func TestIssuedPaymentAddressDeriverPropagatesChainDeriverError(t *testing.T) {
	expectedErr := errors.New("derive failed")
	chainDeriver := &fakeIssuedAddressChainDeriver{
		supportedChains: map[valueobjects.SupportedChain]bool{
			valueobjects.SupportedChainBitcoin: true,
		},
		err: expectedErr,
	}
	deriver := NewIssuedPaymentAddressDeriver(chainDeriver, nil)

	_, err := deriver.DeriveIssuedAddress(context.Background(), outport.DeriveIssuedPaymentAddressInput{
		Policy: entities.AddressIssuancePolicy{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: "bitcoin-mainnet-native-segwit",
				Chain:           valueobjects.SupportedChainBitcoin,
				Network:         valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
				Scheme:          string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			},
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSourceRef:       "xpub-main",
				AddressReferencePrefix: "m/84'/0'/0'",
			},
		}.Normalize(),
		Allocation: entities.PaymentAddressAllocation{
			PaymentAddressID: 44,
			DerivationIndex:  11,
		},
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}
