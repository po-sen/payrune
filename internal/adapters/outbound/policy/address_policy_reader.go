package policy

import (
	"context"
	"strings"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/valueobjects"
)

type addressPolicyReader struct {
	ordered      []entities.AddressPolicy
	issuanceByID map[string]entities.AddressIssuancePolicy
}

var _ outport.AddressPolicyReader = (*addressPolicyReader)(nil)

func NewAddressPolicyReader(policies []entities.AddressIssuancePolicy) outport.AddressPolicyReader {
	ordered := make([]entities.AddressPolicy, 0, len(policies))
	issuanceByID := make(map[string]entities.AddressIssuancePolicy, len(policies))

	for _, issuancePolicy := range policies {
		issuancePolicy = issuancePolicy.Normalize()
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
