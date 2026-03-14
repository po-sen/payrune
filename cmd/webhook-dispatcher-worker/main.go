//go:build js && wasm

package main

import (
	"context"
	"syscall/js"

	"payrune/internal/bootstrap"
)

func main() {
	js.Global().Set("payruneWebhookDispatcherHandle", js.FuncOf(func(this js.Value, args []js.Value) any {
		promiseCtor := js.Global().Get("Promise")
		executor := js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
			resolve := promiseArgs[0]
			reject := promiseArgs[1]

			payload := ""
			if len(args) > 0 {
				payload = args[0].String()
			}

			go func() {
				responseJSON, err := bootstrap.HandleCloudflareReceiptWebhookDispatcherRequestJSON(context.Background(), payload)
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
