package bitcoin

import (
	"payrune/internal/domain/value_objects"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

type nativeSegwitAddressEncoder struct{}

func NewNativeSegwitAddressEncoder() AddressEncoder {
	return nativeSegwitAddressEncoder{}
}

func (nativeSegwitAddressEncoder) Scheme() value_objects.BitcoinAddressScheme {
	return value_objects.BitcoinAddressSchemeNativeSegwit
}

func (nativeSegwitAddressEncoder) Encode(
	publicKey *btcec.PublicKey,
	params *chaincfg.Params,
) (btcutil.Address, error) {
	pubKeyHash := btcutil.Hash160(publicKey.SerializeCompressed())
	return btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, params)
}
