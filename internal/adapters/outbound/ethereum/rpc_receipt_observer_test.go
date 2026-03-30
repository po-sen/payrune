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
	headersByNumber   map[string]map[string]any
	blocksByNumber    map[string]map[string]any
	statusCode        int
	rpcError          map[string]any
	lastAuthHeader    string
	requestedHeaders  []string
	requestedBlocks   []string
}

func newTestEthereumRPCServer(t *testing.T) (*testEthereumRPCServer, *httptest.Server) {
	t.Helper()

	handlerState := &testEthereumRPCServer{
		t:                 t,
		latestBlockHeight: "0x3",
		headersByNumber: map[string]map[string]any{
			"0x0": {"number": "0x0", "timestamp": "0x3e8"},
			"0x1": {"number": "0x1", "timestamp": "0x7d0"},
			"0x2": {"number": "0x2", "timestamp": "0xbb8"},
			"0x3": {"number": "0x3", "timestamp": "0xfa0"},
		},
		blocksByNumber: map[string]map[string]any{
			"0x0": {"number": "0x0", "timestamp": "0x3e8", "transactions": []map[string]any{}},
			"0x1": {"number": "0x1", "timestamp": "0x7d0", "transactions": []map[string]any{
				{"hash": "0xaaa", "to": "0x1111111111111111111111111111111111111111", "value": "0x1"},
			}},
			"0x2": {"number": "0x2", "timestamp": "0xbb8", "transactions": []map[string]any{
				{"hash": "0xbbb", "to": "0x1111111111111111111111111111111111111111", "value": "0x7"},
				{"hash": "0xccc", "to": "0x2222222222222222222222222222222222222222", "value": "0x5"},
			}},
			"0x3": {"number": "0x3", "timestamp": "0xfa0", "transactions": []map[string]any{
				{"hash": "0xddd", "to": "0x1111111111111111111111111111111111111111", "value": "0x3"},
				{"hash": "0xeee", "to": nil, "value": "0x8"},
			}},
		},
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
		case "eth_getBlockByNumber":
			if len(request.Params) != 2 {
				t.Fatalf("unexpected params for eth_getBlockByNumber: %d", len(request.Params))
			}

			var blockNumber string
			if err := json.Unmarshal(request.Params[0], &blockNumber); err != nil {
				t.Fatalf("decode block number: %v", err)
			}
			var fullTransactions bool
			if err := json.Unmarshal(request.Params[1], &fullTransactions); err != nil {
				t.Fatalf("decode fullTransactions: %v", err)
			}

			var result any
			if fullTransactions {
				handlerState.requestedBlocks = append(handlerState.requestedBlocks, blockNumber)
				result = handlerState.blocksByNumber[blockNumber]
			} else {
				handlerState.requestedHeaders = append(handlerState.requestedHeaders, blockNumber)
				result = handlerState.headersByNumber[blockNumber]
			}

			_ = json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      1,
				"result":  result,
			})
		default:
			t.Fatalf("unexpected rpc method: %s", request.Method)
		}
	}))

	return handlerState, server
}

func TestEthereumRPCReceiptObserverObserveAddress(t *testing.T) {
	state, server := newTestEthereumRPCServer(t)
	defer server.Close()

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

func TestEthereumRPCReceiptObserverObserveAddressNoBlocksAfterIssuedAt(t *testing.T) {
	_, server := newTestEthereumRPCServer(t)
	defer server.Close()

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
		IssuedAt:              time.Unix(5000, 0).UTC(),
		ObservedTotalMinor:    0,
		ConfirmedTotalMinor:   0,
		UnconfirmedTotalMinor: 0,
		RequiredConfirmations: 1,
		LatestBlockHeight:     3,
	})
	if err != nil {
		t.Fatalf("ObserveAddress returned error: %v", err)
	}

	if output.ObservedTotalMinor != 0 || output.ConfirmedTotalMinor != 0 || output.UnconfirmedTotalMinor != 0 {
		t.Fatalf("expected zero totals, got %+v", output)
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

func TestEthereumRPCReceiptObserverObserveAddressZeroTotalsUsesSinceBlockHeight(t *testing.T) {
	state, server := newTestEthereumRPCServer(t)
	defer server.Close()

	observer, err := NewEthereumRPCReceiptObserver(map[valueobjects.NetworkID]*EthereumRPCObserverConfig{
		valueobjects.NetworkID("sepolia"): {
			Endpoint: server.URL,
		},
	})
	if err != nil {
		t.Fatalf("NewEthereumRPCReceiptObserver returned error: %v", err)
	}

	output, err := observer.ObserveAddress(context.Background(), outport.ObservePaymentAddressInput{
		Network:               valueobjects.NetworkID("sepolia"),
		Address:               "0x1111111111111111111111111111111111111111",
		IssuedAt:              time.Unix(1000, 0).UTC(),
		ObservedTotalMinor:    0,
		ConfirmedTotalMinor:   0,
		UnconfirmedTotalMinor: 0,
		RequiredConfirmations: 2,
		LatestBlockHeight:     3,
		SinceBlockHeight:      2,
	})
	if err != nil {
		t.Fatalf("ObserveAddress returned error: %v", err)
	}

	if output.ObservedTotalMinor != 3 || output.ConfirmedTotalMinor != 0 || output.UnconfirmedTotalMinor != 3 {
		t.Fatalf("unexpected incremental totals: got %+v", output)
	}
	if got := strings.Join(state.requestedBlocks, ","); got != "0x3" {
		t.Fatalf("expected to scan only block 0x3, got %q", got)
	}
}

func TestEthereumRPCReceiptObserverObserveAddressReusesTotalsWhenAlreadyAtLatest(t *testing.T) {
	state, server := newTestEthereumRPCServer(t)
	defer server.Close()

	observer, err := NewEthereumRPCReceiptObserver(map[valueobjects.NetworkID]*EthereumRPCObserverConfig{
		valueobjects.NetworkID("sepolia"): {
			Endpoint: server.URL,
		},
	})
	if err != nil {
		t.Fatalf("NewEthereumRPCReceiptObserver returned error: %v", err)
	}

	output, err := observer.ObserveAddress(context.Background(), outport.ObservePaymentAddressInput{
		Network:               valueobjects.NetworkID("sepolia"),
		Address:               "0x1111111111111111111111111111111111111111",
		IssuedAt:              time.Unix(1000, 0).UTC(),
		ObservedTotalMinor:    7,
		ConfirmedTotalMinor:   7,
		UnconfirmedTotalMinor: 0,
		RequiredConfirmations: 2,
		LatestBlockHeight:     3,
		SinceBlockHeight:      3,
	})
	if err != nil {
		t.Fatalf("ObserveAddress returned error: %v", err)
	}

	if output.ObservedTotalMinor != 7 || output.ConfirmedTotalMinor != 7 || output.UnconfirmedTotalMinor != 0 {
		t.Fatalf("unexpected reused totals: got %+v", output)
	}
	if len(state.requestedBlocks) != 0 || len(state.requestedHeaders) != 0 {
		t.Fatalf("expected no block fetches when already at latest, got blocks=%v headers=%v", state.requestedBlocks, state.requestedHeaders)
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
