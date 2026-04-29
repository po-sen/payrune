//go:build !js || !wasm

package bitcoin

import (
	"context"
	"errors"
)

type unsupportedCloudflareEsploraBridge struct{}

func NewCloudflareEsploraBridge() cloudflareEsploraBridge {
	return &unsupportedCloudflareEsploraBridge{}
}

func (b *unsupportedCloudflareEsploraBridge) FetchLatestBlockHeight(
	context.Context,
	string,
	string,
) (int64, error) {
	return 0, errors.New("cloudflare bitcoin esplora bridge is only available in js/wasm")
}

func (b *unsupportedCloudflareEsploraBridge) FetchAddressChainTransactions(
	context.Context,
	string,
	string,
	string,
) ([]esploraTransaction, error) {
	return nil, errors.New("cloudflare bitcoin esplora bridge is only available in js/wasm")
}

func (b *unsupportedCloudflareEsploraBridge) FetchAddressMempoolTransactions(
	context.Context,
	string,
	string,
	string,
) ([]esploraTransaction, error) {
	return nil, errors.New("cloudflare bitcoin esplora bridge is only available in js/wasm")
}
