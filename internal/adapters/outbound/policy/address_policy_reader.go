package policy

import (
	"context"
	"strings"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/valueobjects"
)

type AddressPolicyConfig struct {
	AddressPolicyID        string
	Chain                  valueobjects.SupportedChain
	Network                valueobjects.NetworkID
	Scheme                 string
	MinorUnit              string
	Decimals               uint8
	AddressSourceRef       string
	AddressReferencePrefix string
}

type addressPolicyReader struct {
	ordered      []entities.AddressPolicy
	issuanceByID map[string]entities.AddressIssuancePolicy
}

var _ outport.AddressPolicyReader = (*addressPolicyReader)(nil)

func NewAddressPolicyReader(configs []AddressPolicyConfig) outport.AddressPolicyReader {
	ordered := make([]entities.AddressPolicy, 0, len(configs))
	issuanceByID := make(map[string]entities.AddressIssuancePolicy, len(configs))

	for _, cfg := range configs {
		issuancePolicy := entities.AddressIssuancePolicy{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: strings.TrimSpace(cfg.AddressPolicyID),
				Chain:           cfg.Chain,
				Network:         cfg.Network,
				Scheme:          cfg.Scheme,
				MinorUnit:       strings.TrimSpace(cfg.MinorUnit),
				Decimals:        cfg.Decimals,
			},
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSourceRef:       strings.TrimSpace(cfg.AddressSourceRef),
				AddressReferencePrefix: strings.TrimSpace(cfg.AddressReferencePrefix),
			},
		}.Normalize()

		if issuancePolicy.AddressPolicy.AddressPolicyID == "" {
			continue
		}
		if _, exists := issuanceByID[issuancePolicy.AddressPolicy.AddressPolicyID]; exists {
			continue
		}

		ordered = append(ordered, issuancePolicy.AddressPolicy)
		issuanceByID[issuancePolicy.AddressPolicy.AddressPolicyID] = issuancePolicy
	}

	return &addressPolicyReader{
		ordered:      ordered,
		issuanceByID: issuanceByID,
	}
}

func (r *addressPolicyReader) ListByChain(
	_ context.Context,
	chain valueobjects.SupportedChain,
) ([]entities.AddressPolicy, error) {
	policies := make([]entities.AddressPolicy, 0)
	for _, policy := range r.ordered {
		if policy.Chain != chain {
			continue
		}
		policies = append(policies, policy)
	}
	return policies, nil
}

func (r *addressPolicyReader) FindIssuanceByID(
	_ context.Context,
	addressPolicyID string,
) (entities.AddressIssuancePolicy, bool, error) {
	policy, ok := r.issuanceByID[strings.TrimSpace(addressPolicyID)]
	if !ok {
		return entities.AddressIssuancePolicy{}, false, nil
	}
	return policy, true, nil
}
