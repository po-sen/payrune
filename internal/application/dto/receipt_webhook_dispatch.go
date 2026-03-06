package dto

import "time"

type RunReceiptWebhookDispatchCycleInput struct {
	BatchSize   int
	DispatchTTL time.Duration
	RetryDelay  time.Duration
	MaxAttempts int32
}

type RunReceiptWebhookDispatchCycleOutput struct {
	ClaimedCount int
	SentCount    int
	RetriedCount int
	FailedCount  int
}
