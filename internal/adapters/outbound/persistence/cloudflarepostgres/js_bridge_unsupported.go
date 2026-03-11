//go:build !js || !wasm

package cloudflarepostgres

import (
	"context"
	"errors"
)

type unsupportedBridge struct{}

func NewJSBridge() Bridge {
	return &unsupportedBridge{}
}

func (b *unsupportedBridge) BeginTx(context.Context, string) (string, error) {
	return "", errors.New("cloudflare postgres js bridge is only available in js/wasm")
}

func (b *unsupportedBridge) CommitTx(context.Context, string, string) error {
	return errors.New("cloudflare postgres js bridge is only available in js/wasm")
}

func (b *unsupportedBridge) RollbackTx(context.Context, string, string) error {
	return errors.New("cloudflare postgres js bridge is only available in js/wasm")
}

func (b *unsupportedBridge) Exec(context.Context, string, string, string, []any) (int64, error) {
	return 0, errors.New("cloudflare postgres js bridge is only available in js/wasm")
}

func (b *unsupportedBridge) Query(context.Context, string, string, string, []any) ([][]any, error) {
	return nil, errors.New("cloudflare postgres js bridge is only available in js/wasm")
}

func (b *unsupportedBridge) QueryRow(context.Context, string, string, string, []any) ([]any, bool, error) {
	return nil, false, errors.New("cloudflare postgres js bridge is only available in js/wasm")
}
