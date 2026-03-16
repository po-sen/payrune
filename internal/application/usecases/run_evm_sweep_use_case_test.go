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

type fakeSweepVaultStore struct {
	candidates      []outport.EVMSweepCandidateRecord
	findErr         error
	submittedInputs []outport.MarkEVMSweepSubmittedInput
	succeededInputs []outport.MarkEVMSweepResultInput
	failedInputs    []outport.MarkEVMSweepResultInput
}

func (f *fakeSweepVaultStore) Create(context.Context, outport.CreateEVMPaymentVaultInput) (outport.EVMPaymentVaultRecord, error) {
	return outport.EVMPaymentVaultRecord{}, nil
}

func (f *fakeSweepVaultStore) FindSweepCandidates(
	_ context.Context,
	_ outport.FindEVMSweepCandidatesInput,
) ([]outport.EVMSweepCandidateRecord, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	rows := make([]outport.EVMSweepCandidateRecord, len(f.candidates))
	copy(rows, f.candidates)
	return rows, nil
}

func (f *fakeSweepVaultStore) MarkSweepSubmitted(_ context.Context, input outport.MarkEVMSweepSubmittedInput) error {
	f.submittedInputs = append(f.submittedInputs, input)
	return nil
}

func (f *fakeSweepVaultStore) MarkSweepSucceeded(_ context.Context, input outport.MarkEVMSweepResultInput) error {
	f.succeededInputs = append(f.succeededInputs, input)
	return nil
}

func (f *fakeSweepVaultStore) MarkSweepFailed(_ context.Context, input outport.MarkEVMSweepResultInput) error {
	f.failedInputs = append(f.failedInputs, input)
	return nil
}

type fakeSweepExecutor struct {
	executeOutput outport.ExecuteEVMSweepBatchOutput
	executeErr    error
	waitErr       error
	executeInputs []outport.ExecuteEVMSweepBatchInput
	waitInputs    []outport.WaitForEVMSweepTransactionInput
}

func (f *fakeSweepExecutor) ExecuteBatch(
	_ context.Context,
	input outport.ExecuteEVMSweepBatchInput,
) (outport.ExecuteEVMSweepBatchOutput, error) {
	f.executeInputs = append(f.executeInputs, input)
	return f.executeOutput, f.executeErr
}

func (f *fakeSweepExecutor) WaitForTransaction(
	_ context.Context,
	input outport.WaitForEVMSweepTransactionInput,
) error {
	f.waitInputs = append(f.waitInputs, input)
	return f.waitErr
}

func TestRunEVMSweepUseCaseDryRun(t *testing.T) {
	store := &fakeSweepVaultStore{
		candidates: []outport.EVMSweepCandidateRecord{
			{
				PaymentAddressID: 1,
				Network:          valueobjects.NetworkID("sepolia"),
				FactoryAddress:   "0xfactory",
				CollectorAddress: "0xcollector",
				AssetCode:        "usdt",
				AssetType:        "erc20",
				TokenAddress:     "0xtoken",
				SaltHex:          "0x01",
				IssuedAt:         time.Now().UTC(),
			},
			{
				PaymentAddressID: 2,
				Network:          valueobjects.NetworkID("sepolia"),
				FactoryAddress:   "0xfactory",
				CollectorAddress: "0xcollector",
				AssetCode:        "usdt",
				AssetType:        "erc20",
				TokenAddress:     "0xtoken",
				SaltHex:          "0x02",
				IssuedAt:         time.Now().UTC(),
			},
		},
	}

	useCase := NewRunEVMSweepUseCase(store, nil, nil)
	output, err := useCase.Execute(context.Background(), dto.RunEVMSweepInput{
		Network: valueobjects.NetworkID("sepolia"),
		DryRun:  true,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if output.CandidateCount != 2 {
		t.Fatalf("unexpected candidate count: got %d", output.CandidateCount)
	}
	if output.BatchCount != 1 {
		t.Fatalf("unexpected batch count: got %d", output.BatchCount)
	}
	if output.Batches[0].Status != "dry_run" {
		t.Fatalf("unexpected batch status: got %q", output.Batches[0].Status)
	}
}

func TestRunEVMSweepUseCaseExecuteSuccess(t *testing.T) {
	store := &fakeSweepVaultStore{
		candidates: []outport.EVMSweepCandidateRecord{
			{
				PaymentAddressID: 10,
				Network:          valueobjects.NetworkID("sepolia"),
				FactoryAddress:   "0xfactory",
				CollectorAddress: "0xcollector",
				AssetCode:        "eth",
				AssetType:        "native",
				SaltHex:          "0x1111111111111111111111111111111111111111111111111111111111111111",
				IssuedAt:         time.Now().UTC(),
			},
		},
	}
	executor := &fakeSweepExecutor{
		executeOutput: outport.ExecuteEVMSweepBatchOutput{TxHash: "0xtx"},
	}
	useCase := NewRunEVMSweepUseCase(store, executor, map[valueobjects.NetworkID]dto.EVMSweepNetworkRuntime{
		valueobjects.NetworkID("sepolia"): {
			Network:           valueobjects.NetworkID("sepolia"),
			RPCURL:            "https://sepolia.example",
			SweeperPrivateKey: "0xabc",
		},
	})

	output, err := useCase.Execute(context.Background(), dto.RunEVMSweepInput{
		Network: valueobjects.NetworkID("sepolia"),
		DryRun:  false,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(executor.executeInputs) != 1 {
		t.Fatalf("unexpected execute call count: got %d", len(executor.executeInputs))
	}
	if len(store.submittedInputs) != 1 {
		t.Fatalf("unexpected submitted count: got %d", len(store.submittedInputs))
	}
	if len(store.succeededInputs) != 1 {
		t.Fatalf("unexpected succeeded count: got %d", len(store.succeededInputs))
	}
	if output.Batches[0].TxHash != "0xtx" {
		t.Fatalf("unexpected tx hash: got %q", output.Batches[0].TxHash)
	}
	if output.Batches[0].Status != "succeeded" {
		t.Fatalf("unexpected batch status: got %q", output.Batches[0].Status)
	}
}

func TestRunEVMSweepUseCaseMarksFailedWhenReceiptFails(t *testing.T) {
	store := &fakeSweepVaultStore{
		candidates: []outport.EVMSweepCandidateRecord{
			{
				PaymentAddressID: 10,
				Network:          valueobjects.NetworkID("sepolia"),
				FactoryAddress:   "0xfactory",
				CollectorAddress: "0xcollector",
				AssetCode:        "eth",
				AssetType:        "native",
				SaltHex:          "0x1111111111111111111111111111111111111111111111111111111111111111",
				IssuedAt:         time.Now().UTC(),
			},
		},
	}
	executor := &fakeSweepExecutor{
		executeOutput: outport.ExecuteEVMSweepBatchOutput{TxHash: "0xtx"},
		waitErr:       errors.New("transaction reverted"),
	}
	useCase := NewRunEVMSweepUseCase(store, executor, map[valueobjects.NetworkID]dto.EVMSweepNetworkRuntime{
		valueobjects.NetworkID("sepolia"): {
			Network:           valueobjects.NetworkID("sepolia"),
			RPCURL:            "https://sepolia.example",
			SweeperPrivateKey: "0xabc",
		},
	})

	output, err := useCase.Execute(context.Background(), dto.RunEVMSweepInput{
		Network: valueobjects.NetworkID("sepolia"),
	})
	if err == nil {
		t.Fatal("expected execute error")
	}
	if len(store.failedInputs) != 1 {
		t.Fatalf("unexpected failed count: got %d", len(store.failedInputs))
	}
	if output.Batches[0].Status != "failed" {
		t.Fatalf("unexpected batch status: got %q", output.Batches[0].Status)
	}
}
