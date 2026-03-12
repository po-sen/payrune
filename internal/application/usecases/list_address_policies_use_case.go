package usecases

import (
	"context"
	"errors"

	"payrune/internal/application/dto"
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
	chain valueobjects.SupportedChain,
) (dto.ListAddressPoliciesResponse, error) {
	if uc.policyReader == nil {
		return dto.ListAddressPoliciesResponse{}, errors.New("address policy reader is not configured")
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
