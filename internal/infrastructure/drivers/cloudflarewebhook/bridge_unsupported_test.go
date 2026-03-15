//go:build !js || !wasm

package cloudflarewebhookdriver

import (
	"context"
	"testing"

	webhookadapter "payrune/internal/adapters/outbound/webhook"
)

func TestUnsupportedBridgeReturnsError(t *testing.T) {
	bridge := NewBridge()

	err := bridge.PostJSON(context.Background(), webhookadapter.CloudflarePaymentReceiptStatusWebhookPostInput{})
	if err == nil {
		t.Fatal("expected unsupported error")
	}
	if got := err.Error(); got != "cloudflare payment receipt webhook bridge is only available in js/wasm" {
		t.Fatalf("unexpected error: %q", got)
	}
}
