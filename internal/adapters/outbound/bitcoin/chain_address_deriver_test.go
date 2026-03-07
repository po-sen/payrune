package bitcoin

import (
	"context"
	"errors"
	"testing"

	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/value_objects"
)

type fakeBitcoinAddressDeriver struct {
	address     string
	err         error
	path        string
	pathErr     error
	lastNetwork value_objects.BitcoinNetwork
	lastScheme  value_objects.BitcoinAddressScheme
	lastXPub    string
	lastIndex   uint32
}

func (f *fakeBitcoinAddressDeriver) DeriveAddress(
	network value_objects.BitcoinNetwork,
	scheme value_objects.BitcoinAddressScheme,
	xpub string,
	index uint32,
) (string, error) {
	f.lastNetwork = network
	f.lastScheme = scheme
	f.lastXPub = xpub
	f.lastIndex = index
	if f.err != nil {
		return "", f.err
	}
	return f.address, nil
}

func (f *fakeBitcoinAddressDeriver) DerivationPath(_ string, _ uint32) (string, error) {
	if f.pathErr != nil {
		return "", f.pathErr
	}
	if f.path == "" {
		return "0/0", nil
	}
	return f.path, nil
}

func TestChainAddressDeriverSupportsBitcoinOnly(t *testing.T) {
	deriver := NewChainAddressDeriver(&fakeBitcoinAddressDeriver{})

	if deriver.Chain() != value_objects.SupportedChainBitcoin {
		t.Fatalf("unexpected chain: got %q", deriver.Chain())
	}
	if !deriver.SupportsChain(value_objects.SupportedChainBitcoin) {
		t.Fatal("expected bitcoin to be supported")
	}
	if deriver.SupportsChain(value_objects.SupportedChain("eth")) {
		t.Fatal("expected eth not to be supported")
	}
}

func TestChainAddressDeriverDeriveAddress(t *testing.T) {
	deriver := &fakeBitcoinAddressDeriver{address: "bc1qgenerated", path: "0/12"}
	generator := NewChainAddressDeriver(deriver)

	output, err := generator.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:            value_objects.SupportedChainBitcoin,
		Network:          value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
		Scheme:           string(value_objects.BitcoinAddressSchemeNativeSegwit),
		AccountPublicKey: "xpub-main",
		Index:            12,
	})
	if err != nil {
		t.Fatalf("DeriveAddress returned error: %v", err)
	}
	if output.Address != "bc1qgenerated" {
		t.Fatalf("unexpected address: got %q", output.Address)
	}
	if output.RelativeDerivationPath != "0/12" {
		t.Fatalf("unexpected derivation path: got %q", output.RelativeDerivationPath)
	}
	if deriver.lastNetwork != value_objects.BitcoinNetworkMainnet {
		t.Fatalf("unexpected network: got %q", deriver.lastNetwork)
	}
	if deriver.lastScheme != value_objects.BitcoinAddressSchemeNativeSegwit {
		t.Fatalf("unexpected scheme: got %q", deriver.lastScheme)
	}
	if deriver.lastXPub != "xpub-main" {
		t.Fatalf("unexpected public key: got %q", deriver.lastXPub)
	}
	if deriver.lastIndex != 12 {
		t.Fatalf("unexpected index: got %d", deriver.lastIndex)
	}
}

func TestChainAddressDeriverRejectUnsupportedChain(t *testing.T) {
	deriver := NewChainAddressDeriver(&fakeBitcoinAddressDeriver{})

	_, err := deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:            value_objects.SupportedChain("eth"),
		Network:          "mainnet",
		Scheme:           "legacy",
		AccountPublicKey: "xpub-main",
		Index:            1,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestChainAddressDeriverReturnsDeriverError(t *testing.T) {
	expectedErr := errors.New("derive failed")
	deriver := NewChainAddressDeriver(&fakeBitcoinAddressDeriver{err: expectedErr})

	_, err := deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:            value_objects.SupportedChainBitcoin,
		Network:          value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4),
		Scheme:           string(value_objects.BitcoinAddressSchemeTaproot),
		AccountPublicKey: "tpub-testnet4",
		Index:            2,
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestChainAddressDeriverRejectsInvalidNetwork(t *testing.T) {
	deriver := NewChainAddressDeriver(&fakeBitcoinAddressDeriver{})

	_, err := deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:            value_objects.SupportedChainBitcoin,
		Network:          "sepolia",
		Scheme:           string(value_objects.BitcoinAddressSchemeLegacy),
		AccountPublicKey: "xpub-main",
		Index:            1,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestChainAddressDeriverRejectsInvalidScheme(t *testing.T) {
	deriver := NewChainAddressDeriver(&fakeBitcoinAddressDeriver{})

	_, err := deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:            value_objects.SupportedChainBitcoin,
		Network:          value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
		Scheme:           "eip55",
		AccountPublicKey: "xpub-main",
		Index:            1,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
