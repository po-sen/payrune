//go:build !js || !wasm

package cloudflarewebhookdriver

import (
	"context"
	"errors"

	webhookadapter "payrune/internal/adapters/outbound/webhook"
)

type unsupportedBridge struct{}

func NewBridge() webhookadapter.CloudflarePaymentReceiptStatusWebhookBridge {
	return &unsupportedBridge{}
}

func (b *unsupportedBridge) PostJSON(
	context.Context,
	webhookadapter.CloudflarePaymentReceiptStatusWebhookPostInput,
) error {
	return errors.New("cloudflare payment receipt webhook bridge is only available in js/wasm")
}
