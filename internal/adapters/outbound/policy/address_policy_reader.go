package policy

import (
	"context"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
)

type addressPolicyReader struct {
	ordered      []outport.AddressPolicyRecord
	issuanceByID map[valueobjects.AddressPolicyID]policies.AddressIssuancePolicy
}

var _ outport.AddressPolicyReader = (*addressPolicyReader)(nil)

func NewAddressPolicyReader(issuancePolicies []policies.AddressIssuancePolicy) outport.AddressPolicyReader {
	ordered := make([]outport.AddressPolicyRecord, 0, len(issuancePolicies))
	issuanceByID := make(map[valueobjects.AddressPolicyID]policies.AddressIssuancePolicy, len(issuancePolicies))

	for _, issuancePolicy := range issuancePolicies {
		issuancePolicy = issuancePolicy.Normalize()
		if issuancePolicy.AddressPolicyID.IsZero() {
			continue
		}
		if _, exists := issuanceByID[issuancePolicy.AddressPolicyID]; exists {
			continue
		}

		ordered = append(ordered, outport.AddressPolicyRecord{
			AddressPolicyID: issuancePolicy.AddressPolicyID,
			Chain:           issuancePolicy.Chain,
			Network:         issuancePolicy.Network,
			Scheme:          issuancePolicy.Scheme,
			MinorUnit:       issuancePolicy.MinorUnit,
			Decimals:        issuancePolicy.Decimals,
			Enabled:         issuancePolicy.Enabled,
		})
		issuanceByID[issuancePolicy.AddressPolicyID] = issuancePolicy
	}

	return &addressPolicyReader{
		ordered:      ordered,
		issuanceByID: issuanceByID,
	}
}

func (r *addressPolicyReader) ListByChain(
	_ context.Context,
	chain valueobjects.SupportedChain,
) ([]outport.AddressPolicyRecord, error) {
	policies := make([]outport.AddressPolicyRecord, 0)
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
	addressPolicyID valueobjects.AddressPolicyID,
) (policies.AddressIssuancePolicy, bool, error) {
	policy, ok := r.issuanceByID[addressPolicyID]
	if !ok {
		return policies.AddressIssuancePolicy{}, false, nil
	}
	return policy, true, nil
}
