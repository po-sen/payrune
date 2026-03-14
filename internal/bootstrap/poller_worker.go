package bootstrap

import (
	"context"
	"encoding/json"

	scheduleradapter "payrune/internal/adapters/inbound/scheduler"
	"payrune/internal/infrastructure/di"
)

type pollerWorkerRequestEnvelope struct {
	Env              map[string]string `json:"env"`
	PostgresBridgeID string            `json:"postgresBridgeId"`
	BitcoinBridgeID  string            `json:"bitcoinBridgeId"`
	ScheduledTime    string            `json:"scheduledTime"`
	Cron             string            `json:"cron"`
}

type pollerWorkerResponseEnvelope struct {
	Output scheduleradapter.PollerResponse `json:"output"`
}

func HandleCloudflarePollerRequestJSON(ctx context.Context, payload string) (string, error) {
	var envelope pollerWorkerRequestEnvelope
	if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
		return "", err
	}

	output, err := handleCloudflarePollerRequest(ctx, envelope)
	if err != nil {
		return "", err
	}

	encoded, err := json.Marshal(pollerWorkerResponseEnvelope{Output: output})
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func handleCloudflarePollerRequest(
	ctx context.Context,
	envelope pollerWorkerRequestEnvelope,
) (scheduleradapter.PollerResponse, error) {
	handler, request, err := di.BuildCloudflarePollerRuntime(
		envelope.Env,
		envelope.PostgresBridgeID,
		envelope.BitcoinBridgeID,
	)
	if err != nil {
		return scheduleradapter.PollerResponse{}, err
	}

	return handler.Handle(ctx, request)
}
