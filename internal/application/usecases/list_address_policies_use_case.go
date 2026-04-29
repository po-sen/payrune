package usecases

import (
	"context"
	"strings"

	inport "payrune/internal/application/ports/inbound"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

type listAddressPoliciesUseCase struct {
	policyReader outport.AddressPolicyReader
}

func NewListAddressPoliciesUseCase(policyReader outport.AddressPolicyReader) inport.ListAddressPoliciesUseCase {
	return &listAddressPoliciesUseCase{policyReader: policyReader}
}

func (uc *listAddressPoliciesUseCase) Execute(
	ctx context.Context,
	chain string,
) (inport.ListAddressPoliciesResponse, error) {
	if uc.policyReader == nil {
		return inport.ListAddressPoliciesResponse{}, inport.ErrAddressPolicyReaderNotConfigured
	}
	normalizedChain, ok := valueobjects.ParseSupportedChain(chain)
	if !ok {
		return inport.ListAddressPoliciesResponse{}, inport.ErrChainNotSupported
	}

	policyRecords, err := uc.policyReader.ListByChain(ctx, string(normalizedChain))
	if err != nil {
		return inport.ListAddressPoliciesResponse{}, inport.ErrDependencyFailure
	}

	policies := make([]inport.AddressPolicy, 0)
	for _, policy := range policyRecords {
		policies = append(policies, inport.AddressPolicy{
			AddressPolicyID: policy.AddressPolicyID,
			Chain:           policy.Chain,
			Network:         policy.Network,
			Scheme:          policy.Scheme,
			AssetReference:  strings.TrimSpace(policy.AssetReference),
			Decimals:        policy.Decimals,
			Enabled:         policy.Enabled,
		})
	}

	return inport.ListAddressPoliciesResponse{
		Chain:           string(normalizedChain),
		AddressPolicies: policies,
	}, nil
}
