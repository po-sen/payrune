package cloudflareworker

import (
	"context"
	"errors"
	"time"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
)

type WebhookDispatcherRequest struct {
	BatchSize   int
	DispatchTTL time.Duration
	RetryDelay  time.Duration
	MaxAttempts int32
}

type WebhookDispatcherResponse struct {
	ClaimedCount int `json:"claimedCount"`
	SentCount    int `json:"sentCount"`
	RetriedCount int `json:"retriedCount"`
	FailedCount  int `json:"failedCount"`
}

type WebhookDispatcherDependencies struct {
	RunReceiptWebhookDispatchCycleUseCase inport.RunReceiptWebhookDispatchCycleUseCase
}

type WebhookDispatcherHandler struct {
	useCase inport.RunReceiptWebhookDispatchCycleUseCase
}

func NewWebhookDispatcherHandler(deps WebhookDispatcherDependencies) *WebhookDispatcherHandler {
	return &WebhookDispatcherHandler{useCase: deps.RunReceiptWebhookDispatchCycleUseCase}
}

func (h *WebhookDispatcherHandler) Handle(
	ctx context.Context,
	request WebhookDispatcherRequest,
) (WebhookDispatcherResponse, error) {
	if h == nil || h.useCase == nil {
		return WebhookDispatcherResponse{}, errors.New("cloudflare worker webhook dispatcher use case is not configured")
	}

	output, err := h.useCase.Execute(ctx, dto.RunReceiptWebhookDispatchCycleInput{
		BatchSize:   request.BatchSize,
		DispatchTTL: request.DispatchTTL,
		RetryDelay:  request.RetryDelay,
		MaxAttempts: request.MaxAttempts,
	})
	if err != nil {
		return WebhookDispatcherResponse{}, err
	}

	return WebhookDispatcherResponse{
		ClaimedCount: output.ClaimedCount,
		SentCount:    output.SentCount,
		RetriedCount: output.RetriedCount,
		FailedCount:  output.FailedCount,
	}, nil
}
