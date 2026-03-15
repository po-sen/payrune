//go:build !js || !wasm

package cloudflarepostgresdriver

import (
	"context"
	"testing"
)

func TestUnsupportedBridgeReturnsError(t *testing.T) {
	bridge := NewJSBridge()

	_, err := bridge.BeginTx(context.Background(), "bridge-1")
	if err == nil {
		t.Fatal("expected unsupported error")
	}
	if got := err.Error(); got != "cloudflare postgres js bridge is only available in js/wasm" {
		t.Fatalf("unexpected error: %q", got)
	}
}
