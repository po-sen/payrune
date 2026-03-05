package entities

import (
	"errors"
	"strings"

	"payrune/internal/domain/value_objects"
)

type AddressPolicy struct {
	AddressPolicyID      string
	Chain                value_objects.Chain
	Network              value_objects.BitcoinNetwork
	Scheme               value_objects.BitcoinAddressScheme
	MinorUnit            string
	Decimals             uint8
	XPub                 string
	XPubFingerprintAlgo  string
	XPubFingerprint      string
	DerivationPathPrefix string
}

func (p AddressPolicy) IsEnabled() bool {
	return strings.TrimSpace(p.XPub) != ""
}

func (p AddressPolicy) Normalize() AddressPolicy {
	p.AddressPolicyID = strings.TrimSpace(p.AddressPolicyID)
	p.MinorUnit = strings.TrimSpace(p.MinorUnit)
	p.XPub = strings.TrimSpace(p.XPub)
	p.XPubFingerprintAlgo = strings.TrimSpace(p.XPubFingerprintAlgo)
	p.XPubFingerprint = strings.TrimSpace(p.XPubFingerprint)
	p.DerivationPathPrefix = normalizeDerivationPathPrefix(p.DerivationPathPrefix)
	return p
}

func (p AddressPolicy) AbsoluteDerivationPath(relative string) (string, error) {
	normalizedRelative := strings.TrimSpace(relative)
	if normalizedRelative == "" {
		return "", errors.New("derivation path is required")
	}
	if strings.HasPrefix(normalizedRelative, "m/") {
		return normalizedRelative, nil
	}

	prefix := normalizeDerivationPathPrefix(p.DerivationPathPrefix)
	if prefix == "" {
		return "", errors.New("derivation path prefix is required")
	}

	normalizedRelative = strings.TrimPrefix(normalizedRelative, "/")
	return prefix + "/" + normalizedRelative, nil
}

func normalizeDerivationPathPrefix(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimSuffix(trimmed, "/")
	if trimmed == "" {
		return ""
	}
	if !strings.HasPrefix(trimmed, "m/") {
		return ""
	}
	return trimmed
}
