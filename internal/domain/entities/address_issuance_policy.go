package entities

import (
	"errors"

	"payrune/internal/domain/valueobjects"
)

var (
	ErrAddressPolicyChainMismatch       = errors.New("address policy chain mismatch")
	ErrAddressPolicyNotEnabled          = errors.New("address policy is not enabled")
	ErrAddressPolicyPreviewNotSupported = errors.New("address preview is not supported for this address policy")
	ErrExpectedAmountMinorInvalid       = errors.New("expected amount minor must be greater than zero")
)

type AddressIssuancePolicy struct {
	AddressPolicy  AddressPolicy
	IssuanceConfig valueobjects.AddressIssuanceConfig
}

func (p AddressIssuancePolicy) Normalize() AddressIssuancePolicy {
	p.AddressPolicy = p.AddressPolicy.Normalize()
	p.IssuanceConfig = p.IssuanceConfig.Normalize()
	p.AddressPolicy.Enabled = p.IssuanceConfig.IsEnabled()
	return p
}

func (p AddressIssuancePolicy) IsEnabled() bool {
	return p.Normalize().IssuanceConfig.IsEnabled()
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
	return normalized, nil
}

func (p AddressIssuancePolicy) SupportsAddressPreview() bool {
	normalized := p.Normalize()
	return !(normalized.AddressPolicy.Chain == valueobjects.SupportedChainEthereum &&
		normalized.AddressPolicy.Scheme == "create2")
}

func (p AddressIssuancePolicy) ValidateForAddressPreview(
	requestedChain valueobjects.SupportedChain,
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
	if !normalized.SupportsAddressPreview() {
		return AddressIssuancePolicy{}, ErrAddressPolicyPreviewNotSupported
	}
	return normalized, nil
}
