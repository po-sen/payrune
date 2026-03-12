package entities

import (
	"strings"

	"payrune/internal/domain/valueobjects"
)

type AddressPolicy struct {
	AddressPolicyID string
	Chain           valueobjects.SupportedChain
	Network         valueobjects.NetworkID
	Scheme          string
	MinorUnit       string
	Decimals        uint8
	Enabled         bool
}

func (p AddressPolicy) IsEnabled() bool {
	return p.Enabled
}

func (p AddressPolicy) Normalize() AddressPolicy {
	p.AddressPolicyID = strings.TrimSpace(p.AddressPolicyID)
	p.MinorUnit = strings.TrimSpace(p.MinorUnit)
	if normalizedNetwork, ok := valueobjects.ParseNetworkID(string(p.Network)); ok {
		p.Network = normalizedNetwork
	} else {
		p.Network = valueobjects.NetworkID(strings.ToLower(strings.TrimSpace(string(p.Network))))
	}
	p.Scheme = strings.TrimSpace(p.Scheme)
	return p
}
