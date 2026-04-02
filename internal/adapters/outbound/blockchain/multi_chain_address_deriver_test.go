package blockchain

import (
	"context"
	"errors"
	"testing"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

type fakeChainSpecificAddressDeriver struct {
	chain     valueobjects.SupportedChain
	output    outport.DeriveChainAddressOutput
	err       error
	lastInput outport.DeriveChainAddressInput
	calls     int
}

func (f *fakeChainSpecificAddressDeriver) Chain() valueobjects.SupportedChain {
	return f.chain
}

func (f *fakeChainSpecificAddressDeriver) DeriveAddress(
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

func TestNewMultiChainAddressDeriverValidation(t *testing.T) {
	_, err := NewMultiChainAddressDeriver()
	if err == nil {
		t.Fatal("expected error for empty deriver list")
	}

	_, err = NewMultiChainAddressDeriver((*fakeChainSpecificAddressDeriver)(nil))
	if err == nil {
		t.Fatal("expected error for nil deriver")
	}

	_, err = NewMultiChainAddressDeriver(&fakeChainSpecificAddressDeriver{
		chain: valueobjects.SupportedChain("eth/mainnet"),
	})
	if err == nil {
		t.Fatal("expected error for invalid chain key")
	}

	_, err = NewMultiChainAddressDeriver(&fakeChainSpecificAddressDeriver{
		chain: valueobjects.SupportedChain("eth"),
	})
	if err == nil {
		t.Fatal("expected error for unsupported chain key")
	}

	_, err = NewMultiChainAddressDeriver(
		&fakeChainSpecificAddressDeriver{chain: valueobjects.SupportedChainBitcoin},
		&fakeChainSpecificAddressDeriver{chain: valueobjects.SupportedChain("bitcoin")},
	)
	if err == nil {
		t.Fatal("expected error for duplicate chain")
	}
}

func TestMultiChainAddressDeriverSupportsChain(t *testing.T) {
	deriver, err := NewMultiChainAddressDeriver(
		&fakeChainSpecificAddressDeriver{chain: valueobjects.SupportedChainBitcoin},
		&fakeChainSpecificAddressDeriver{chain: valueobjects.SupportedChainEthereum},
	)
	if err != nil {
		t.Fatalf("setup deriver: %v", err)
	}

	if !deriver.SupportsChain(valueobjects.SupportedChain("BitCoin")) {
		t.Fatal("expected bitcoin support")
	}
	if !deriver.SupportsChain(valueobjects.SupportedChain("ethereum")) {
		t.Fatal("expected ethereum support")
	}
	if deriver.SupportsChain(valueobjects.SupportedChain("eth")) {
		t.Fatal("expected eth alias unsupported")
	}
	if deriver.SupportsChain(valueobjects.SupportedChain("eth/mainnet")) {
		t.Fatal("expected invalid chain unsupported")
	}
}

func TestMultiChainAddressDeriverRoutesToChainSpecificDeriver(t *testing.T) {
	bitcoin := &fakeChainSpecificAddressDeriver{
		chain: valueobjects.SupportedChainBitcoin,
		output: outport.DeriveChainAddressOutput{
			Address:             "bc1qgenerated",
			IssuanceRefKind:     valueobjects.IssuanceRefKindHDPathAbsolute,
			RelativeIssuanceRef: "0/12",
			IssuanceRef:         "m/84'/0'/5'/0/12",
		},
	}
	deriver, err := NewMultiChainAddressDeriver(bitcoin)
	if err != nil {
		t.Fatalf("setup deriver: %v", err)
	}

	output, err := deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:               valueobjects.SupportedChain(" BitCoin "),
		Network:             valueobjects.NetworkID(" MainNet "),
		Scheme:              " nativeSegwit ",
		AddressSpaceRef:     " xpub-main ",
		IssuanceRefPrefix:   " m/84'/0'/0' ",
		RelativeIssuanceRef: " 0/12 ",
		SlotIndex:           12,
	})
	if err != nil {
		t.Fatalf("DeriveAddress returned error: %v", err)
	}
	if output.Address != "bc1qgenerated" {
		t.Fatalf("unexpected address: got %q", output.Address)
	}
	if output.IssuanceRef != "m/84'/0'/5'/0/12" {
		t.Fatalf("unexpected address reference: got %q", output.IssuanceRef)
	}
	if bitcoin.calls != 1 {
		t.Fatalf("unexpected deriver calls: got %d", bitcoin.calls)
	}
	if bitcoin.lastInput.Chain != valueobjects.SupportedChainBitcoin {
		t.Fatalf("unexpected normalized chain: got %q", bitcoin.lastInput.Chain)
	}
	if bitcoin.lastInput.Network != valueobjects.NetworkIDMainnet {
		t.Fatalf("unexpected normalized network: got %q", bitcoin.lastInput.Network)
	}
	if bitcoin.lastInput.Scheme != "nativeSegwit" {
		t.Fatalf("unexpected normalized scheme: got %q", bitcoin.lastInput.Scheme)
	}
	if bitcoin.lastInput.AddressSpaceRef != "xpub-main" {
		t.Fatalf("unexpected normalized address source ref: got %q", bitcoin.lastInput.AddressSpaceRef)
	}
	if bitcoin.lastInput.IssuanceRefPrefix != "m/84'/0'/0'" {
		t.Fatalf("unexpected normalized address reference prefix: got %q", bitcoin.lastInput.IssuanceRefPrefix)
	}
	if bitcoin.lastInput.RelativeIssuanceRef != "0/12" {
		t.Fatalf("unexpected normalized relative address reference: got %q", bitcoin.lastInput.RelativeIssuanceRef)
	}
}

func TestMultiChainAddressDeriverDeriveAddressValidation(t *testing.T) {
	deriver, err := NewMultiChainAddressDeriver(&fakeChainSpecificAddressDeriver{
		chain: valueobjects.SupportedChainBitcoin,
	})
	if err != nil {
		t.Fatalf("setup deriver: %v", err)
	}

	_, err = deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:   valueobjects.SupportedChain("eth/mainnet"),
		Network: valueobjects.NetworkIDMainnet,
	})
	if err == nil {
		t.Fatal("expected invalid chain error")
	}

	_, err = deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:   valueobjects.SupportedChainBitcoin,
		Network: valueobjects.NetworkID("main/net"),
	})
	if err == nil {
		t.Fatal("expected invalid network error")
	}

	_, err = deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:   valueobjects.SupportedChain("eth"),
		Network: valueobjects.NetworkIDMainnet,
	})
	if err == nil {
		t.Fatal("expected invalid chain error")
	}

	_, err = deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:   valueobjects.SupportedChainEthereum,
		Network: valueobjects.NetworkIDMainnet,
	})
	if err == nil {
		t.Fatal("expected missing deriver error")
	}
}

func TestMultiChainAddressDeriverPassesThroughErrors(t *testing.T) {
	deriver, err := NewMultiChainAddressDeriver(&fakeChainSpecificAddressDeriver{
		chain: valueobjects.SupportedChainBitcoin,
		err:   errors.New("boom"),
	})
	if err != nil {
		t.Fatalf("setup deriver: %v", err)
	}

	_, err = deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:   valueobjects.SupportedChainBitcoin,
		Network: valueobjects.NetworkIDMainnet,
	})
	if !errors.Is(err, outport.ErrChainAddressDerivationFailed) {
		t.Fatalf("expected %v, got %v", outport.ErrChainAddressDerivationFailed, err)
	}
}
