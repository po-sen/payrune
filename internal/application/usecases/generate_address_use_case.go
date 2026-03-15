package usecases

import (
	"context"
	"errors"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
	outport "payrune/internal/application/ports/outbound"
)

type generateAddressUseCase struct {
	deriver      outport.ChainAddressDeriver
	policyReader outport.AddressPolicyReader
}

func NewGenerateAddressUseCase(
	deriver outport.ChainAddressDeriver,
	policyReader outport.AddressPolicyReader,
) inport.GenerateAddressUseCase {
	return &generateAddressUseCase{
		deriver:      deriver,
		policyReader: policyReader,
	}
}

func (uc *generateAddressUseCase) Execute(
	ctx context.Context,
	input dto.GenerateAddressInput,
) (dto.GenerateAddressResponse, error) {
	if uc.deriver == nil {
		return dto.GenerateAddressResponse{}, errors.New("chain address deriver is not configured")
	}
	if uc.policyReader == nil {
		return dto.GenerateAddressResponse{}, errors.New("address policy reader is not configured")
	}
	if !uc.deriver.SupportsChain(input.Chain) {
		return dto.GenerateAddressResponse{}, inport.ErrChainNotSupported
	}

	policy, ok, err := uc.policyReader.FindIssuanceByID(ctx, input.AddressPolicyID)
	if err != nil {
		return dto.GenerateAddressResponse{}, err
	}
	if !ok || policy.AddressPolicy.Chain != input.Chain {
		return dto.GenerateAddressResponse{}, inport.ErrAddressPolicyNotFound
	}
	if !policy.IsEnabled() {
		return dto.GenerateAddressResponse{}, inport.ErrAddressPolicyNotEnabled
	}

	output, err := uc.deriver.DeriveAddress(ctx, outport.DeriveChainAddressInput{
		Chain:                input.Chain,
		Network:              policy.AddressPolicy.Network,
		Scheme:               policy.AddressPolicy.Scheme,
		AccountPublicKey:     policy.DerivationConfig.AccountPublicKey,
		DerivationPathPrefix: policy.DerivationConfig.DerivationPathPrefix,
		Index:                input.Index,
	})
	if err != nil {
		return dto.GenerateAddressResponse{}, err
	}

	return dto.GenerateAddressResponse{
		AddressPolicyID: policy.AddressPolicy.AddressPolicyID,
		Chain:           string(policy.AddressPolicy.Chain),
		Network:         string(policy.AddressPolicy.Network),
		Scheme:          string(policy.AddressPolicy.Scheme),
		MinorUnit:       policy.AddressPolicy.MinorUnit,
		Decimals:        policy.AddressPolicy.Decimals,
		Index:           input.Index,
		Address:         output.Address,
	}, nil
}
