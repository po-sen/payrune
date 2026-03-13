//go:build !js || !wasm

package webhook

import (
	"context"
	"errors"
)

type unsupportedCloudflarePaymentReceiptStatusWebhookBridge struct{}

func NewCloudflarePaymentReceiptStatusWebhookBridge() CloudflarePaymentReceiptStatusWebhookBridge {
	return &unsupportedCloudflarePaymentReceiptStatusWebhookBridge{}
}

func (b *unsupportedCloudflarePaymentReceiptStatusWebhookBridge) PostJSON(
	context.Context,
	CloudflarePaymentReceiptStatusWebhookPostInput,
) error {
	return errors.New("cloudflare payment receipt webhook bridge is only available in js/wasm")
}
