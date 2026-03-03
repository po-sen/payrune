package bitcoin

import (
	"payrune/internal/domain/value_objects"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

type segwitAddressEncoder struct{}

func NewSegwitAddressEncoder() AddressEncoder {
	return segwitAddressEncoder{}
}

func (segwitAddressEncoder) Scheme() value_objects.BitcoinAddressScheme {
	return value_objects.BitcoinAddressSchemeSegwit
}

func (segwitAddressEncoder) Encode(
	publicKey *btcec.PublicKey,
	params *chaincfg.Params,
) (btcutil.Address, error) {
	pubKeyHash := btcutil.Hash160(publicKey.SerializeCompressed())

	nativeSegwitAddr, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, params)
	if err != nil {
		return nil, err
	}

	redeemScript, err := txscript.PayToAddrScript(nativeSegwitAddr)
	if err != nil {
		return nil, err
	}

	return btcutil.NewAddressScriptHash(redeemScript, params)
}
