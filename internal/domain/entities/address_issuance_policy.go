package entities

import (
	"errors"

	"payrune/internal/domain/valueobjects"
)

var (
	ErrAddressPolicyChainMismatch            = errors.New("address policy chain mismatch")
	ErrAddressPolicyNotEnabled               = errors.New("address policy is not enabled")
	ErrAddressPolicyFingerprintNotConfigured = errors.New("address policy fingerprint is not configured")
	ErrExpectedAmountMinorInvalid            = errors.New("expected amount minor must be greater than zero")
)

type AddressIssuancePolicy struct {
	AddressPolicy    AddressPolicy
	DerivationConfig valueobjects.AddressDerivationConfig
}

func (p AddressIssuancePolicy) Normalize() AddressIssuancePolicy {
	p.AddressPolicy = p.AddressPolicy.Normalize()
	p.DerivationConfig = p.DerivationConfig.Normalize()
	p.AddressPolicy.Enabled = p.DerivationConfig.IsEnabled()
	return p
}

func (p AddressIssuancePolicy) IsEnabled() bool {
	return p.Normalize().DerivationConfig.IsEnabled()
}

func (p AddressIssuancePolicy) ValidateForAllocationIssuance(
	requestedChain valueobjects.SupportedChain,
	expectedAmountMinor int64,
) (AddressIssuancePolicy, error) {
	normalized := p.Normalize()
	if normalized.AddressPolicy.AddressPolicyID == "" {
		return AddressIssuancePolicy{}, errors.New("address policy id is required")
	}
	if normalized.AddressPolicy.Chain != requestedChain {
		return AddressIssuancePolicy{}, ErrAddressPolicyChainMismatch
	}
	if !normalized.IsEnabled() {
		return AddressIssuancePolicy{}, ErrAddressPolicyNotEnabled
	}
	if expectedAmountMinor <= 0 {
		return AddressIssuancePolicy{}, ErrExpectedAmountMinorInvalid
	}
	if normalized.DerivationConfig.PublicKeyFingerprintAlgo == "" ||
		normalized.DerivationConfig.PublicKeyFingerprint == "" {
		return AddressIssuancePolicy{}, ErrAddressPolicyFingerprintNotConfigured
	}
	return normalized, nil
}
