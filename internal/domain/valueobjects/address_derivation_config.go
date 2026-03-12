package valueobjects

import (
	"errors"
	"strings"
)

type AddressDerivationConfig struct {
	AccountPublicKey         string
	PublicKeyFingerprintAlgo string
	PublicKeyFingerprint     string
	DerivationPathPrefix     string
}

func (c AddressDerivationConfig) Normalize() AddressDerivationConfig {
	c.AccountPublicKey = strings.TrimSpace(c.AccountPublicKey)
	c.PublicKeyFingerprintAlgo = strings.TrimSpace(c.PublicKeyFingerprintAlgo)
	c.PublicKeyFingerprint = strings.TrimSpace(c.PublicKeyFingerprint)
	c.DerivationPathPrefix = normalizeDerivationPathPrefix(c.DerivationPathPrefix)
	return c
}

func (c AddressDerivationConfig) IsEnabled() bool {
	return strings.TrimSpace(c.AccountPublicKey) != ""
}

func (c AddressDerivationConfig) AbsoluteDerivationPath(relative string) (string, error) {
	normalizedRelative := strings.TrimSpace(relative)
	if normalizedRelative == "" {
		return "", errors.New("derivation path is required")
	}
	if strings.HasPrefix(normalizedRelative, "m/") {
		return normalizedRelative, nil
	}

	prefix := normalizeDerivationPathPrefix(c.DerivationPathPrefix)
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
