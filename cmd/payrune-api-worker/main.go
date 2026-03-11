//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"syscall/js"

	inboundadapter "payrune/internal/adapters/inbound/cloudflareworker"
)

type requestEnvelope struct {
	Request  inboundadapter.Request `json:"request"`
	Env      map[string]string      `json:"env"`
	BridgeID string                 `json:"bridgeId"`
}

type responseEnvelope struct {
	Response inboundadapter.Response `json:"response"`
}

type responsePayload = inboundadapter.Response

func main() {
	js.Global().Set("payruneHandle", js.FuncOf(func(this js.Value, args []js.Value) any {
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

	response, err := handleRequest(context.Background(), envelope)
	if err != nil {
		return "", err
	}

	encoded, err := json.Marshal(responseEnvelope{Response: response})
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func handleRequest(ctx context.Context, envelope requestEnvelope) (responsePayload, error) {
	handler, err := buildHTTPHandler(envelope.Env, envelope.BridgeID)
	if err != nil {
		return inboundadapter.Response{}, err
	}

	adapter := inboundadapter.NewAdapter(handler)
	return adapter.Handle(ctx, envelope.Request)
}
