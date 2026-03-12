//go:build js && wasm

package bitcoin

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"syscall/js"

	"payrune/internal/domain/value_objects"
)

const (
	jsFnBitcoinFetchLatestBlockHeight        = "__payruneBitcoinFetchLatestBlockHeight"
	jsFnBitcoinFetchAddressChainTransactions = "__payruneBitcoinFetchAddressChainTransactions"
	jsFnBitcoinFetchAddressMempoolTxs        = "__payruneBitcoinFetchAddressMempoolTransactions"
)

type jsCloudflareEsploraBridge struct{}

func NewCloudflareEsploraBridge() CloudflareEsploraBridge {
	return &jsCloudflareEsploraBridge{}
}

func (b *jsCloudflareEsploraBridge) FetchLatestBlockHeight(
	ctx context.Context,
	bridgeID string,
	network value_objects.NetworkID,
) (int64, error) {
	value, err := awaitBitcoinPromise(
		ctx,
		js.Global().Call(jsFnBitcoinFetchLatestBlockHeight, bridgeID, string(network)),
	)
	if err != nil {
		return 0, err
	}

	height := value.Float()
	if math.Trunc(height) != height {
		return 0, errors.New("latest block height must be an integer")
	}
	return int64(height), nil
}

func (b *jsCloudflareEsploraBridge) FetchAddressChainTransactions(
	ctx context.Context,
	bridgeID string,
	network value_objects.NetworkID,
	address string,
) ([]esploraTransaction, error) {
	return fetchBitcoinTransactions(
		ctx,
		js.Global().Call(jsFnBitcoinFetchAddressChainTransactions, bridgeID, string(network), address),
	)
}

func (b *jsCloudflareEsploraBridge) FetchAddressMempoolTransactions(
	ctx context.Context,
	bridgeID string,
	network value_objects.NetworkID,
	address string,
) ([]esploraTransaction, error) {
	return fetchBitcoinTransactions(
		ctx,
		js.Global().Call(jsFnBitcoinFetchAddressMempoolTxs, bridgeID, string(network), address),
	)
}

func fetchBitcoinTransactions(ctx context.Context, promise js.Value) ([]esploraTransaction, error) {
	value, err := awaitBitcoinPromise(ctx, promise)
	if err != nil {
		return nil, err
	}

	var transactions []esploraTransaction
	if err := json.Unmarshal([]byte(value.String()), &transactions); err != nil {
		return nil, err
	}
	return transactions, nil
}

func awaitBitcoinPromise(ctx context.Context, promise js.Value) (js.Value, error) {
	resultCh := make(chan js.Value, 1)
	errCh := make(chan error, 1)

	thenFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		resultCh <- args[0]
		return nil
	})
	catchFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			errCh <- errors.New("javascript promise rejected")
			return nil
		}
		errCh <- jsBitcoinError(args[0])
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

func jsBitcoinError(value js.Value) error {
	if value.IsNull() || value.IsUndefined() {
		return errors.New("bitcoin observer bridge error is missing")
	}

	message := value.Get("message")
	if message.Truthy() {
		return errors.New(message.String())
	}
	return errors.New(value.String())
}
