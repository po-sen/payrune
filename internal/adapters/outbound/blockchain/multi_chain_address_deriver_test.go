package blockchain

import (
	"context"
	"errors"
	"testing"

	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/value_objects"
)

type fakeChainSpecificAddressDeriver struct {
	chain     value_objects.SupportedChain
	output    outport.DeriveChainAddressOutput
	err       error
	lastInput outport.DeriveChainAddressInput
	calls     int
}

func (f *fakeChainSpecificAddressDeriver) Chain() value_objects.SupportedChain {
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
		chain: value_objects.SupportedChain("eth/mainnet"),
	})
	if err == nil {
		t.Fatal("expected error for invalid chain key")
	}

	_, err = NewMultiChainAddressDeriver(
		&fakeChainSpecificAddressDeriver{chain: value_objects.SupportedChainBitcoin},
		&fakeChainSpecificAddressDeriver{chain: value_objects.SupportedChain("bitcoin")},
	)
	if err == nil {
		t.Fatal("expected error for duplicate chain")
	}
}

func TestMultiChainAddressDeriverSupportsChain(t *testing.T) {
	deriver, err := NewMultiChainAddressDeriver(&fakeChainSpecificAddressDeriver{
		chain: value_objects.SupportedChainBitcoin,
	})
	if err != nil {
		t.Fatalf("setup deriver: %v", err)
	}

	if !deriver.SupportsChain(value_objects.SupportedChain("BitCoin")) {
		t.Fatal("expected bitcoin support")
	}
	if deriver.SupportsChain(value_objects.SupportedChain("eth")) {
		t.Fatal("expected ethereum unsupported")
	}
	if deriver.SupportsChain(value_objects.SupportedChain("eth/mainnet")) {
		t.Fatal("expected invalid chain unsupported")
	}
}

func TestMultiChainAddressDeriverRoutesToChainSpecificDeriver(t *testing.T) {
	bitcoin := &fakeChainSpecificAddressDeriver{
		chain: value_objects.SupportedChainBitcoin,
		output: outport.DeriveChainAddressOutput{
			Address:                "bc1qgenerated",
			RelativeDerivationPath: "0/12",
		},
	}
	deriver, err := NewMultiChainAddressDeriver(bitcoin)
	if err != nil {
		t.Fatalf("setup deriver: %v", err)
	}

	output, err := deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:            value_objects.SupportedChain(" BitCoin "),
		Network:          value_objects.NetworkID(" MainNet "),
		Scheme:           " nativeSegwit ",
		AccountPublicKey: " xpub-main ",
		Index:            12,
	})
	if err != nil {
		t.Fatalf("DeriveAddress returned error: %v", err)
	}
	if output.Address != "bc1qgenerated" {
		t.Fatalf("unexpected address: got %q", output.Address)
	}
	if bitcoin.calls != 1 {
		t.Fatalf("unexpected deriver calls: got %d", bitcoin.calls)
	}
	if bitcoin.lastInput.Chain != value_objects.SupportedChainBitcoin {
		t.Fatalf("unexpected normalized chain: got %q", bitcoin.lastInput.Chain)
	}
	if bitcoin.lastInput.Network != value_objects.NetworkID("mainnet") {
		t.Fatalf("unexpected normalized network: got %q", bitcoin.lastInput.Network)
	}
	if bitcoin.lastInput.Scheme != "nativeSegwit" {
		t.Fatalf("unexpected normalized scheme: got %q", bitcoin.lastInput.Scheme)
	}
	if bitcoin.lastInput.AccountPublicKey != "xpub-main" {
		t.Fatalf("unexpected normalized account public key: got %q", bitcoin.lastInput.AccountPublicKey)
	}
}

func TestMultiChainAddressDeriverDeriveAddressValidation(t *testing.T) {
	deriver, err := NewMultiChainAddressDeriver(&fakeChainSpecificAddressDeriver{
		chain: value_objects.SupportedChainBitcoin,
	})
	if err != nil {
		t.Fatalf("setup deriver: %v", err)
	}

	_, err = deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:   value_objects.SupportedChain("eth/mainnet"),
		Network: value_objects.NetworkID("mainnet"),
	})
	if err == nil {
		t.Fatal("expected invalid chain error")
	}

	_, err = deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:   value_objects.SupportedChainBitcoin,
		Network: value_objects.NetworkID("main/net"),
	})
	if err == nil {
		t.Fatal("expected invalid network error")
	}

	_, err = deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:   value_objects.SupportedChain("eth"),
		Network: value_objects.NetworkID("mainnet"),
	})
	if err == nil {
		t.Fatal("expected missing deriver error")
	}
}

func TestMultiChainAddressDeriverPassesThroughErrors(t *testing.T) {
	expectedErr := errors.New("boom")
	deriver, err := NewMultiChainAddressDeriver(&fakeChainSpecificAddressDeriver{
		chain: value_objects.SupportedChainBitcoin,
		err:   expectedErr,
	})
	if err != nil {
		t.Fatalf("setup deriver: %v", err)
	}

	_, err = deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:   value_objects.SupportedChainBitcoin,
		Network: value_objects.NetworkID("mainnet"),
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected downstream error, got %v", err)
	}
}
