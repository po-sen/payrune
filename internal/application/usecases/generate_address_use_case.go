package usecases

import (
	"context"
	"errors"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
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

	addressPolicyID, err := valueobjects.NewAddressPolicyID(input.AddressPolicyID)
	if err != nil {
		return dto.GenerateAddressResponse{}, inport.ErrInvalidAddressPolicyID
	}

	policy, ok, err := uc.policyReader.FindIssuanceByID(ctx, addressPolicyID)
	if err != nil {
		return dto.GenerateAddressResponse{}, inport.ErrDependencyFailure
	}
	if !ok {
		return dto.GenerateAddressResponse{}, inport.ErrAddressPolicyNotFound
	}
	policy, err = policy.ValidateForAddressPreview(input.Chain)
	if err != nil {
		switch {
		case errors.Is(err, policies.ErrAddressPolicyChainMismatch):
			return dto.GenerateAddressResponse{}, inport.ErrAddressPolicyNotFound
		case errors.Is(err, policies.ErrAddressPolicyNotEnabled):
			return dto.GenerateAddressResponse{}, inport.ErrAddressPolicyNotEnabled
		case errors.Is(err, policies.ErrAddressPolicyPreviewNotSupported):
			return dto.GenerateAddressResponse{}, inport.ErrAddressPreviewNotSupported
		default:
			return dto.GenerateAddressResponse{}, inport.ErrInternalFailure
		}
	}

	output, err := uc.deriver.DeriveAddress(ctx, outport.DeriveChainAddressInput{
		Chain:             input.Chain,
		Network:           policy.Network,
		Scheme:            policy.Scheme,
		AddressSpaceRef:   policy.IssuanceConfig.AddressSpaceRef,
		IssuanceRefPrefix: policy.IssuanceConfig.IssuanceRefPrefix,
		SlotIndex:         input.Index,
	})
	if err != nil {
		return dto.GenerateAddressResponse{}, inport.ErrDependencyFailure
	}

	return dto.GenerateAddressResponse{
		AddressPolicyID: string(policy.AddressPolicyID),
		Chain:           string(policy.Chain),
		Network:         string(policy.Network),
		Scheme:          string(policy.Scheme),
		MinorUnit:       policy.MinorUnit,
		Decimals:        policy.Decimals,
		Index:           input.Index,
		Address:         output.Address,
	}, nil
}
