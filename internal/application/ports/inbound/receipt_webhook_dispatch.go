package inbound

import (
	"context"
	"time"
)

type RunReceiptWebhookDispatchCycleInput struct {
	BatchSize   int
	MaxAttempts int32
	RetryDelay  time.Duration
	DispatchTTL time.Duration
}

type RunReceiptWebhookDispatchCycleOutput struct {
	ClaimedCount int
	SentCount    int
	RetriedCount int
	FailedCount  int
}

type RunReceiptWebhookDispatchCycleUseCase interface {
	Execute(ctx context.Context, input RunReceiptWebhookDispatchCycleInput) (RunReceiptWebhookDispatchCycleOutput, error)
}
