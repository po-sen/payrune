package dto

import "time"

type RunReceiptPollingCycleInput struct {
	BatchSize                    int
	PollInterval                 time.Duration
	ClaimTTL                     time.Duration
	DefaultRequiredConfirmations int32
	Chain                        string
	Network                      string
}

type RunReceiptPollingCycleOutput struct {
	RegisteredCount int
	ClaimedCount    int
	UpdatedCount    int
	FailedCount     int
}
