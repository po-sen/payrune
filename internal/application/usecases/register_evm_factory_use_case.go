package usecases

import (
	"context"
	"errors"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
	outport "payrune/internal/application/ports/outbound"
)

var errActiveEVMFactoryAlreadyExists = errors.New("active evm factory already exists for network")

type registerEVMFactoryUseCase struct {
	unitOfWork outport.UnitOfWork
	clock      outport.Clock
}

func NewRegisterEVMFactoryUseCase(
	unitOfWork outport.UnitOfWork,
	clock outport.Clock,
) inport.RegisterEVMFactoryUseCase {
	return &registerEVMFactoryUseCase{
		unitOfWork: unitOfWork,
		clock:      clock,
	}
}

func (uc *registerEVMFactoryUseCase) Execute(
	ctx context.Context,
	input dto.RegisterEVMFactoryInput,
) (dto.RegisterEVMFactoryResponse, error) {
	if uc.unitOfWork == nil {
		return dto.RegisterEVMFactoryResponse{}, errors.New("unit of work is not configured")
	}
	if uc.clock == nil {
		return dto.RegisterEVMFactoryResponse{}, errors.New("clock is not configured")
	}

	now := uc.clock.NowUTC()
	replaceInput, err := outport.ReplaceActiveEVMFactoryInput{
		Network:               input.Network,
		FactoryAddress:        input.FactoryAddress,
		CollectorAddress:      input.CollectorAddress,
		VaultCreationCodeHash: input.VaultCreationCodeHash,
		DeploymentTxHash:      input.DeploymentTxHash,
		DeployedAt:            input.DeployedAt,
	}.Validate()
	if err != nil {
		return dto.RegisterEVMFactoryResponse{}, err
	}

	var response dto.RegisterEVMFactoryResponse
	if err := uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
		if txScope.EVMFactoryRegistry == nil {
			return errors.New("evm factory registry store is not configured")
		}

		activeRecord, found, err := txScope.EVMFactoryRegistry.FindActiveByNetwork(ctx, replaceInput.Network)
		if err != nil {
			return err
		}
		if found &&
			activeRecord.FactoryAddress != replaceInput.FactoryAddress &&
			!input.AllowReplaceActive {
			return errActiveEVMFactoryAlreadyExists
		}

		record, err := txScope.EVMFactoryRegistry.ReplaceActive(ctx, replaceInput, now)
		if err != nil {
			return err
		}

		response = dto.RegisterEVMFactoryResponse{
			ID:                    record.ID,
			Network:               string(record.Network),
			FactoryAddress:        record.FactoryAddress,
			CollectorAddress:      record.CollectorAddress,
			VaultCreationCodeHash: record.VaultCreationCodeHash,
			Status:                string(record.Status),
			DeploymentTxHash:      record.DeploymentTxHash,
			DeployedAt:            record.DeployedAt,
		}
		return nil
	}); err != nil {
		return dto.RegisterEVMFactoryResponse{}, err
	}

	return response, nil
}
