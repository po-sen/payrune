package scheduler

import (
	"context"
	"errors"
	"time"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
)

type PollerRequest struct {
	BatchSize          int
	RescheduleInterval time.Duration
	ClaimTTL           time.Duration
	Chain              string
	Network            string
}

type PollerResponse struct {
	ClaimedCount         int `json:"claimedCount"`
	UpdatedCount         int `json:"updatedCount"`
	TerminalFailedCount  int `json:"terminalFailedCount"`
	ProcessingErrorCount int `json:"processingErrorCount"`
}

type PollerDependencies struct {
	RunReceiptPollingCycleUseCase inport.RunReceiptPollingCycleUseCase
}

type PollerHandler struct {
	useCase inport.RunReceiptPollingCycleUseCase
}

func NewPollerHandler(deps PollerDependencies) *PollerHandler {
	return &PollerHandler{useCase: deps.RunReceiptPollingCycleUseCase}
}

func (h *PollerHandler) Handle(ctx context.Context, request PollerRequest) (PollerResponse, error) {
	if h == nil || h.useCase == nil {
		return PollerResponse{}, errors.New("cloudflare worker poller use case is not configured")
	}

	output, err := h.useCase.Execute(ctx, dto.RunReceiptPollingCycleInput{
		BatchSize:          request.BatchSize,
		RescheduleInterval: request.RescheduleInterval,
		ClaimTTL:           request.ClaimTTL,
		Chain:              request.Chain,
		Network:            request.Network,
	})
	if err != nil {
		return PollerResponse{}, err
	}

	return PollerResponse{
		ClaimedCount:         output.ClaimedCount,
		UpdatedCount:         output.UpdatedCount,
		TerminalFailedCount:  output.TerminalFailedCount,
		ProcessingErrorCount: output.ProcessingErrorCount,
	}, nil
}
