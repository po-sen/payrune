package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"payrune/internal/application/dto"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

func TestRegisterEVMFactoryUseCaseSuccess(t *testing.T) {
	factoryStore := &fakeEVMFactoryStore{
		record: outport.EVMFactoryRecord{
			ID:                    17,
			Network:               valueobjects.NetworkID("sepolia"),
			FactoryAddress:        "0x1111111111111111111111111111111111111111",
			CollectorAddress:      "0x2222222222222222222222222222222222222222",
			VaultCreationCodeHash: "0x1234",
			Status:                outport.EVMFactoryStatusActive,
			DeploymentTxHash:      "0xabc",
			DeployedAt:            time.Date(2026, 3, 16, 4, 0, 0, 0, time.UTC),
		},
	}
	uow := &fakeUnitOfWork{evmFactoryStore: factoryStore}
	useCase := NewRegisterEVMFactoryUseCase(uow, fixedClock(time.Date(2026, 3, 16, 5, 0, 0, 0, time.UTC)))

	response, err := useCase.Execute(context.Background(), dto.RegisterEVMFactoryInput{
		Network:               valueobjects.NetworkID("sepolia"),
		FactoryAddress:        "0x1111111111111111111111111111111111111111",
		CollectorAddress:      "0x2222222222222222222222222222222222222222",
		VaultCreationCodeHash: "0x1234",
		DeploymentTxHash:      "0xabc",
		DeployedAt:            time.Date(2026, 3, 16, 4, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if response.ID != 17 {
		t.Fatalf("unexpected id: got %d", response.ID)
	}
	if response.Network != "sepolia" {
		t.Fatalf("unexpected network: got %q", response.Network)
	}
	if response.VaultCreationCodeHash != "0x1234" {
		t.Fatalf("unexpected vault creation code hash: got %q", response.VaultCreationCodeHash)
	}
	if factoryStore.replaceCalls != 1 {
		t.Fatalf("expected replace to be called once, got %d", factoryStore.replaceCalls)
	}
}

func TestRegisterEVMFactoryUseCaseValidationError(t *testing.T) {
	useCase := NewRegisterEVMFactoryUseCase(
		&fakeUnitOfWork{evmFactoryStore: &fakeEVMFactoryStore{}},
		fixedClock(time.Now().UTC()),
	)

	_, err := useCase.Execute(context.Background(), dto.RegisterEVMFactoryInput{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestRegisterEVMFactoryUseCaseRejectsUnexpectedReplace(t *testing.T) {
	factoryStore := &fakeEVMFactoryStore{
		record: outport.EVMFactoryRecord{
			ID:             9,
			Network:        valueobjects.NetworkID("mainnet"),
			FactoryAddress: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Status:         outport.EVMFactoryStatusActive,
		},
		found: true,
	}
	useCase := NewRegisterEVMFactoryUseCase(
		&fakeUnitOfWork{evmFactoryStore: factoryStore},
		fixedClock(time.Now().UTC()),
	)

	_, err := useCase.Execute(context.Background(), dto.RegisterEVMFactoryInput{
		Network:               valueobjects.NetworkID("mainnet"),
		FactoryAddress:        "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		CollectorAddress:      "0x2222222222222222222222222222222222222222",
		VaultCreationCodeHash: "0x1234",
	})
	if !errors.Is(err, errActiveEVMFactoryAlreadyExists) {
		t.Fatalf("expected %v, got %v", errActiveEVMFactoryAlreadyExists, err)
	}
}

func TestRegisterEVMFactoryUseCasePropagatesStoreError(t *testing.T) {
	expectedErr := errors.New("replace failed")
	factoryStore := &fakeEVMFactoryStore{replaceErr: expectedErr}
	useCase := NewRegisterEVMFactoryUseCase(
		&fakeUnitOfWork{evmFactoryStore: factoryStore},
		fixedClock(time.Now().UTC()),
	)

	_, err := useCase.Execute(context.Background(), dto.RegisterEVMFactoryInput{
		Network:               valueobjects.NetworkID("mainnet"),
		FactoryAddress:        "0x1111111111111111111111111111111111111111",
		CollectorAddress:      "0x2222222222222222222222222222222222222222",
		VaultCreationCodeHash: "0x1234",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

type fixedClock time.Time

func (c fixedClock) NowUTC() time.Time {
	return time.Time(c).UTC()
}
