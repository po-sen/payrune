package bitcoin

import (
	"testing"

	"payrune/internal/domain/valueobjects"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
)

type providedAddressVector struct {
	name     string
	network  valueobjects.BitcoinNetwork
	scheme   valueobjects.BitcoinAddressScheme
	xpub     string
	expected string
}

func newEncoderTestPublicKey(t *testing.T, params *chaincfg.Params, index uint32) *btcec.PublicKey {
	t.Helper()

	xpub := newTestXPub(t, params)
	extendedKey, err := hdkeychain.NewKeyFromString(xpub)
	if err != nil {
		t.Fatalf("failed to parse test xpub: %v", err)
	}

	childKey, err := extendedKey.Derive(index)
	if err != nil {
		t.Fatalf("failed to derive child key: %v", err)
	}

	publicKey, err := childKey.ECPubKey()
	if err != nil {
		t.Fatalf("failed to get child public key: %v", err)
	}

	return publicKey
}

func newVectorTestDeriver() *HDXPubAddressDeriver {
	return NewHDXPubAddressDeriver(
		NewLegacyAddressEncoder(),
		NewSegwitAddressEncoder(),
		NewNativeSegwitAddressEncoder(),
		NewTaprootAddressEncoder(),
	)
}

func assertProvidedVector(t *testing.T, deriver *HDXPubAddressDeriver, tc providedAddressVector) {
	t.Helper()

	if tc.xpub == "" {
		t.Fatalf("%s fixture missing: xpub", tc.name)
	}
	if tc.expected == "" {
		t.Fatalf("%s fixture missing: expected address", tc.name)
	}

	got, err := deriver.DeriveAddress(tc.network, tc.scheme, tc.xpub, 0)
	if err != nil {
		t.Fatalf("%s derive failed: %v", tc.name, err)
	}

	if got != tc.expected {
		t.Fatalf("%s mismatch: expected=%s got=%s", tc.name, tc.expected, got)
	}
}
