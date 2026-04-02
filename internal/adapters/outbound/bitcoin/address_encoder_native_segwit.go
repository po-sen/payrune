package bitcoin

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

type nativeSegwitAddressEncoder struct{}

func NewNativeSegwitAddressEncoder() addressEncoder {
	return nativeSegwitAddressEncoder{}
}

func (nativeSegwitAddressEncoder) Scheme() addressScheme {
	return addressSchemeNativeSegwit
}

func (nativeSegwitAddressEncoder) Encode(
	publicKey *btcec.PublicKey,
	params *chaincfg.Params,
) (btcutil.Address, error) {
	pubKeyHash := btcutil.Hash160(publicKey.SerializeCompressed())
	return btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, params)
}
