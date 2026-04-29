package policy

import (
	"context"
	"strings"

	outport "payrune/internal/application/ports/outbound"
)

type addressPolicyReader struct {
	ordered      []outport.AddressPolicyRecord
	issuanceByID map[string]outport.AddressIssuancePolicyRecord
}

var _ outport.AddressPolicyReader = (*addressPolicyReader)(nil)

func NewAddressPolicyReader(issuancePolicies []outport.AddressIssuancePolicyRecord) outport.AddressPolicyReader {
	ordered := make([]outport.AddressPolicyRecord, 0, len(issuancePolicies))
	issuanceByID := make(map[string]outport.AddressIssuancePolicyRecord, len(issuancePolicies))

	for _, issuancePolicy := range issuancePolicies {
		addressPolicyID, ok := outport.NormalizeAddressPolicyID(issuancePolicy.AddressPolicyID)
		if !ok {
			continue
		}
		if _, exists := issuanceByID[addressPolicyID]; exists {
			continue
		}
		chain, ok := outport.NormalizeSupportedChain(issuancePolicy.Chain)
		if !ok {
			continue
		}
		network, ok := outport.NormalizeNetworkID(issuancePolicy.Network)
		if !ok {
			continue
		}
		scheme, ok := outport.NormalizeAddressScheme(issuancePolicy.Scheme)
		if !ok {
			continue
		}
		issuancePolicy.AddressPolicyID = addressPolicyID
		issuancePolicy.Chain = chain
		issuancePolicy.Network = network
		issuancePolicy.Scheme = scheme
		issuancePolicy.AssetReference = strings.TrimSpace(issuancePolicy.AssetReference)
		if issuancePolicy.Chain == outport.SupportedChainEthereum && issuancePolicy.AssetReference != "" {
			issuancePolicy.AssetReference = strings.ToLower(issuancePolicy.AssetReference)
		}
		issuancePolicy.AddressSpaceRef = strings.TrimSpace(issuancePolicy.AddressSpaceRef)
		issuancePolicy.IssuanceRefPrefix = strings.TrimSuffix(strings.TrimSpace(issuancePolicy.IssuanceRefPrefix), "/")

		ordered = append(ordered, outport.AddressPolicyRecord{
			AddressPolicyID: issuancePolicy.AddressPolicyID,
			Chain:           issuancePolicy.Chain,
			Network:         issuancePolicy.Network,
			Scheme:          issuancePolicy.Scheme,
			AssetReference:  issuancePolicy.AssetReference,
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
	chain string,
) ([]outport.AddressPolicyRecord, error) {
	normalizedChain, ok := outport.NormalizeSupportedChain(chain)
	if !ok {
		return []outport.AddressPolicyRecord{}, nil
	}
	policies := make([]outport.AddressPolicyRecord, 0)
	for _, policy := range r.ordered {
		if policy.Chain != normalizedChain {
			continue
		}
		policies = append(policies, policy)
	}
	return policies, nil
}

func (r *addressPolicyReader) FindIssuanceByID(
	_ context.Context,
	addressPolicyID string,
) (outport.AddressIssuancePolicyRecord, bool, error) {
	normalizedID, ok := outport.NormalizeAddressPolicyID(addressPolicyID)
	if !ok {
		return outport.AddressIssuancePolicyRecord{}, false, nil
	}
	policy, ok := r.issuanceByID[normalizedID]
	if !ok {
		return outport.AddressIssuancePolicyRecord{}, false, nil
	}
	return policy, true, nil
}
