//go:build js && wasm

package cloudflarepostgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"syscall/js"
	"time"
)

const (
	jsFnBeginTx  = "__payrunePgBeginTx"
	jsFnCommitTx = "__payrunePgCommitTx"
	jsFnRollback = "__payrunePgRollbackTx"
	jsFnExec     = "__payrunePgExec"
	jsFnQuery    = "__payrunePgQuery"
	jsFnQueryRow = "__payrunePgQueryRow"
)

type JSBridge struct{}

func NewJSBridge() Bridge {
	return &JSBridge{}
}

func (b *JSBridge) BeginTx(ctx context.Context, bridgeID string) (string, error) {
	value, err := awaitPromise(ctx, js.Global().Call(jsFnBeginTx, bridgeID))
	if err != nil {
		return "", err
	}
	return value.String(), nil
}

func (b *JSBridge) CommitTx(ctx context.Context, bridgeID string, txID string) error {
	_, err := awaitPromise(ctx, js.Global().Call(jsFnCommitTx, bridgeID, txID))
	return err
}

func (b *JSBridge) RollbackTx(ctx context.Context, bridgeID string, txID string) error {
	_, err := awaitPromise(ctx, js.Global().Call(jsFnRollback, bridgeID, txID))
	return err
}

func (b *JSBridge) Exec(ctx context.Context, bridgeID string, txID string, query string, args []any) (int64, error) {
	value, err := awaitPromise(ctx, js.Global().Call(jsFnExec, bridgeID, txID, query, jsArgs(args)))
	if err != nil {
		return 0, err
	}
	return int64(value.Get("rowCount").Int()), nil
}

func (b *JSBridge) Query(ctx context.Context, bridgeID string, txID string, query string, args []any) ([][]any, error) {
	value, err := awaitPromise(ctx, js.Global().Call(jsFnQuery, bridgeID, txID, query, jsArgs(args)))
	if err != nil {
		return nil, err
	}
	return jsRowsToGo(value.Get("rows"))
}

func (b *JSBridge) QueryRow(
	ctx context.Context,
	bridgeID string,
	txID string,
	query string,
	args []any,
) ([]any, bool, error) {
	value, err := awaitPromise(ctx, js.Global().Call(jsFnQueryRow, bridgeID, txID, query, jsArgs(args)))
	if err != nil {
		return nil, false, err
	}
	if !value.Get("found").Bool() {
		return nil, false, nil
	}
	row, err := jsRowToGo(value.Get("row"))
	if err != nil {
		return nil, false, err
	}
	return row, true, nil
}

func jsArgs(args []any) js.Value {
	values := make([]any, 0, len(args))
	for _, arg := range args {
		values = append(values, normalizeArg(arg))
	}
	return js.ValueOf(values)
}

func normalizeArg(value any) any {
	switch typed := value.(type) {
	case nil:
		return nil
	case string:
		return typed
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case uint32:
		return int(typed)
	case time.Time:
		return typed.UTC().Format(time.RFC3339Nano)
	case *time.Time:
		if typed == nil {
			return nil
		}
		return typed.UTC().Format(time.RFC3339Nano)
	default:
		return fmt.Sprint(value)
	}
}

func jsRowsToGo(rowsValue js.Value) ([][]any, error) {
	length := rowsValue.Length()
	rows := make([][]any, 0, length)
	for i := 0; i < length; i++ {
		row, err := jsRowToGo(rowsValue.Index(i))
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func jsRowToGo(rowValue js.Value) ([]any, error) {
	if rowValue.IsNull() || rowValue.IsUndefined() {
		return nil, nil
	}
	length := rowValue.Length()
	row := make([]any, 0, length)
	for i := 0; i < length; i++ {
		row = append(row, jsValueToGo(rowValue.Index(i)))
	}
	return row, nil
}

func jsValueToGo(value js.Value) any {
	if value.IsNull() || value.IsUndefined() {
		return nil
	}

	switch value.Type() {
	case js.TypeBoolean:
		return value.Bool()
	case js.TypeNumber:
		return value.Float()
	case js.TypeString:
		return value.String()
	default:
		if value.InstanceOf(js.Global().Get("Date")) {
			return value.Call("toISOString").String()
		}
		return value.String()
	}
}

func awaitPromise(ctx context.Context, promise js.Value) (js.Value, error) {
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
		return errors.New("javascript error is missing")
	}
	message := value.Get("message")
	code := value.Get("code")
	constraint := value.Get("constraint")
	if message.Truthy() || code.Truthy() || constraint.Truthy() {
		return &QueryError{
			Message:    message.String(),
			Code:       code.String(),
			Constraint: constraint.String(),
		}
	}
	return errors.New(value.String())
}
