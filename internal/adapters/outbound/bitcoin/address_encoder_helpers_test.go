package bitcoin

import (
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
)

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
