package bitcoin

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"payrune/internal/domain/value_objects"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
)

func TestHDXPubAddressDeriverDeriveDeterministicLegacy(t *testing.T) {
	xpub := newTestXPub(t, &chaincfg.MainNetParams)
	deriver := newTestDeriver()

	first, err := deriver.DeriveAddress(
		value_objects.BitcoinNetworkMainnet,
		value_objects.BitcoinAddressSchemeLegacy,
		xpub,
		7,
	)
	if err != nil {
		t.Fatalf("DeriveAddress returned error: %v", err)
	}
	second, err := deriver.DeriveAddress(
		value_objects.BitcoinNetworkMainnet,
		value_objects.BitcoinAddressSchemeLegacy,
		xpub,
		7,
	)
	if err != nil {
		t.Fatalf("DeriveAddress returned error on second call: %v", err)
	}

	if first != second {
		t.Fatalf("expected deterministic address, got %q and %q", first, second)
	}
}

func TestHDXPubAddressDeriverAccountXPubUsesExternalChainBranch(t *testing.T) {
	xpub := newAccountLevelXPub(t, &chaincfg.MainNetParams)
	deriver := newTestDeriver()
	index := uint32(9)

	got, err := deriver.DeriveAddress(
		value_objects.BitcoinNetworkMainnet,
		value_objects.BitcoinAddressSchemeNativeSegwit,
		xpub,
		index,
	)
	if err != nil {
		t.Fatalf("DeriveAddress returned error: %v", err)
	}

	want := deriveExpectedAddress(
		t,
		xpub,
		&chaincfg.MainNetParams,
		NewNativeSegwitAddressEncoder(),
		index,
		true,
	)
	if got != want {
		t.Fatalf("unexpected address: got %q, want %q", got, want)
	}
}

func TestHDXPubAddressDeriverChangeXPubUsesDirectIndex(t *testing.T) {
	xpub := newChangeLevelXPub(t, &chaincfg.MainNetParams)
	deriver := newTestDeriver()
	index := uint32(9)

	got, err := deriver.DeriveAddress(
		value_objects.BitcoinNetworkMainnet,
		value_objects.BitcoinAddressSchemeNativeSegwit,
		xpub,
		index,
	)
	if err != nil {
		t.Fatalf("DeriveAddress returned error: %v", err)
	}

	want := deriveExpectedAddress(
		t,
		xpub,
		&chaincfg.MainNetParams,
		NewNativeSegwitAddressEncoder(),
		index,
		false,
	)
	if got != want {
		t.Fatalf("unexpected address: got %q, want %q", got, want)
	}
}

func TestHDXPubAddressDeriverAddressTypeByScheme(t *testing.T) {
	tests := []struct {
		name       string
		network    value_objects.BitcoinNetwork
		scheme     value_objects.BitcoinAddressScheme
		params     *chaincfg.Params
		assertType func(t *testing.T, address btcutil.Address)
	}{
		{
			name:    "mainnet legacy p2pkh",
			network: value_objects.BitcoinNetworkMainnet,
			scheme:  value_objects.BitcoinAddressSchemeLegacy,
			params:  &chaincfg.MainNetParams,
			assertType: func(t *testing.T, address btcutil.Address) {
				t.Helper()
				if _, ok := address.(*btcutil.AddressPubKeyHash); !ok {
					t.Fatalf("unexpected address type: %T", address)
				}
			},
		},
		{
			name:    "mainnet segwit nested p2sh-p2wpkh",
			network: value_objects.BitcoinNetworkMainnet,
			scheme:  value_objects.BitcoinAddressSchemeSegwit,
			params:  &chaincfg.MainNetParams,
			assertType: func(t *testing.T, address btcutil.Address) {
				t.Helper()
				if _, ok := address.(*btcutil.AddressScriptHash); !ok {
					t.Fatalf("unexpected address type: %T", address)
				}
			},
		},
		{
			name:    "mainnet native segwit p2wpkh",
			network: value_objects.BitcoinNetworkMainnet,
			scheme:  value_objects.BitcoinAddressSchemeNativeSegwit,
			params:  &chaincfg.MainNetParams,
			assertType: func(t *testing.T, address btcutil.Address) {
				t.Helper()
				if _, ok := address.(*btcutil.AddressWitnessPubKeyHash); !ok {
					t.Fatalf("unexpected address type: %T", address)
				}
			},
		},
		{
			name:    "mainnet taproot p2tr",
			network: value_objects.BitcoinNetworkMainnet,
			scheme:  value_objects.BitcoinAddressSchemeTaproot,
			params:  &chaincfg.MainNetParams,
			assertType: func(t *testing.T, address btcutil.Address) {
				t.Helper()
				if _, ok := address.(*btcutil.AddressTaproot); !ok {
					t.Fatalf("unexpected address type: %T", address)
				}
			},
		},
		{
			name:    "testnet4 legacy p2pkh",
			network: value_objects.BitcoinNetworkTestnet4,
			scheme:  value_objects.BitcoinAddressSchemeLegacy,
			params:  &chaincfg.TestNet4Params,
			assertType: func(t *testing.T, address btcutil.Address) {
				t.Helper()
				if _, ok := address.(*btcutil.AddressPubKeyHash); !ok {
					t.Fatalf("unexpected address type: %T", address)
				}
			},
		},
		{
			name:    "testnet4 segwit nested p2sh-p2wpkh",
			network: value_objects.BitcoinNetworkTestnet4,
			scheme:  value_objects.BitcoinAddressSchemeSegwit,
			params:  &chaincfg.TestNet4Params,
			assertType: func(t *testing.T, address btcutil.Address) {
				t.Helper()
				if _, ok := address.(*btcutil.AddressScriptHash); !ok {
					t.Fatalf("unexpected address type: %T", address)
				}
			},
		},
		{
			name:    "testnet4 native segwit p2wpkh",
			network: value_objects.BitcoinNetworkTestnet4,
			scheme:  value_objects.BitcoinAddressSchemeNativeSegwit,
			params:  &chaincfg.TestNet4Params,
			assertType: func(t *testing.T, address btcutil.Address) {
				t.Helper()
				if _, ok := address.(*btcutil.AddressWitnessPubKeyHash); !ok {
					t.Fatalf("unexpected address type: %T", address)
				}
			},
		},
		{
			name:    "testnet4 taproot p2tr",
			network: value_objects.BitcoinNetworkTestnet4,
			scheme:  value_objects.BitcoinAddressSchemeTaproot,
			params:  &chaincfg.TestNet4Params,
			assertType: func(t *testing.T, address btcutil.Address) {
				t.Helper()
				if _, ok := address.(*btcutil.AddressTaproot); !ok {
					t.Fatalf("unexpected address type: %T", address)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			xpub := newTestXPub(t, tc.params)
			deriver := newTestDeriver()

			encoded, err := deriver.DeriveAddress(tc.network, tc.scheme, xpub, 3)
			if err != nil {
				t.Fatalf("DeriveAddress returned error: %v", err)
			}

			decoded, err := btcutil.DecodeAddress(encoded, tc.params)
			if err != nil {
				t.Fatalf("failed to decode address: %v", err)
			}
			if !decoded.IsForNet(tc.params) {
				t.Fatalf("address is not for network: %q", encoded)
			}

			tc.assertType(t, decoded)
		})
	}
}

func TestHDXPubAddressDeriverInvalidXPub(t *testing.T) {
	deriver := newTestDeriver()

	_, err := deriver.DeriveAddress(
		value_objects.BitcoinNetworkMainnet,
		value_objects.BitcoinAddressSchemeLegacy,
		"not-an-xpub",
		0,
	)
	if err == nil {
		t.Fatal("expected error for invalid xpub")
	}
	if !strings.Contains(err.Error(), "parse xpub") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHDXPubAddressDeriverUnsupportedNetwork(t *testing.T) {
	deriver := newTestDeriver()
	xpub := newTestXPub(t, &chaincfg.MainNetParams)

	_, err := deriver.DeriveAddress(
		value_objects.BitcoinNetwork("regtest"),
		value_objects.BitcoinAddressSchemeLegacy,
		xpub,
		0,
	)
	if err == nil {
		t.Fatal("expected unsupported network error")
	}
	if !strings.Contains(err.Error(), "unsupported bitcoin network") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHDXPubAddressDeriverUnsupportedScheme(t *testing.T) {
	deriver := newTestDeriver()
	xpub := newTestXPub(t, &chaincfg.MainNetParams)

	_, err := deriver.DeriveAddress(
		value_objects.BitcoinNetworkMainnet,
		value_objects.BitcoinAddressScheme("weird"),
		xpub,
		0,
	)
	if err == nil {
		t.Fatal("expected unsupported scheme error")
	}
	if !strings.Contains(err.Error(), "unsupported bitcoin address scheme") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHDXPubAddressDeriverMissingEncoderConfiguration(t *testing.T) {
	deriver := NewHDXPubAddressDeriver()
	xpub := newTestXPub(t, &chaincfg.MainNetParams)

	_, err := deriver.DeriveAddress(
		value_objects.BitcoinNetworkMainnet,
		value_objects.BitcoinAddressSchemeLegacy,
		xpub,
		0,
	)
	if err == nil {
		t.Fatal("expected error when encoder registry is empty")
	}
	if !strings.Contains(err.Error(), "unsupported bitcoin address scheme") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func newTestXPub(t *testing.T, params *chaincfg.Params) string {
	t.Helper()

	seed := bytes.Repeat([]byte{0x24}, 32)
	master, err := hdkeychain.NewMaster(seed, params)
	if err != nil {
		t.Fatalf("failed to create master key: %v", err)
	}

	child, err := master.Derive(0)
	if err != nil {
		t.Fatalf("failed to derive child key: %v", err)
	}

	xpub, err := child.Neuter()
	if err != nil {
		t.Fatalf("failed to neuter key: %v", err)
	}

	if xpub.IsPrivate() {
		t.Fatalf("unexpected private key generated for test xpub: %s", fmt.Sprint(xpub))
	}

	return xpub.String()
}

func newAccountLevelXPub(t *testing.T, params *chaincfg.Params) string {
	t.Helper()

	seed := bytes.Repeat([]byte{0x42}, 32)
	master, err := hdkeychain.NewMaster(seed, params)
	if err != nil {
		t.Fatalf("failed to create master key: %v", err)
	}

	purposeKey, err := master.Derive(hdkeychain.HardenedKeyStart + 84)
	if err != nil {
		t.Fatalf("failed to derive purpose key: %v", err)
	}

	coinType := uint32(0)
	if params.Net != chaincfg.MainNetParams.Net {
		coinType = 1
	}
	coinTypeKey, err := purposeKey.Derive(hdkeychain.HardenedKeyStart + coinType)
	if err != nil {
		t.Fatalf("failed to derive coin type key: %v", err)
	}

	accountKey, err := coinTypeKey.Derive(hdkeychain.HardenedKeyStart)
	if err != nil {
		t.Fatalf("failed to derive account key: %v", err)
	}

	accountXPub, err := accountKey.Neuter()
	if err != nil {
		t.Fatalf("failed to neuter account key: %v", err)
	}

	return accountXPub.String()
}

func newChangeLevelXPub(t *testing.T, params *chaincfg.Params) string {
	t.Helper()

	accountXPub := newAccountLevelXPub(t, params)
	accountKey, err := hdkeychain.NewKeyFromString(accountXPub)
	if err != nil {
		t.Fatalf("failed to parse account xpub: %v", err)
	}

	changeKey, err := accountKey.Derive(0)
	if err != nil {
		t.Fatalf("failed to derive external chain branch: %v", err)
	}

	return changeKey.String()
}

func deriveExpectedAddress(
	t *testing.T,
	xpub string,
	params *chaincfg.Params,
	encoder AddressEncoder,
	index uint32,
	useExternalBranch bool,
) string {
	t.Helper()

	extendedKey, err := hdkeychain.NewKeyFromString(xpub)
	if err != nil {
		t.Fatalf("failed to parse xpub: %v", err)
	}

	derivationKey := extendedKey
	if useExternalBranch {
		derivationKey, err = extendedKey.Derive(0)
		if err != nil {
			t.Fatalf("failed to derive external branch: %v", err)
		}
	}

	childKey, err := derivationKey.Derive(index)
	if err != nil {
		t.Fatalf("failed to derive child index: %v", err)
	}

	publicKey, err := childKey.ECPubKey()
	if err != nil {
		t.Fatalf("failed to extract child public key: %v", err)
	}

	address, err := encoder.Encode(publicKey, params)
	if err != nil {
		t.Fatalf("failed to encode address: %v", err)
	}

	return address.EncodeAddress()
}

func newTestDeriver() *HDXPubAddressDeriver {
	return NewHDXPubAddressDeriver(
		NewLegacyAddressEncoder(),
		NewSegwitAddressEncoder(),
		NewNativeSegwitAddressEncoder(),
		NewTaprootAddressEncoder(),
	)
}
