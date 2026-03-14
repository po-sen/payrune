package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"payrune/internal/application/dto"
)

type fakeRunReceiptWebhookDispatchCycleUseCase struct {
	input  dto.RunReceiptWebhookDispatchCycleInput
	output dto.RunReceiptWebhookDispatchCycleOutput
	err    error
}

func (f *fakeRunReceiptWebhookDispatchCycleUseCase) Execute(
	_ context.Context,
	input dto.RunReceiptWebhookDispatchCycleInput,
) (dto.RunReceiptWebhookDispatchCycleOutput, error) {
	f.input = input
	if f.err != nil {
		return dto.RunReceiptWebhookDispatchCycleOutput{}, f.err
	}
	return f.output, nil
}

func TestWebhookDispatcherHandlerHandle(t *testing.T) {
	useCase := &fakeRunReceiptWebhookDispatchCycleUseCase{
		output: dto.RunReceiptWebhookDispatchCycleOutput{
			ClaimedCount: 3,
			SentCount:    2,
			RetriedCount: 1,
			FailedCount:  4,
		},
	}

	handler := NewWebhookDispatcherHandler(WebhookDispatcherDependencies{
		RunReceiptWebhookDispatchCycleUseCase: useCase,
	})

	response, err := handler.Handle(context.Background(), WebhookDispatcherRequest{
		BatchSize:   5,
		DispatchTTL: 30 * time.Second,
		RetryDelay:  time.Minute,
		MaxAttempts: 10,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if useCase.input.BatchSize != 5 || useCase.input.MaxAttempts != 10 {
		t.Fatalf("unexpected use case input: %+v", useCase.input)
	}
	if response.ClaimedCount != 3 || response.FailedCount != 4 {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestWebhookDispatcherHandlerHandleReturnsUseCaseError(t *testing.T) {
	handler := NewWebhookDispatcherHandler(WebhookDispatcherDependencies{
		RunReceiptWebhookDispatchCycleUseCase: &fakeRunReceiptWebhookDispatchCycleUseCase{err: errors.New("boom")},
	})

	_, err := handler.Handle(context.Background(), WebhookDispatcherRequest{BatchSize: 1})
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}
