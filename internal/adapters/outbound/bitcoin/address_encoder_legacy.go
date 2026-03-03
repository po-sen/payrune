package bitcoin

import (
	"payrune/internal/domain/value_objects"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

type legacyAddressEncoder struct{}

func NewLegacyAddressEncoder() AddressEncoder {
	return legacyAddressEncoder{}
}

func (legacyAddressEncoder) Scheme() value_objects.BitcoinAddressScheme {
	return value_objects.BitcoinAddressSchemeLegacy
}

func (legacyAddressEncoder) Encode(
	publicKey *btcec.PublicKey,
	params *chaincfg.Params,
) (btcutil.Address, error) {
	pubKeyHash := btcutil.Hash160(publicKey.SerializeCompressed())
	return btcutil.NewAddressPubKeyHash(pubKeyHash, params)
}
