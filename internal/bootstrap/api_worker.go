package bootstrap

import (
	"context"
	"encoding/json"

	httpcloudflare "payrune/internal/adapters/inbound/http/cloudflare"
	"payrune/internal/infrastructure/di"
)

type apiWorkerRequestEnvelope struct {
	Request  httpcloudflare.Request `json:"request"`
	Env      map[string]string      `json:"env"`
	BridgeID string                 `json:"bridgeId"`
}

type apiWorkerResponseEnvelope struct {
	Response httpcloudflare.Response `json:"response"`
}

func HandleCloudflareAPIRequestJSON(ctx context.Context, payload string) (string, error) {
	var envelope apiWorkerRequestEnvelope
	if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
		return "", err
	}

	response, err := handleCloudflareAPIRequest(ctx, envelope)
	if err != nil {
		return "", err
	}

	encoded, err := json.Marshal(apiWorkerResponseEnvelope{Response: response})
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func handleCloudflareAPIRequest(
	ctx context.Context,
	envelope apiWorkerRequestEnvelope,
) (httpcloudflare.Response, error) {
	handler, err := di.BuildCloudflareAPIHTTPHandler(envelope.Env, envelope.BridgeID)
	if err != nil {
		return httpcloudflare.Response{}, err
	}

	return httpcloudflare.HandleRequest(ctx, handler, envelope.Request)
}
