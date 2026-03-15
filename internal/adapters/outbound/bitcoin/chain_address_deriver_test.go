package bitcoin

import (
	"context"
	"errors"
	"testing"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

type fakeBitcoinAddressDeriver struct {
	address                  string
	err                      error
	path                     string
	absolutePath             string
	pathErr                  error
	lastNetwork              valueobjects.BitcoinNetwork
	lastScheme               valueobjects.BitcoinAddressScheme
	lastXPub                 string
	lastDerivationPathPrefix string
	lastIndex                uint32
}

func (f *fakeBitcoinAddressDeriver) DeriveAddress(
	network valueobjects.BitcoinNetwork,
	scheme valueobjects.BitcoinAddressScheme,
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

func (f *fakeBitcoinAddressDeriver) AbsoluteDerivationPath(_ string, prefix string, _ uint32) (string, error) {
	f.lastDerivationPathPrefix = prefix
	if f.pathErr != nil {
		return "", f.pathErr
	}
	if f.absolutePath != "" {
		return f.absolutePath, nil
	}
	if f.path == "" {
		return prefix + "/0/0", nil
	}
	return prefix + "/" + f.path, nil
}

func TestChainAddressDeriverSupportsBitcoinOnly(t *testing.T) {
	deriver := NewChainAddressDeriver(&fakeBitcoinAddressDeriver{})

	if deriver.Chain() != valueobjects.SupportedChainBitcoin {
		t.Fatalf("unexpected chain: got %q", deriver.Chain())
	}
	if !deriver.SupportsChain(valueobjects.SupportedChainBitcoin) {
		t.Fatal("expected bitcoin to be supported")
	}
	if deriver.SupportsChain(valueobjects.SupportedChain("eth")) {
		t.Fatal("expected eth not to be supported")
	}
}

func TestChainAddressDeriverDeriveAddress(t *testing.T) {
	deriver := &fakeBitcoinAddressDeriver{
		address:      "bc1qgenerated",
		path:         "0/12",
		absolutePath: "m/84'/0'/5'/0/12",
	}
	generator := NewChainAddressDeriver(deriver)

	output, err := generator.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:                valueobjects.SupportedChainBitcoin,
		Network:              valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
		Scheme:               string(valueobjects.BitcoinAddressSchemeNativeSegwit),
		AccountPublicKey:     "xpub-main",
		DerivationPathPrefix: "m/84'/0'/0'",
		Index:                12,
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
	if output.DerivationPath != "m/84'/0'/5'/0/12" {
		t.Fatalf("unexpected absolute derivation path: got %q", output.DerivationPath)
	}
	if deriver.lastNetwork != valueobjects.BitcoinNetworkMainnet {
		t.Fatalf("unexpected network: got %q", deriver.lastNetwork)
	}
	if deriver.lastScheme != valueobjects.BitcoinAddressSchemeNativeSegwit {
		t.Fatalf("unexpected scheme: got %q", deriver.lastScheme)
	}
	if deriver.lastXPub != "xpub-main" {
		t.Fatalf("unexpected public key: got %q", deriver.lastXPub)
	}
	if deriver.lastDerivationPathPrefix != "m/84'/0'/0'" {
		t.Fatalf("unexpected derivation path prefix: got %q", deriver.lastDerivationPathPrefix)
	}
	if deriver.lastIndex != 12 {
		t.Fatalf("unexpected index: got %d", deriver.lastIndex)
	}
}

func TestChainAddressDeriverRejectUnsupportedChain(t *testing.T) {
	deriver := NewChainAddressDeriver(&fakeBitcoinAddressDeriver{})

	_, err := deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:            valueobjects.SupportedChain("eth"),
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
		Chain:            valueobjects.SupportedChainBitcoin,
		Network:          valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
		Scheme:           string(valueobjects.BitcoinAddressSchemeTaproot),
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
		Chain:            valueobjects.SupportedChainBitcoin,
		Network:          "sepolia",
		Scheme:           string(valueobjects.BitcoinAddressSchemeLegacy),
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
		Chain:            valueobjects.SupportedChainBitcoin,
		Network:          valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
		Scheme:           "eip55",
		AccountPublicKey: "xpub-main",
		Index:            1,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
