package usecases

import (
	"context"
	"errors"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
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
		return dto.GenerateAddressResponse{}, inport.ErrChainAddressDeriverNotConfigured
	}
	if uc.policyReader == nil {
		return dto.GenerateAddressResponse{}, inport.ErrAddressPolicyReaderNotConfigured
	}
	if !uc.deriver.SupportsChain(input.Chain) {
		return dto.GenerateAddressResponse{}, inport.ErrChainNotSupported
	}

	policy, ok, err := uc.policyReader.FindIssuanceByID(ctx, input.AddressPolicyID)
	if err != nil {
		return dto.GenerateAddressResponse{}, inport.ErrDependencyFailure
	}
	if !ok {
		return dto.GenerateAddressResponse{}, inport.ErrAddressPolicyNotFound
	}
	policy, err = policy.ValidateForAddressPreview(input.Chain)
	if err != nil {
		switch {
		case errors.Is(err, entities.ErrAddressPolicyChainMismatch):
			return dto.GenerateAddressResponse{}, inport.ErrAddressPolicyNotFound
		case errors.Is(err, entities.ErrAddressPolicyNotEnabled):
			return dto.GenerateAddressResponse{}, inport.ErrAddressPolicyNotEnabled
		case errors.Is(err, entities.ErrAddressPolicyPreviewNotSupported):
			return dto.GenerateAddressResponse{}, inport.ErrAddressPreviewNotSupported
		default:
			return dto.GenerateAddressResponse{}, inport.ErrInternalFailure
		}
	}

	output, err := uc.deriver.DeriveAddress(ctx, outport.DeriveChainAddressInput{
		Chain:             input.Chain,
		Network:           policy.AddressPolicy.Network,
		Scheme:            policy.AddressPolicy.Scheme,
		AddressSpaceRef:   policy.IssuanceConfig.AddressSpaceRef,
		IssuanceRefPrefix: policy.IssuanceConfig.IssuanceRefPrefix,
		SlotIndex:         input.Index,
	})
	if err != nil {
		return dto.GenerateAddressResponse{}, inport.ErrDependencyFailure
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
