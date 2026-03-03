package bitcoin

import (
	"payrune/internal/domain/value_objects"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

type taprootAddressEncoder struct{}

func NewTaprootAddressEncoder() AddressEncoder {
	return taprootAddressEncoder{}
}

func (taprootAddressEncoder) Scheme() value_objects.BitcoinAddressScheme {
	return value_objects.BitcoinAddressSchemeTaproot
}

func (taprootAddressEncoder) Encode(
	publicKey *btcec.PublicKey,
	params *chaincfg.Params,
) (btcutil.Address, error) {
	taprootKey := txscript.ComputeTaprootKeyNoScript(publicKey)
	return btcutil.NewAddressTaproot(schnorr.SerializePubKey(taprootKey), params)
}
