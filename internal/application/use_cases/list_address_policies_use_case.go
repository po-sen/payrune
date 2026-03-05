package use_cases

import (
	"context"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/in"
	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/value_objects"
)

type listAddressPoliciesUseCase struct {
	policyReader outport.AddressPolicyReader
}

func NewListAddressPoliciesUseCase(policyReader outport.AddressPolicyReader) inport.ListAddressPoliciesUseCase {
	return &listAddressPoliciesUseCase{policyReader: policyReader}
}

func (uc *listAddressPoliciesUseCase) Execute(
	ctx context.Context,
	chain value_objects.Chain,
) (dto.ListAddressPoliciesResponse, error) {
	if chain != value_objects.ChainBitcoin {
		return dto.ListAddressPoliciesResponse{}, inport.ErrChainNotSupported
	}

	policyEntities, err := uc.policyReader.ListByChain(ctx, chain)
	if err != nil {
		return dto.ListAddressPoliciesResponse{}, err
	}

	policies := make([]dto.AddressPolicy, 0)
	for _, policy := range policyEntities {
		policies = append(policies, dto.AddressPolicy{
			AddressPolicyID: policy.AddressPolicyID,
			Chain:           string(policy.Chain),
			Network:         string(policy.Network),
			Scheme:          string(policy.Scheme),
			MinorUnit:       policy.MinorUnit,
			Decimals:        policy.Decimals,
			Enabled:         policy.IsEnabled(),
		})
	}

	return dto.ListAddressPoliciesResponse{
		Chain:           string(chain),
		AddressPolicies: policies,
	}, nil
}
