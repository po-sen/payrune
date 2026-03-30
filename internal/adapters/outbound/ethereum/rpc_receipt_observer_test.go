package ethereum

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

type testEthereumRPCServer struct {
	t                 *testing.T
	latestBlockHeight string
	balancesByKey     map[string]string
	statusCode        int
	rpcError          map[string]any
	lastAuthHeader    string
	requestedBalances []string
	requestedBlocks   []string
}

func newTestEthereumRPCServer(t *testing.T) (*testEthereumRPCServer, *httptest.Server) {
	t.Helper()

	handlerState := &testEthereumRPCServer{
		t:                 t,
		latestBlockHeight: "0x3",
		balancesByKey:     make(map[string]string),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerState.lastAuthHeader = r.Header.Get("Authorization")
		if handlerState.statusCode != 0 {
			w.WriteHeader(handlerState.statusCode)
			_, _ = w.Write([]byte("upstream error"))
			return
		}

		var request struct {
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode rpc request: %v", err)
		}

		if handlerState.rpcError != nil {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      1,
				"error":   handlerState.rpcError,
			})
			return
		}

		switch request.Method {
		case "eth_blockNumber":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      1,
				"result":  handlerState.latestBlockHeight,
			})
		case "eth_getBalance":
			if len(request.Params) != 2 {
				t.Fatalf("unexpected params for eth_getBalance: %d", len(request.Params))
			}

			var address string
			if err := json.Unmarshal(request.Params[0], &address); err != nil {
				t.Fatalf("decode balance address: %v", err)
			}
			var blockNumber string
			if err := json.Unmarshal(request.Params[1], &blockNumber); err != nil {
				t.Fatalf("decode balance block number: %v", err)
			}

			key := ethereumBalanceKey(address, blockNumber)
			handlerState.requestedBalances = append(handlerState.requestedBalances, key)

			result, ok := handlerState.balancesByKey[key]
			if !ok {
				result = "0x0"
			}

			_ = json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      1,
				"result":  result,
			})
		case "eth_getBlockByNumber":
			if len(request.Params) > 0 {
				var blockNumber string
				if err := json.Unmarshal(request.Params[0], &blockNumber); err == nil {
					handlerState.requestedBlocks = append(handlerState.requestedBlocks, blockNumber)
				}
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      1,
				"result":  nil,
			})
		default:
			t.Fatalf("unexpected rpc method: %s", request.Method)
		}
	}))

	return handlerState, server
}

func ethereumBalanceKey(address string, blockNumber string) string {
	return strings.ToLower(strings.TrimSpace(address)) + "@" + strings.ToLower(strings.TrimSpace(blockNumber))
}

func TestEthereumRPCReceiptObserverObserveAddress(t *testing.T) {
	state, server := newTestEthereumRPCServer(t)
	defer server.Close()

	state.balancesByKey[ethereumBalanceKey("0x1111111111111111111111111111111111111111", "0x3")] = "0xa"
	state.balancesByKey[ethereumBalanceKey("0x1111111111111111111111111111111111111111", "0x2")] = "0x7"

	observer, err := NewEthereumRPCReceiptObserver(map[valueobjects.NetworkID]*EthereumRPCObserverConfig{
		valueobjects.NetworkID("mainnet"): {
			Endpoint: server.URL,
			Username: "user",
			Password: "pass",
			Timeout:  5 * time.Second,
		},
	})
	if err != nil {
		t.Fatalf("NewEthereumRPCReceiptObserver returned error: %v", err)
	}

	output, err := observer.ObserveAddress(context.Background(), outport.ObservePaymentAddressInput{
		Network:               valueobjects.NetworkID("mainnet"),
		Address:               "0x1111111111111111111111111111111111111111",
		IssuedAt:              time.Unix(2500, 0).UTC(),
		ObservedTotalMinor:    0,
		ConfirmedTotalMinor:   0,
		UnconfirmedTotalMinor: 0,
		RequiredConfirmations: 2,
		LatestBlockHeight:     3,
		SinceBlockHeight:      0,
	})
	if err != nil {
		t.Fatalf("ObserveAddress returned error: %v", err)
	}

	if output.ObservedTotalMinor != 10 {
		t.Fatalf("unexpected observed total: got %d", output.ObservedTotalMinor)
	}
	if output.ConfirmedTotalMinor != 7 {
		t.Fatalf("unexpected confirmed total: got %d", output.ConfirmedTotalMinor)
	}
	if output.UnconfirmedTotalMinor != 3 {
		t.Fatalf("unexpected unconfirmed total: got %d", output.UnconfirmedTotalMinor)
	}
	if output.LatestBlockHeight != 3 {
		t.Fatalf("unexpected latest block height: got %d", output.LatestBlockHeight)
	}
	if !strings.HasPrefix(state.lastAuthHeader, "Basic ") {
		t.Fatalf("expected basic auth header, got %q", state.lastAuthHeader)
	}
	if len(state.requestedBlocks) != 0 {
		t.Fatalf("expected no block scan requests, got %v", state.requestedBlocks)
	}
	if got := strings.Join(state.requestedBalances, ","); got != ethereumBalanceKey("0x1111111111111111111111111111111111111111", "0x3")+","+ethereumBalanceKey("0x1111111111111111111111111111111111111111", "0x2") {
		t.Fatalf("unexpected balance requests: got %q", got)
	}
}

func TestEthereumRPCReceiptObserverObserveAddressWithInsufficientConfirmations(t *testing.T) {
	state, server := newTestEthereumRPCServer(t)
	defer server.Close()

	state.balancesByKey[ethereumBalanceKey("0x1111111111111111111111111111111111111111", "0x3")] = "0xa"

	observer, err := NewEthereumRPCReceiptObserver(map[valueobjects.NetworkID]*EthereumRPCObserverConfig{
		valueobjects.NetworkID("mainnet"): {
			Endpoint: server.URL,
		},
	})
	if err != nil {
		t.Fatalf("NewEthereumRPCReceiptObserver returned error: %v", err)
	}

	output, err := observer.ObserveAddress(context.Background(), outport.ObservePaymentAddressInput{
		Network:               valueobjects.NetworkID("mainnet"),
		Address:               "0x1111111111111111111111111111111111111111",
		IssuedAt:              time.Unix(2500, 0).UTC(),
		ObservedTotalMinor:    0,
		ConfirmedTotalMinor:   0,
		UnconfirmedTotalMinor: 0,
		RequiredConfirmations: 5,
		LatestBlockHeight:     3,
	})
	if err != nil {
		t.Fatalf("ObserveAddress returned error: %v", err)
	}

	if output.ObservedTotalMinor != 10 || output.ConfirmedTotalMinor != 0 || output.UnconfirmedTotalMinor != 10 {
		t.Fatalf("unexpected totals: got %+v", output)
	}
	if len(state.requestedBlocks) != 0 {
		t.Fatalf("expected no block scan requests, got %v", state.requestedBlocks)
	}
	if got := strings.Join(state.requestedBalances, ","); got != ethereumBalanceKey("0x1111111111111111111111111111111111111111", "0x3") {
		t.Fatalf("unexpected balance requests: got %q", got)
	}
}

func TestEthereumRPCReceiptObserverObserveAddressRejectsInconsistentBalances(t *testing.T) {
	state, server := newTestEthereumRPCServer(t)
	defer server.Close()

	state.balancesByKey[ethereumBalanceKey("0x1111111111111111111111111111111111111111", "0x3")] = "0x5"
	state.balancesByKey[ethereumBalanceKey("0x1111111111111111111111111111111111111111", "0x2")] = "0x7"

	observer, err := NewEthereumRPCReceiptObserver(map[valueobjects.NetworkID]*EthereumRPCObserverConfig{
		valueobjects.NetworkID("mainnet"): {
			Endpoint: server.URL,
		},
	})
	if err != nil {
		t.Fatalf("NewEthereumRPCReceiptObserver returned error: %v", err)
	}

	_, err = observer.ObserveAddress(context.Background(), outport.ObservePaymentAddressInput{
		Network:               valueobjects.NetworkID("mainnet"),
		Address:               "0x1111111111111111111111111111111111111111",
		IssuedAt:              time.Unix(2500, 0).UTC(),
		ObservedTotalMinor:    0,
		ConfirmedTotalMinor:   0,
		UnconfirmedTotalMinor: 0,
		RequiredConfirmations: 2,
		LatestBlockHeight:     3,
	})
	if err == nil {
		t.Fatal("expected inconsistent balance error")
	}
}

func TestEthereumRPCReceiptObserverFetchLatestBlockHeight(t *testing.T) {
	_, server := newTestEthereumRPCServer(t)
	defer server.Close()

	observer, err := NewEthereumRPCReceiptObserver(map[valueobjects.NetworkID]*EthereumRPCObserverConfig{
		valueobjects.NetworkID("sepolia"): {
			Endpoint: server.URL,
		},
	})
	if err != nil {
		t.Fatalf("NewEthereumRPCReceiptObserver returned error: %v", err)
	}

	height, err := observer.FetchLatestBlockHeight(context.Background(), valueobjects.NetworkID("sepolia"))
	if err != nil {
		t.Fatalf("FetchLatestBlockHeight returned error: %v", err)
	}
	if height != 3 {
		t.Fatalf("unexpected latest block height: got %d", height)
	}
}

func TestEthereumRPCReceiptObserverValidation(t *testing.T) {
	_, server := newTestEthereumRPCServer(t)
	defer server.Close()

	observer, err := NewEthereumRPCReceiptObserver(map[valueobjects.NetworkID]*EthereumRPCObserverConfig{
		valueobjects.NetworkID("mainnet"): {Endpoint: server.URL},
	})
	if err != nil {
		t.Fatalf("NewEthereumRPCReceiptObserver returned error: %v", err)
	}

	_, err = observer.ObserveAddress(context.Background(), outport.ObservePaymentAddressInput{
		Network:               valueobjects.NetworkID("mainnet"),
		Address:               "",
		IssuedAt:              time.Unix(2500, 0).UTC(),
		RequiredConfirmations: 1,
		LatestBlockHeight:     3,
	})
	if err == nil {
		t.Fatal("expected validation error for missing address")
	}

	_, err = observer.ObserveAddress(context.Background(), outport.ObservePaymentAddressInput{
		Network:               valueobjects.NetworkID("unknown"),
		Address:               "0x1111111111111111111111111111111111111111",
		IssuedAt:              time.Unix(2500, 0).UTC(),
		ObservedTotalMinor:    0,
		ConfirmedTotalMinor:   0,
		UnconfirmedTotalMinor: 0,
		RequiredConfirmations: 1,
		LatestBlockHeight:     3,
	})
	if err == nil {
		t.Fatal("expected missing network error")
	}
}

func TestEthereumRPCReceiptObserverEndpointError(t *testing.T) {
	state, server := newTestEthereumRPCServer(t)
	state.statusCode = http.StatusBadGateway
	defer server.Close()

	observer, err := NewEthereumRPCReceiptObserver(map[valueobjects.NetworkID]*EthereumRPCObserverConfig{
		valueobjects.NetworkID("mainnet"): {Endpoint: server.URL},
	})
	if err != nil {
		t.Fatalf("NewEthereumRPCReceiptObserver returned error: %v", err)
	}

	_, err = observer.FetchLatestBlockHeight(context.Background(), valueobjects.NetworkID("mainnet"))
	if err == nil {
		t.Fatal("expected endpoint error")
	}
}

func TestNewEthereumRPCReceiptObserverValidation(t *testing.T) {
	_, err := NewEthereumRPCReceiptObserver(nil)
	if err == nil {
		t.Fatal("expected error when configs missing")
	}

	_, err = NewEthereumRPCReceiptObserver(map[valueobjects.NetworkID]*EthereumRPCObserverConfig{
		valueobjects.NetworkID("!!!"): {Endpoint: "https://rpc.example"},
	})
	if err == nil {
		t.Fatal("expected invalid network error")
	}
}
