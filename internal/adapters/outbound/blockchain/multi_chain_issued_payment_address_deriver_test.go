package blockchain

import (
	"context"
	"errors"
	"testing"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
)

type fakeChainSpecificIssuedPaymentAddressDeriver struct {
	chain     valueobjects.SupportedChain
	output    outport.DeriveIssuedPaymentAddressOutput
	err       error
	lastInput outport.DeriveIssuedPaymentAddressInput
	calls     int
}

func (f *fakeChainSpecificIssuedPaymentAddressDeriver) Chain() valueobjects.SupportedChain {
	return f.chain
}

func (f *fakeChainSpecificIssuedPaymentAddressDeriver) DeriveIssuedAddress(
	_ context.Context,
	input outport.DeriveIssuedPaymentAddressInput,
) (outport.DeriveIssuedPaymentAddressOutput, error) {
	f.calls++
	f.lastInput = input
	if f.err != nil {
		return outport.DeriveIssuedPaymentAddressOutput{}, f.err
	}
	return f.output, nil
}

func TestNewMultiChainIssuedPaymentAddressDeriverRequiresDerivers(t *testing.T) {
	_, err := NewMultiChainIssuedPaymentAddressDeriver()
	if err == nil {
		t.Fatal("expected missing-derivers error")
	}
}

func TestNewMultiChainIssuedPaymentAddressDeriverRejectsDuplicateChains(t *testing.T) {
	_, err := NewMultiChainIssuedPaymentAddressDeriver(
		&fakeChainSpecificIssuedPaymentAddressDeriver{chain: valueobjects.SupportedChainBitcoin},
		&fakeChainSpecificIssuedPaymentAddressDeriver{chain: valueobjects.SupportedChainBitcoin},
	)
	if err == nil {
		t.Fatal("expected duplicate-chain error")
	}
}

func TestMultiChainIssuedPaymentAddressDeriverSupportsConfiguredChain(t *testing.T) {
	deriver, err := NewMultiChainIssuedPaymentAddressDeriver(
		&fakeChainSpecificIssuedPaymentAddressDeriver{chain: valueobjects.SupportedChainBitcoin},
	)
	if err != nil {
		t.Fatalf("NewMultiChainIssuedPaymentAddressDeriver returned error: %v", err)
	}

	if !deriver.SupportsChain(valueobjects.SupportedChainBitcoin) {
		t.Fatal("expected bitcoin to be supported")
	}
	if deriver.SupportsChain(valueobjects.SupportedChainEthereum) {
		t.Fatal("expected ethereum to be unsupported")
	}
}

func TestMultiChainIssuedPaymentAddressDeriverDispatchesByChain(t *testing.T) {
	bitcoinDeriver := &fakeChainSpecificIssuedPaymentAddressDeriver{
		chain: valueobjects.SupportedChainBitcoin,
		output: outport.DeriveIssuedPaymentAddressOutput{
			Address:         "bc1qissued",
			IssuanceRefKind: valueobjects.IssuanceRefKindHDPathAbsolute,
			IssuanceRef:     "m/84'/0'/0'/0/8",
		},
	}
	ethereumDeriver := &fakeChainSpecificIssuedPaymentAddressDeriver{
		chain: valueobjects.SupportedChainEthereum,
	}

	deriver, err := NewMultiChainIssuedPaymentAddressDeriver(bitcoinDeriver, ethereumDeriver)
	if err != nil {
		t.Fatalf("NewMultiChainIssuedPaymentAddressDeriver returned error: %v", err)
	}

	output, err := deriver.DeriveIssuedAddress(context.Background(), outport.DeriveIssuedPaymentAddressInput{
		Policy: policies.AddressIssuancePolicy{
			AddressPolicyID: "bitcoin-mainnet-native-segwit",
			Chain:           valueobjects.SupportedChainBitcoin,
			Network:         valueobjects.NetworkIDMainnet,
			Scheme:          valueobjects.AddressSchemeNativeSegwit,
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef:   "xpub-main",
				IssuanceRefPrefix: "m/84'/0'/0'",
			},
		},
		Allocation: entities.PaymentAddressAllocation{
			PaymentAddressID: 88,
			SlotIndex:        8,
		},
	})
	if err != nil {
		t.Fatalf("DeriveIssuedAddress returned error: %v", err)
	}
	if output.Address != "bc1qissued" {
		t.Fatalf("unexpected address: got %q", output.Address)
	}
	if bitcoinDeriver.calls != 1 {
		t.Fatalf("expected bitcoin deriver call count 1, got %d", bitcoinDeriver.calls)
	}
	if ethereumDeriver.calls != 0 {
		t.Fatalf("expected ethereum deriver call count 0, got %d", ethereumDeriver.calls)
	}
}

func TestMultiChainIssuedPaymentAddressDeriverPropagatesChainSpecificError(t *testing.T) {
	bitcoinDeriver := &fakeChainSpecificIssuedPaymentAddressDeriver{
		chain: valueobjects.SupportedChainBitcoin,
		err:   errors.New("derive failed"),
	}

	deriver, err := NewMultiChainIssuedPaymentAddressDeriver(bitcoinDeriver)
	if err != nil {
		t.Fatalf("NewMultiChainIssuedPaymentAddressDeriver returned error: %v", err)
	}

	_, err = deriver.DeriveIssuedAddress(context.Background(), outport.DeriveIssuedPaymentAddressInput{
		Policy: policies.AddressIssuancePolicy{
			AddressPolicyID: "bitcoin-mainnet-native-segwit",
			Chain:           valueobjects.SupportedChainBitcoin,
			Network:         valueobjects.NetworkIDMainnet,
			Scheme:          valueobjects.AddressSchemeNativeSegwit,
		},
	})
	if !errors.Is(err, outport.ErrIssuedPaymentAddressDerivationFailed) {
		t.Fatalf("expected %v, got %v", outport.ErrIssuedPaymentAddressDerivationFailed, err)
	}
}
