//go:build js && wasm

package main

import (
	"context"
	"fmt"
	"strings"
	"syscall/js"

	"payrune/internal/bootstrap"
)

const (
	operationAPI               = "api"
	operationPoller            = "poller"
	operationWebhookDispatcher = "webhook_dispatcher"
)

func main() {
	js.Global().Set("payruneWorkerHandle", js.FuncOf(func(this js.Value, args []js.Value) any {
		promiseCtor := js.Global().Get("Promise")
		executor := js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
			resolve := promiseArgs[0]
			reject := promiseArgs[1]

			operation := ""
			if len(args) > 0 {
				operation = args[0].String()
			}

			payload := ""
			if len(args) > 1 {
				payload = args[1].String()
			}

			go func() {
				responseJSON, err := dispatchOperation(context.Background(), operation, payload)
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

func dispatchOperation(ctx context.Context, operation string, payload string) (string, error) {
	switch strings.TrimSpace(operation) {
	case operationAPI:
		return bootstrap.HandleCloudflareAPIRequestJSON(ctx, payload)
	case operationPoller:
		return bootstrap.HandleCloudflarePollerRequestJSON(ctx, payload)
	case operationWebhookDispatcher:
		return bootstrap.HandleCloudflareReceiptWebhookDispatcherRequestJSON(ctx, payload)
	default:
		return "", fmt.Errorf("unsupported payrune worker operation: %s", operation)
	}
}
