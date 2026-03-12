package bitcoin

import (
	"payrune/internal/domain/valueobjects"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

type legacyAddressEncoder struct{}

func NewLegacyAddressEncoder() AddressEncoder {
	return legacyAddressEncoder{}
}

func (legacyAddressEncoder) Scheme() valueobjects.BitcoinAddressScheme {
	return valueobjects.BitcoinAddressSchemeLegacy
}

func (legacyAddressEncoder) Encode(
	publicKey *btcec.PublicKey,
	params *chaincfg.Params,
) (btcutil.Address, error) {
	pubKeyHash := btcutil.Hash160(publicKey.SerializeCompressed())
	return btcutil.NewAddressPubKeyHash(pubKeyHash, params)
}
