//go:build !js || !wasm

package bitcoin

import (
	"context"
	"errors"

	"payrune/internal/domain/value_objects"
)

type unsupportedCloudflareEsploraBridge struct{}

func NewCloudflareEsploraBridge() CloudflareEsploraBridge {
	return &unsupportedCloudflareEsploraBridge{}
}

func (b *unsupportedCloudflareEsploraBridge) FetchLatestBlockHeight(
	context.Context,
	string,
	value_objects.NetworkID,
) (int64, error) {
	return 0, errors.New("cloudflare bitcoin esplora bridge is only available in js/wasm")
}

func (b *unsupportedCloudflareEsploraBridge) FetchAddressChainTransactions(
	context.Context,
	string,
	value_objects.NetworkID,
	string,
) ([]esploraTransaction, error) {
	return nil, errors.New("cloudflare bitcoin esplora bridge is only available in js/wasm")
}

func (b *unsupportedCloudflareEsploraBridge) FetchAddressMempoolTransactions(
	context.Context,
	string,
	value_objects.NetworkID,
	string,
) ([]esploraTransaction, error) {
	return nil, errors.New("cloudflare bitcoin esplora bridge is only available in js/wasm")
}
