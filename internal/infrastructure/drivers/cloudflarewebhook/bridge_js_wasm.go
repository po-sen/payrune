//go:build js && wasm

package cloudflarewebhookdriver

import (
	"context"
	"errors"
	"syscall/js"
	"time"

	webhookadapter "payrune/internal/adapters/outbound/webhook"
)

const (
	jsFnWebhookPost       = "__payruneWebhookPost"
	defaultWebhookTimeout = 10 * time.Second
)

type jsBridge struct{}

func NewBridge() webhookadapter.CloudflarePaymentReceiptStatusWebhookBridge {
	return &jsBridge{}
}

func (b *jsBridge) PostJSON(
	ctx context.Context,
	input webhookadapter.CloudflarePaymentReceiptStatusWebhookPostInput,
) error {
	headers := js.Global().Get("Object").New()
	for key, value := range input.Headers {
		headers.Set(key, value)
	}

	timeoutMs := input.Timeout.Milliseconds()
	if input.Timeout <= 0 {
		timeoutMs = defaultWebhookTimeout.Milliseconds()
	}

	_, err := awaitPromise(
		ctx,
		js.Global().Call(jsFnWebhookPost, input.Binding, input.Path, timeoutMs, headers, string(input.Body)),
	)
	return err
}

func awaitPromise(ctx context.Context, promise js.Value) (js.Value, error) {
	resultCh := make(chan js.Value, 1)
	errCh := make(chan error, 1)

	thenFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			resultCh <- args[0]
		} else {
			resultCh <- js.Null()
		}
		return nil
	})
	catchFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			errCh <- errors.New("javascript webhook promise rejected")
			return nil
		}
		errCh <- jsError(args[0])
		return nil
	})

	promise.Call("then", thenFunc).Call("catch", catchFunc)

	select {
	case result := <-resultCh:
		thenFunc.Release()
		catchFunc.Release()
		return result, nil
	case err := <-errCh:
		thenFunc.Release()
		catchFunc.Release()
		return js.Value{}, err
	case <-ctx.Done():
		thenFunc.Release()
		catchFunc.Release()
		return js.Value{}, ctx.Err()
	}
}

func jsError(value js.Value) error {
	if value.IsNull() || value.IsUndefined() {
		return errors.New("webhook bridge error is missing")
	}

	message := value.Get("message")
	if message.Truthy() {
		return errors.New(message.String())
	}
	return errors.New(value.String())
}
