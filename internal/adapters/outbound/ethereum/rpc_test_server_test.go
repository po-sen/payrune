package ethereum

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testEthereumRPCServer struct {
	t                      *testing.T
	latestBlockHeight      string
	codesByAddress         map[string]string
	balancesByKey          map[string]string
	tokenBalancesByKey     map[string]string
	tokenDecimalsByAddress map[string]string
	statusCode             int
	rpcError               map[string]any
	lastAuthHeader         string
	requestedCodes         []string
	requestedBalances      []string
	requestedTokenBalances []string
	requestedTokenDecimals []string
	requestedBlocks        []string
}

func newTestEthereumRPCServer(t *testing.T) (*testEthereumRPCServer, *httptest.Server) {
	t.Helper()

	handlerState := &testEthereumRPCServer{
		t:                      t,
		latestBlockHeight:      "0x3",
		codesByAddress:         make(map[string]string),
		balancesByKey:          make(map[string]string),
		tokenBalancesByKey:     make(map[string]string),
		tokenDecimalsByAddress: make(map[string]string),
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
		case "eth_getCode":
			if len(request.Params) != 2 {
				t.Fatalf("unexpected params for eth_getCode: %d", len(request.Params))
			}

			var address string
			if err := json.Unmarshal(request.Params[0], &address); err != nil {
				t.Fatalf("decode code address: %v", err)
			}
			normalizedAddress := strings.ToLower(strings.TrimSpace(address))
			handlerState.requestedCodes = append(handlerState.requestedCodes, normalizedAddress)

			result, ok := handlerState.codesByAddress[normalizedAddress]
			if !ok {
				result = "0x"
			}

			_ = json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      1,
				"result":  result,
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
			switch selector, err := decodeCallSelector(callObject.Data); {
			case err != nil:
				t.Fatalf("decode eth_call selector: %v", err)
			case selector == "70a08231":
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
			case selector == "313ce567":
				handlerState.requestedTokenDecimals = append(handlerState.requestedTokenDecimals, assetReference)

				result, ok := handlerState.tokenDecimalsByAddress[assetReference]
				if !ok {
					result = "0x0"
				}

				_ = json.NewEncoder(w).Encode(map[string]any{
					"jsonrpc": "2.0",
					"id":      1,
					"result":  result,
				})
			default:
				t.Fatalf("unexpected eth_call selector: %s", selector)
			}
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

func decodeCallSelector(callData string) (string, error) {
	trimmed := strings.TrimSpace(callData)
	if !strings.HasPrefix(trimmed, "0x") || len(trimmed) < 10 {
		return "", errors.New("call data is invalid")
	}
	return strings.ToLower(trimmed[2:10]), nil
}
