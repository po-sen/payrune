package bitcoin

import (
	"fmt"

	"payrune/internal/domain/value_objects"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
)

type AddressEncoder interface {
	Scheme() value_objects.BitcoinAddressScheme
	Encode(publicKey *btcec.PublicKey, params *chaincfg.Params) (btcutil.Address, error)
}

type HDXPubAddressDeriver struct {
	encoders map[value_objects.BitcoinAddressScheme]AddressEncoder
}

func NewHDXPubAddressDeriver(encoders ...AddressEncoder) *HDXPubAddressDeriver {
	registry := make(map[value_objects.BitcoinAddressScheme]AddressEncoder, len(encoders))
	for _, encoder := range encoders {
		registry[encoder.Scheme()] = encoder
	}

	return &HDXPubAddressDeriver{
		encoders: registry,
	}
}

func (d *HDXPubAddressDeriver) DeriveAddress(
	network value_objects.BitcoinNetwork,
	scheme value_objects.BitcoinAddressScheme,
	xpub string,
	index uint32,
) (string, error) {
	params, err := networkParams(network)
	if err != nil {
		return "", err
	}

	extendedKey, err := hdkeychain.NewKeyFromString(xpub)
	if err != nil {
		return "", fmt.Errorf("parse xpub: %w", err)
	}

	childKey, err := deriveAddressExtendedKey(extendedKey, index)
	if err != nil {
		return "", err
	}

	publicKey, err := childKey.ECPubKey()
	if err != nil {
		return "", fmt.Errorf("extract public key: %w", err)
	}

	encoder, ok := d.encoders[scheme]
	if !ok {
		return "", fmt.Errorf("unsupported bitcoin address scheme: %s", scheme)
	}

	address, err := encoder.Encode(publicKey, params)
	if err != nil {
		return "", fmt.Errorf("build address: %w", err)
	}

	return address.EncodeAddress(), nil
}

func (d *HDXPubAddressDeriver) DerivationPath(xpub string, index uint32) (string, error) {
	extendedKey, err := hdkeychain.NewKeyFromString(xpub)
	if err != nil {
		return "", fmt.Errorf("parse xpub: %w", err)
	}

	// Return path relative to account level (m/purpose'/coin_type'/account').
	if extendedKey.Depth() <= 3 {
		return fmt.Sprintf("0/%d", index), nil
	}
	if extendedKey.Depth() == 4 {
		return fmt.Sprintf("%d/%d", extendedKey.ChildIndex(), index), nil
	}

	return fmt.Sprintf("%d", index), nil
}

func deriveAddressExtendedKey(
	extendedKey *hdkeychain.ExtendedKey,
	index uint32,
) (*hdkeychain.ExtendedKey, error) {
	derivationKey := extendedKey

	// Account-level xpubs (depth <= 3) require the external branch first.
	if extendedKey.Depth() <= 3 {
		externalKey, err := extendedKey.Derive(0)
		if err != nil {
			return nil, fmt.Errorf("derive external chain branch: %w", err)
		}
		derivationKey = externalKey
	}

	childKey, err := derivationKey.Derive(index)
	if err != nil {
		return nil, fmt.Errorf("derive child key: %w", err)
	}

	return childKey, nil
}

func networkParams(network value_objects.BitcoinNetwork) (*chaincfg.Params, error) {
	switch network {
	case value_objects.BitcoinNetworkMainnet:
		return &chaincfg.MainNetParams, nil
	case value_objects.BitcoinNetworkTestnet4:
		return &chaincfg.TestNet4Params, nil
	default:
		return nil, fmt.Errorf("unsupported bitcoin network: %s", network)
	}
}
