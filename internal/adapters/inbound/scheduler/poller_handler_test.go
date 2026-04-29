package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	inport "payrune/internal/application/ports/inbound"
)

type fakeRunReceiptPollingCycleUseCase struct {
	input  inport.RunReceiptPollingCycleInput
	output inport.RunReceiptPollingCycleOutput
	err    error
}

func (f *fakeRunReceiptPollingCycleUseCase) Execute(
	_ context.Context,
	input inport.RunReceiptPollingCycleInput,
) (inport.RunReceiptPollingCycleOutput, error) {
	f.input = input
	if f.err != nil {
		return inport.RunReceiptPollingCycleOutput{}, f.err
	}
	return f.output, nil
}

func TestPollerHandlerHandle(t *testing.T) {
	useCase := &fakeRunReceiptPollingCycleUseCase{
		output: inport.RunReceiptPollingCycleOutput{
			ClaimedCount:         3,
			UpdatedCount:         2,
			TerminalFailedCount:  1,
			ProcessingErrorCount: 4,
		},
	}

	handler := NewPollerHandler(PollerDependencies{
		RunReceiptPollingCycleUseCase: useCase,
	})
	response, err := handler.Handle(context.Background(), PollerRequest{
		BatchSize:          5,
		RescheduleInterval: 10 * time.Minute,
		ClaimTTL:           30 * time.Second,
		Chain:              "bitcoin",
		Network:            "mainnet",
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if useCase.input.BatchSize != 5 || useCase.input.Network != "mainnet" {
		t.Fatalf("unexpected use case input: %+v", useCase.input)
	}
	if response.ClaimedCount != 3 || response.TerminalFailedCount != 1 {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestPollerHandlerHandleReturnsUseCaseError(t *testing.T) {
	handler := NewPollerHandler(PollerDependencies{
		RunReceiptPollingCycleUseCase: &fakeRunReceiptPollingCycleUseCase{err: inport.ErrDependencyFailure},
	})

	_, err := handler.Handle(context.Background(), PollerRequest{BatchSize: 1})
	if !errors.Is(err, inport.ErrDependencyFailure) {
		t.Fatalf("expected ErrDependencyFailure, got %v", err)
	}
}
