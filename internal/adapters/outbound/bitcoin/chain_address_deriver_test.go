package bitcoin

import (
	"context"
	"errors"
	"testing"

	outport "payrune/internal/application/ports/outbound"
)

type fakeBitcoinAddressDeriver struct {
	address               string
	err                   error
	path                  string
	absolutePath          string
	pathErr               error
	lastNetwork           network
	lastScheme            addressScheme
	lastXPub              string
	lastIssuanceRefPrefix string
	lastIndex             uint32
}

func (f *fakeBitcoinAddressDeriver) DeriveAddress(
	network network,
	scheme addressScheme,
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
	f.lastIssuanceRefPrefix = prefix
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

	if deriver.Chain() != outport.SupportedChainBitcoin {
		t.Fatalf("unexpected chain: got %q", deriver.Chain())
	}
	if !deriver.SupportsChain(outport.SupportedChainBitcoin) {
		t.Fatal("expected bitcoin to be supported")
	}
	if deriver.SupportsChain("eth") {
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
		Chain:             outport.SupportedChainBitcoin,
		Network:           outport.NetworkIDMainnet,
		Scheme:            outport.AddressSchemeNativeSegwit,
		AddressSpaceRef:   "xpub-main",
		IssuanceRefPrefix: "m/84'/0'/0'",
		SlotIndex:         12,
	})
	if err != nil {
		t.Fatalf("DeriveAddress returned error: %v", err)
	}
	if output.Address != "bc1qgenerated" {
		t.Fatalf("unexpected address: got %q", output.Address)
	}
	if output.RelativeIssuanceRef != "0/12" {
		t.Fatalf("unexpected relative address reference: got %q", output.RelativeIssuanceRef)
	}
	if output.IssuanceRef != "m/84'/0'/5'/0/12" {
		t.Fatalf("unexpected address reference: got %q", output.IssuanceRef)
	}
	if deriver.lastNetwork != networkMainnet {
		t.Fatalf("unexpected network: got %q", deriver.lastNetwork)
	}
	if deriver.lastScheme != addressSchemeNativeSegwit {
		t.Fatalf("unexpected scheme: got %q", deriver.lastScheme)
	}
	if deriver.lastXPub != "xpub-main" {
		t.Fatalf("unexpected address source ref: got %q", deriver.lastXPub)
	}
	if deriver.lastIssuanceRefPrefix != "m/84'/0'/0'" {
		t.Fatalf("unexpected address reference prefix: got %q", deriver.lastIssuanceRefPrefix)
	}
	if deriver.lastIndex != 12 {
		t.Fatalf("unexpected index: got %d", deriver.lastIndex)
	}
}

func TestChainAddressDeriverRejectUnsupportedChain(t *testing.T) {
	deriver := NewChainAddressDeriver(&fakeBitcoinAddressDeriver{})

	_, err := deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:           "eth",
		Network:         "mainnet",
		Scheme:          "legacy",
		AddressSpaceRef: "xpub-main",
		SlotIndex:       1,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestChainAddressDeriverReturnsDeriverError(t *testing.T) {
	deriver := NewChainAddressDeriver(&fakeBitcoinAddressDeriver{err: errors.New("derive failed")})

	_, err := deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:           outport.SupportedChainBitcoin,
		Network:         outport.NetworkIDTestnet4,
		Scheme:          outport.AddressSchemeTaproot,
		AddressSpaceRef: "tpub-testnet4",
		SlotIndex:       2,
	})
	if !errors.Is(err, outport.ErrChainAddressDerivationFailed) {
		t.Fatalf("expected %v, got %v", outport.ErrChainAddressDerivationFailed, err)
	}
}

func TestChainAddressDeriverRejectsInvalidNetwork(t *testing.T) {
	deriver := NewChainAddressDeriver(&fakeBitcoinAddressDeriver{})

	_, err := deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:           outport.SupportedChainBitcoin,
		Network:         "sepolia",
		Scheme:          outport.AddressSchemeLegacy,
		AddressSpaceRef: "xpub-main",
		SlotIndex:       1,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestChainAddressDeriverRejectsInvalidScheme(t *testing.T) {
	deriver := NewChainAddressDeriver(&fakeBitcoinAddressDeriver{})

	_, err := deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:           outport.SupportedChainBitcoin,
		Network:         outport.NetworkIDMainnet,
		Scheme:          "eip55",
		AddressSpaceRef: "xpub-main",
		SlotIndex:       1,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
