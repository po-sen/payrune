package bitcoin

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

type legacyAddressEncoder struct{}

func NewLegacyAddressEncoder() addressEncoder {
	return legacyAddressEncoder{}
}

func (legacyAddressEncoder) Scheme() addressScheme {
	return addressSchemeLegacy
}

func (legacyAddressEncoder) Encode(
	publicKey *btcec.PublicKey,
	params *chaincfg.Params,
) (btcutil.Address, error) {
	pubKeyHash := btcutil.Hash160(publicKey.SerializeCompressed())
	return btcutil.NewAddressPubKeyHash(pubKeyHash, params)
}
