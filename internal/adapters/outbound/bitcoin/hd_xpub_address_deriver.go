package bitcoin

import (
	"errors"
	"fmt"
	"strings"

	"payrune/internal/domain/valueobjects"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
)

type AddressEncoder interface {
	Scheme() valueobjects.BitcoinAddressScheme
	Encode(publicKey *btcec.PublicKey, params *chaincfg.Params) (btcutil.Address, error)
}

type HDXPubAddressDeriver struct {
	encoders map[valueobjects.BitcoinAddressScheme]AddressEncoder
}

func NewHDXPubAddressDeriver(encoders ...AddressEncoder) *HDXPubAddressDeriver {
	registry := make(map[valueobjects.BitcoinAddressScheme]AddressEncoder, len(encoders))
	for _, encoder := range encoders {
		registry[encoder.Scheme()] = encoder
	}

	return &HDXPubAddressDeriver{
		encoders: registry,
	}
}

func (d *HDXPubAddressDeriver) DeriveAddress(
	network valueobjects.BitcoinNetwork,
	scheme valueobjects.BitcoinAddressScheme,
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

	return relativeDerivationPath(extendedKey, index), nil
}

func (d *HDXPubAddressDeriver) AbsoluteDerivationPath(
	xpub string,
	derivationPathPrefix string,
	index uint32,
) (string, error) {
	extendedKey, err := hdkeychain.NewKeyFromString(xpub)
	if err != nil {
		return "", fmt.Errorf("parse xpub: %w", err)
	}

	prefix, err := normalizedDerivationPathPrefix(derivationPathPrefix)
	if err != nil {
		return "", err
	}
	if extendedKey.Depth() == 3 {
		prefix, err = replaceAccountPathSegment(prefix, formatDerivationPathIndex(extendedKey.ChildIndex()))
		if err != nil {
			return "", err
		}
	}

	return prefix + "/" + relativeDerivationPath(extendedKey, index), nil
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

func relativeDerivationPath(extendedKey *hdkeychain.ExtendedKey, index uint32) string {
	// Return path relative to account level (m/purpose'/coin_type'/account').
	if extendedKey.Depth() <= 3 {
		return fmt.Sprintf("0/%d", index)
	}
	if extendedKey.Depth() == 4 {
		return fmt.Sprintf("%s/%d", formatDerivationPathIndex(extendedKey.ChildIndex()), index)
	}

	return fmt.Sprintf("%d", index)
}

func normalizedDerivationPathPrefix(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimSuffix(trimmed, "/")
	if trimmed == "" {
		return "", errors.New("derivation path prefix is required")
	}
	if !strings.HasPrefix(trimmed, "m/") {
		return "", errors.New("derivation path prefix is required")
	}
	return trimmed, nil
}

func replaceAccountPathSegment(prefix string, accountSegment string) (string, error) {
	segments := strings.Split(prefix, "/")
	if len(segments) < 4 {
		return "", errors.New("derivation path prefix must include account segment")
	}
	segments[len(segments)-1] = accountSegment
	return strings.Join(segments, "/"), nil
}

func formatDerivationPathIndex(index uint32) string {
	if index >= hdkeychain.HardenedKeyStart {
		return fmt.Sprintf("%d'", index-hdkeychain.HardenedKeyStart)
	}
	return fmt.Sprintf("%d", index)
}

func networkParams(network valueobjects.BitcoinNetwork) (*chaincfg.Params, error) {
	switch network {
	case valueobjects.BitcoinNetworkMainnet:
		return &chaincfg.MainNetParams, nil
	case valueobjects.BitcoinNetworkTestnet4:
		return &chaincfg.TestNet4Params, nil
	default:
		return nil, fmt.Errorf("unsupported bitcoin network: %s", network)
	}
}
