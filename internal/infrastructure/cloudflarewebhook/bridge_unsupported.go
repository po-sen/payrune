//go:build !js || !wasm

package cloudflarewebhook

import (
	"context"
	"errors"
)

type unsupportedBridge struct{}

func NewBridge() Bridge {
	return &unsupportedBridge{}
}

func (b *unsupportedBridge) PostJSON(
	context.Context,
	PostInput,
) error {
	return errors.New("cloudflare payment receipt webhook bridge is only available in js/wasm")
}
