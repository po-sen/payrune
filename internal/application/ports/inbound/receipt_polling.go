package inbound

import (
	"context"
	"time"
)

type RunReceiptPollingCycleInput struct {
	BatchSize          int
	RescheduleInterval time.Duration
	ClaimTTL           time.Duration
	Chain              string
	Network            string
}

type RunReceiptPollingCycleOutput struct {
	ClaimedCount         int
	UpdatedCount         int
	TerminalFailedCount  int
	ProcessingErrorCount int
}

type RunReceiptPollingCycleUseCase interface {
	Execute(ctx context.Context, input RunReceiptPollingCycleInput) (RunReceiptPollingCycleOutput, error)
}
