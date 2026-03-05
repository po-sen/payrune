package use_cases

import (
	"context"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/in"
	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/value_objects"
)

type generateAddressUseCase struct {
	deriver      outport.BitcoinAddressDeriver
	policyReader outport.AddressPolicyReader
}

func NewGenerateAddressUseCase(
	deriver outport.BitcoinAddressDeriver,
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
	if input.Chain != value_objects.ChainBitcoin {
		return dto.GenerateAddressResponse{}, inport.ErrChainNotSupported
	}

	policy, ok, err := uc.policyReader.FindByID(ctx, input.AddressPolicyID)
	if err != nil {
		return dto.GenerateAddressResponse{}, err
	}
	if !ok || policy.Chain != input.Chain {
		return dto.GenerateAddressResponse{}, inport.ErrAddressPolicyNotFound
	}
	if !policy.IsEnabled() {
		return dto.GenerateAddressResponse{}, inport.ErrAddressPolicyNotEnabled
	}

	address, err := uc.deriver.DeriveAddress(policy.Network, policy.Scheme, policy.XPub, input.Index)
	if err != nil {
		return dto.GenerateAddressResponse{}, err
	}

	return dto.GenerateAddressResponse{
		AddressPolicyID: policy.AddressPolicyID,
		Chain:           string(policy.Chain),
		Network:         string(policy.Network),
		Scheme:          string(policy.Scheme),
		MinorUnit:       policy.MinorUnit,
		Decimals:        policy.Decimals,
		Index:           input.Index,
		Address:         address,
	}, nil
}
