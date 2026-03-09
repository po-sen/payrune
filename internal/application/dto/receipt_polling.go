package dto

import "time"

type RunReceiptPollingCycleInput struct {
	BatchSize           int
	ReceiptPollInterval time.Duration
	ClaimTTL            time.Duration
	Chain               string
	Network             string
}

type RunReceiptPollingCycleOutput struct {
	ClaimedCount         int
	UpdatedCount         int
	TerminalFailedCount  int
	ProcessingErrorCount int
}
