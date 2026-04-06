package ethereum

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

type testEthereumRPCServer struct {
	t                      *testing.T
	latestBlockHeight      string
	balancesByKey          map[string]string
	tokenBalancesByKey     map[string]string
	statusCode             int
	rpcError               map[string]any
	lastAuthHeader         string
	requestedBalances      []string
	requestedTokenBalances []string
	requestedBlocks        []string
}

func newTestEthereumRPCServer(t *testing.T) (*testEthereumRPCServer, *httptest.Server) {
	t.Helper()

	handlerState := &testEthereumRPCServer{
		t:                  t,
		latestBlockHeight:  "0x3",
		balancesByKey:      make(map[string]string),
		tokenBalancesByKey: make(map[string]string),
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
		case "eth_call":
			if len(request.Params) != 2 {
				t.Fatalf("unexpected params for eth_call: %d", len(request.Params))
			}

			var callObject struct {
				To   string `json:"to"`
				Data string `json:"data"`
			}
			if err := json.Unmarshal(request.Params[0], &callObject); err != nil {
				t.Fatalf("decode call object: %v", err)
			}
			var blockNumber string
			if err := json.Unmarshal(request.Params[1], &blockNumber); err != nil {
				t.Fatalf("decode call block number: %v", err)
			}

			assetReference := strings.ToLower(strings.TrimSpace(callObject.To))
			address, err := decodeERC20BalanceOfCallAddress(callObject.Data)
			if err != nil {
				t.Fatalf("decode erc20 balanceOf call: %v", err)
			}
			key := ethereumTokenBalanceKey(assetReference, address, blockNumber)
			handlerState.requestedTokenBalances = append(handlerState.requestedTokenBalances, key)

			result, ok := handlerState.tokenBalancesByKey[key]
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

func ethereumTokenBalanceKey(assetReference string, address string, blockNumber string) string {
	return strings.ToLower(strings.TrimSpace(assetReference)) +
		":" +
		strings.ToLower(strings.TrimSpace(address)) +
		"@" +
		strings.ToLower(strings.TrimSpace(blockNumber))
}

func TestEthereumRPCReceiptObserverObserveAddress(t *testing.T) {
	state, server := newTestEthereumRPCServer(t)
	defer server.Close()

	state.balancesByKey[ethereumBalanceKey("0x1111111111111111111111111111111111111111", "0x3")] = "0xa"
	state.balancesByKey[ethereumBalanceKey("0x1111111111111111111111111111111111111111", "0x2")] = "0x7"

	observer, err := NewEthereumRPCReceiptObserver(map[valueobjects.NetworkID]*EthereumRPCObserverConfig{
		valueobjects.NetworkIDMainnet: {
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
		Network:               valueobjects.NetworkIDMainnet,
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
		valueobjects.NetworkIDMainnet: {
			Endpoint: server.URL,
		},
	})
	if err != nil {
		t.Fatalf("NewEthereumRPCReceiptObserver returned error: %v", err)
	}

	output, err := observer.ObserveAddress(context.Background(), outport.ObservePaymentAddressInput{
		Network:               valueobjects.NetworkIDMainnet,
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

func TestEthereumRPCReceiptObserverObserveAddressERC20(t *testing.T) {
	state, server := newTestEthereumRPCServer(t)
	defer server.Close()

	assetReference := "0xdac17f958d2ee523a2206206994597c13d831ec7"
	receiverAddress := "0x1111111111111111111111111111111111111111"
	state.tokenBalancesByKey[ethereumTokenBalanceKey(assetReference, receiverAddress, "0x3")] = "0xf4240"
	state.tokenBalancesByKey[ethereumTokenBalanceKey(assetReference, receiverAddress, "0x2")] = "0xc3500"

	observer, err := NewEthereumRPCReceiptObserver(map[valueobjects.NetworkID]*EthereumRPCObserverConfig{
		valueobjects.NetworkIDMainnet: {
			Endpoint: server.URL,
		},
	})
	if err != nil {
		t.Fatalf("NewEthereumRPCReceiptObserver returned error: %v", err)
	}

	output, err := observer.ObserveAddress(context.Background(), outport.ObservePaymentAddressInput{
		AssetReference:        strings.ToUpper(assetReference),
		Network:               valueobjects.NetworkIDMainnet,
		Address:               receiverAddress,
		IssuedAt:              time.Unix(2500, 0).UTC(),
		ObservedTotalMinor:    0,
		ConfirmedTotalMinor:   0,
		UnconfirmedTotalMinor: 0,
		RequiredConfirmations: 2,
		LatestBlockHeight:     3,
	})
	if err != nil {
		t.Fatalf("ObserveAddress returned error: %v", err)
	}

	if output.ObservedTotalMinor != 1000000 {
		t.Fatalf("unexpected observed total: got %d", output.ObservedTotalMinor)
	}
	if output.ConfirmedTotalMinor != 800000 {
		t.Fatalf("unexpected confirmed total: got %d", output.ConfirmedTotalMinor)
	}
	if output.UnconfirmedTotalMinor != 200000 {
		t.Fatalf("unexpected unconfirmed total: got %d", output.UnconfirmedTotalMinor)
	}
	if got := strings.Join(state.requestedTokenBalances, ","); got != ethereumTokenBalanceKey(assetReference, receiverAddress, "0x3")+","+ethereumTokenBalanceKey(assetReference, receiverAddress, "0x2") {
		t.Fatalf("unexpected token balance requests: got %q", got)
	}
	if len(state.requestedBalances) != 0 {
		t.Fatalf("expected no native balance requests, got %v", state.requestedBalances)
	}
}

func TestEthereumRPCReceiptObserverObserveAddressRejectsInconsistentBalances(t *testing.T) {
	state, server := newTestEthereumRPCServer(t)
	defer server.Close()

	state.balancesByKey[ethereumBalanceKey("0x1111111111111111111111111111111111111111", "0x3")] = "0x5"
	state.balancesByKey[ethereumBalanceKey("0x1111111111111111111111111111111111111111", "0x2")] = "0x7"

	observer, err := NewEthereumRPCReceiptObserver(map[valueobjects.NetworkID]*EthereumRPCObserverConfig{
		valueobjects.NetworkIDMainnet: {
			Endpoint: server.URL,
		},
	})
	if err != nil {
		t.Fatalf("NewEthereumRPCReceiptObserver returned error: %v", err)
	}

	_, err = observer.ObserveAddress(context.Background(), outport.ObservePaymentAddressInput{
		Network:               valueobjects.NetworkIDMainnet,
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
		valueobjects.NetworkIDSepolia: {
			Endpoint: server.URL,
		},
	})
	if err != nil {
		t.Fatalf("NewEthereumRPCReceiptObserver returned error: %v", err)
	}

	height, err := observer.FetchLatestBlockHeight(context.Background(), valueobjects.NetworkIDSepolia)
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
		valueobjects.NetworkIDMainnet: {Endpoint: server.URL},
	})
	if err != nil {
		t.Fatalf("NewEthereumRPCReceiptObserver returned error: %v", err)
	}

	_, err = observer.ObserveAddress(context.Background(), outport.ObservePaymentAddressInput{
		Network:               valueobjects.NetworkIDMainnet,
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

	_, err = observer.ObserveAddress(context.Background(), outport.ObservePaymentAddressInput{
		AssetReference:        " ",
		Network:               valueobjects.NetworkIDMainnet,
		Address:               "0x1111111111111111111111111111111111111111",
		IssuedAt:              time.Unix(2500, 0).UTC(),
		RequiredConfirmations: 1,
		LatestBlockHeight:     3,
	})
	if err != nil {
		t.Fatalf("expected blank asset reference to use native path, got %v", err)
	}
}

func TestEthereumRPCReceiptObserverEndpointError(t *testing.T) {
	state, server := newTestEthereumRPCServer(t)
	state.statusCode = http.StatusBadGateway
	defer server.Close()

	observer, err := NewEthereumRPCReceiptObserver(map[valueobjects.NetworkID]*EthereumRPCObserverConfig{
		valueobjects.NetworkIDMainnet: {Endpoint: server.URL},
	})
	if err != nil {
		t.Fatalf("NewEthereumRPCReceiptObserver returned error: %v", err)
	}

	_, err = observer.FetchLatestBlockHeight(context.Background(), valueobjects.NetworkIDMainnet)
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

func decodeERC20BalanceOfCallAddress(callData string) (string, error) {
	trimmed := strings.TrimSpace(callData)
	if !strings.HasPrefix(trimmed, "0x") || len(trimmed) != 2+8+64 {
		return "", errors.New("call data is invalid")
	}
	if strings.ToLower(trimmed[2:10]) != "70a08231" {
		return "", errors.New("call selector is invalid")
	}
	return "0x" + strings.ToLower(trimmed[len(trimmed)-40:]), nil
}
