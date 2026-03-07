package entities

import (
	"strings"

	"payrune/internal/domain/value_objects"
)

type AddressPolicy struct {
	AddressPolicyID string
	Chain           value_objects.SupportedChain
	Network         value_objects.NetworkID
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
	if normalizedNetwork, ok := value_objects.ParseNetworkID(string(p.Network)); ok {
		p.Network = normalizedNetwork
	} else {
		p.Network = value_objects.NetworkID(strings.ToLower(strings.TrimSpace(string(p.Network))))
	}
	p.Scheme = strings.TrimSpace(p.Scheme)
	return p
}
