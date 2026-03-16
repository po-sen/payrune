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

func TestBlockscoutReceiptObserverFetchLatestBlockHeight(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/eth-rpc" {
			t.Fatalf("unexpected path: got %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  "0x64",
		})
	}))
	defer server.Close()

	observer, err := NewBlockscoutReceiptObserver(map[valueobjects.NetworkID]*BlockscoutObserverConfig{
		valueobjects.NetworkID("sepolia"): {
			BaseURL: server.URL,
			Timeout: time.Second,
		},
	})
	if err != nil {
		t.Fatalf("NewBlockscoutReceiptObserver returned error: %v", err)
	}

	latestBlockHeight, err := observer.FetchLatestBlockHeight(context.Background(), valueobjects.NetworkID("sepolia"))
	if err != nil {
		t.Fatalf("FetchLatestBlockHeight returned error: %v", err)
	}
	if latestBlockHeight != 100 {
		t.Fatalf("unexpected latest block height: got %d", latestBlockHeight)
	}
}

func TestBlockscoutReceiptObserverObserveNativeAddress(t *testing.T) {
	address := "0x1111111111111111111111111111111111111111"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api" {
			t.Fatalf("unexpected path: got %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("action"); got != "txlist" {
			t.Fatalf("unexpected action: got %q", got)
		}
		if got := r.URL.Query().Get("address"); got != address {
			t.Fatalf("unexpected address: got %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "1",
			"message": "OK",
			"result": []map[string]any{
				{
					"hash":          "0xaaa",
					"timeStamp":     "1710000100",
					"blockNumber":   "90",
					"to":            address,
					"value":         "100",
					"confirmations": "12",
					"isError":       "0",
				},
				{
					"hash":          "0xbbb",
					"timeStamp":     "1710000200",
					"blockNumber":   "99",
					"to":            strings.ToUpper(address),
					"value":         "50",
					"confirmations": "1",
					"isError":       "0",
				},
				{
					"hash":          "0xccc",
					"timeStamp":     "1709999000",
					"blockNumber":   "80",
					"to":            address,
					"value":         "25",
					"confirmations": "20",
					"isError":       "0",
				},
				{
					"hash":          "0xddd",
					"timeStamp":     "1710000300",
					"blockNumber":   "99",
					"to":            "0x2222222222222222222222222222222222222222",
					"value":         "77",
					"confirmations": "1",
					"isError":       "0",
				},
			},
		})
	}))
	defer server.Close()

	observer, err := NewBlockscoutReceiptObserver(map[valueobjects.NetworkID]*BlockscoutObserverConfig{
		valueobjects.NetworkID("mainnet"): {
			BaseURL: server.URL,
			Timeout: time.Second,
		},
	})
	if err != nil {
		t.Fatalf("NewBlockscoutReceiptObserver returned error: %v", err)
	}

	output, err := observer.ObserveAddress(context.Background(), outport.ObservePaymentAddressInput{
		Network:               valueobjects.NetworkID("mainnet"),
		Address:               address,
		AssetCode:             "eth",
		AssetType:             "native",
		IssuedAt:              time.Unix(1710000000, 0).UTC(),
		RequiredConfirmations: 2,
		LatestBlockHeight:     100,
	})
	if err != nil {
		t.Fatalf("ObserveAddress returned error: %v", err)
	}
	if output.ConfirmedTotalMinor != 100 {
		t.Fatalf("unexpected confirmed total: got %d", output.ConfirmedTotalMinor)
	}
	if output.UnconfirmedTotalMinor != 50 {
		t.Fatalf("unexpected unconfirmed total: got %d", output.UnconfirmedTotalMinor)
	}
	if output.ObservedTotalMinor != 150 {
		t.Fatalf("unexpected observed total: got %d", output.ObservedTotalMinor)
	}
}

func TestBlockscoutReceiptObserverObserveTokenAddress(t *testing.T) {
	address := "0x1111111111111111111111111111111111111111"
	tokenAddress := "0x9999999999999999999999999999999999999999"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api" {
			t.Fatalf("unexpected path: got %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("action"); got != "tokentx" {
			t.Fatalf("unexpected action: got %q", got)
		}
		if got := r.URL.Query().Get("contractaddress"); got != tokenAddress {
			t.Fatalf("unexpected contract address: got %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "1",
			"message": "OK",
			"result": []map[string]any{
				{
					"hash":            "0xaaa",
					"timeStamp":       "1710000100",
					"blockNumber":     "90",
					"to":              address,
					"value":           "1000000",
					"confirmations":   "12",
					"contractAddress": tokenAddress,
				},
				{
					"hash":            "0xbbb",
					"timeStamp":       "1710000200",
					"blockNumber":     "99",
					"to":              address,
					"value":           "500000",
					"confirmations":   "1",
					"contractAddress": tokenAddress,
				},
			},
		})
	}))
	defer server.Close()

	observer, err := NewBlockscoutReceiptObserver(map[valueobjects.NetworkID]*BlockscoutObserverConfig{
		valueobjects.NetworkID("sepolia"): {
			BaseURL: server.URL,
			Timeout: time.Second,
		},
	})
	if err != nil {
		t.Fatalf("NewBlockscoutReceiptObserver returned error: %v", err)
	}

	output, err := observer.ObserveAddress(context.Background(), outport.ObservePaymentAddressInput{
		Network:               valueobjects.NetworkID("sepolia"),
		Address:               address,
		AssetCode:             "usdt",
		AssetType:             "erc20",
		TokenAddress:          tokenAddress,
		IssuedAt:              time.Unix(1710000000, 0).UTC(),
		RequiredConfirmations: 2,
		LatestBlockHeight:     100,
	})
	if err != nil {
		t.Fatalf("ObserveAddress returned error: %v", err)
	}
	if output.ConfirmedTotalMinor != 1000000 {
		t.Fatalf("unexpected confirmed total: got %d", output.ConfirmedTotalMinor)
	}
	if output.UnconfirmedTotalMinor != 500000 {
		t.Fatalf("unexpected unconfirmed total: got %d", output.UnconfirmedTotalMinor)
	}
}

func TestBlockscoutReceiptObserverRejectsMissingTokenAddress(t *testing.T) {
	observer, err := NewBlockscoutReceiptObserver(map[valueobjects.NetworkID]*BlockscoutObserverConfig{
		valueobjects.NetworkID("sepolia"): {
			BaseURL: "https://eth-sepolia.blockscout.com",
			Timeout: time.Second,
		},
	})
	if err != nil {
		t.Fatalf("NewBlockscoutReceiptObserver returned error: %v", err)
	}

	_, err = observer.ObserveAddress(context.Background(), outport.ObservePaymentAddressInput{
		Network:               valueobjects.NetworkID("sepolia"),
		Address:               "0x1111111111111111111111111111111111111111",
		AssetCode:             "usdt",
		AssetType:             "erc20",
		IssuedAt:              time.Unix(1710000000, 0).UTC(),
		RequiredConfirmations: 2,
		LatestBlockHeight:     100,
	})
	if err == nil || !strings.Contains(err.Error(), "token address is invalid") {
		t.Fatalf("unexpected error: got %v", err)
	}
}
