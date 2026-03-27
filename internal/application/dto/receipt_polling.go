package dto

import (
	"time"

	"payrune/internal/domain/valueobjects"
)

type RunReceiptPollingCycleInput struct {
	BatchSize          int
	RescheduleInterval time.Duration
	ClaimTTL           time.Duration
	Chain              valueobjects.ChainID
	Network            valueobjects.NetworkID
}

type RunReceiptPollingCycleOutput struct {
	ClaimedCount         int
	UpdatedCount         int
	TerminalFailedCount  int
	ProcessingErrorCount int
}
