//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"syscall/js"

	inboundadapter "payrune/internal/adapters/inbound/cloudflareworker"
	"payrune/internal/infrastructure/di"
)

type requestEnvelope struct {
	Env              map[string]string `json:"env"`
	PostgresBridgeID string            `json:"postgresBridgeId"`
	BitcoinBridgeID  string            `json:"bitcoinBridgeId"`
	ScheduledTime    string            `json:"scheduledTime"`
	Cron             string            `json:"cron"`
}

type responseEnvelope struct {
	Output inboundadapter.PollerResponse `json:"output"`
}

type responsePayload = inboundadapter.PollerResponse

func main() {
	js.Global().Set("payrunePollerHandle", js.FuncOf(func(this js.Value, args []js.Value) any {
		promiseCtor := js.Global().Get("Promise")
		executor := js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
			resolve := promiseArgs[0]
			reject := promiseArgs[1]

			payload := ""
			if len(args) > 0 {
				payload = args[0].String()
			}

			go func() {
				responseJSON, err := handleRequestJSON(payload)
				if err != nil {
					reject.Invoke(err.Error())
					return
				}
				resolve.Invoke(responseJSON)
			}()

			return nil
		})
		promise := promiseCtor.New(executor)
		executor.Release()
		return promise
	}))

	select {}
}

func handleRequestJSON(payload string) (string, error) {
	var envelope requestEnvelope
	if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
		return "", err
	}

	output, err := handleRequest(context.Background(), envelope)
	if err != nil {
		return "", err
	}

	encoded, err := json.Marshal(responseEnvelope{Output: output})
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func handleRequest(ctx context.Context, envelope requestEnvelope) (responsePayload, error) {
	handler, request, err := di.BuildCloudflarePollerRuntime(envelope.Env, envelope.PostgresBridgeID, envelope.BitcoinBridgeID)
	if err != nil {
		return inboundadapter.PollerResponse{}, err
	}

	return handler.Handle(ctx, request)
}
