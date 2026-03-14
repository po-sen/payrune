package bootstrap

import (
	"context"
	"encoding/json"

	scheduleradapter "payrune/internal/adapters/inbound/scheduler"
	"payrune/internal/infrastructure/di"
)

type receiptWebhookDispatcherWorkerRequestEnvelope struct {
	Env              map[string]string `json:"env"`
	PostgresBridgeID string            `json:"postgresBridgeId"`
	ScheduledTime    string            `json:"scheduledTime"`
	Cron             string            `json:"cron"`
}

type receiptWebhookDispatcherWorkerResponseEnvelope struct {
	Output scheduleradapter.WebhookDispatcherResponse `json:"output"`
}

func HandleCloudflareReceiptWebhookDispatcherRequestJSON(ctx context.Context, payload string) (string, error) {
	var envelope receiptWebhookDispatcherWorkerRequestEnvelope
	if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
		return "", err
	}

	output, err := handleCloudflareReceiptWebhookDispatcherRequest(ctx, envelope)
	if err != nil {
		return "", err
	}

	encoded, err := json.Marshal(receiptWebhookDispatcherWorkerResponseEnvelope{Output: output})
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func handleCloudflareReceiptWebhookDispatcherRequest(
	ctx context.Context,
	envelope receiptWebhookDispatcherWorkerRequestEnvelope,
) (scheduleradapter.WebhookDispatcherResponse, error) {
	handler, request, err := di.BuildCloudflareWebhookDispatcherRuntime(
		envelope.Env,
		envelope.PostgresBridgeID,
	)
	if err != nil {
		return scheduleradapter.WebhookDispatcherResponse{}, err
	}

	return handler.Handle(ctx, request)
}
