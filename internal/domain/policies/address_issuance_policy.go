package policies

import (
	"strings"

	"payrune/internal/domain/valueobjects"
)

type AddressIssuancePolicy struct {
	AddressPolicyID valueobjects.AddressPolicyID
	Chain           valueobjects.SupportedChain
	Network         valueobjects.NetworkID
	Scheme          valueobjects.AddressScheme
	MinorUnit       string
	Decimals        uint8
	Enabled         bool
	IssuanceConfig  valueobjects.AddressIssuanceConfig
}

func (p AddressIssuancePolicy) Normalize() AddressIssuancePolicy {
	p.AddressPolicyID = p.AddressPolicyID.Normalize()
	p.IssuanceConfig = p.IssuanceConfig.Normalize()
	if normalizedChain, ok := valueobjects.ParseSupportedChain(string(p.Chain)); ok {
		p.Chain = normalizedChain
	}
	if normalizedNetwork, ok := valueobjects.ParseNetworkID(string(p.Network)); ok {
		p.Network = normalizedNetwork
	} else {
		p.Network = valueobjects.NetworkID(strings.ToLower(strings.TrimSpace(string(p.Network))))
	}
	p.Scheme = p.Scheme.Normalize()
	p.MinorUnit = strings.TrimSpace(p.MinorUnit)
	p.Enabled = p.IssuanceConfig.IsEnabled()
	return p
}

func (p AddressIssuancePolicy) IsEnabled() bool {
	return p.Normalize().Enabled
}

func (p AddressIssuancePolicy) ValidateForAllocationIssuance(
	requestedChain valueobjects.SupportedChain,
	expectedAmountMinor int64,
) (AddressIssuancePolicy, error) {
	normalized := p.Normalize()
	if normalized.AddressPolicyID.IsZero() {
		return AddressIssuancePolicy{}, ErrAddressPolicyIDRequired
	}
	if normalized.Chain != requestedChain {
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
	return !(normalized.Chain == valueobjects.SupportedChainEthereum && normalized.Scheme.IsCreate2())
}

func (p AddressIssuancePolicy) ValidateForAddressPreview(
	requestedChain valueobjects.SupportedChain,
) (AddressIssuancePolicy, error) {
	normalized := p.Normalize()
	if normalized.AddressPolicyID.IsZero() {
		return AddressIssuancePolicy{}, ErrAddressPolicyIDRequired
	}
	if normalized.Chain != requestedChain {
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
