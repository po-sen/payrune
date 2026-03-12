package bitcoin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

func TestBitcoinEsploraReceiptObserverObserveAddressIssueTimeScoped(t *testing.T) {
	issuedAt := time.Date(2026, 3, 5, 13, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/address/tb1qexample/txs/chain":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{
					"txid": "old-confirmed",
					"vout": []map[string]any{
						{"scriptpubkey_address": "tb1qexample", "value": 10000},
					},
					"status": map[string]any{
						"confirmed":    true,
						"block_height": 99,
						"block_time":   issuedAt.Add(-1 * time.Minute).Unix(),
					},
				},
				{
					"txid": "new-unconfirmed-by-confirmation",
					"vout": []map[string]any{
						{"scriptpubkey_address": "tb1qexample", "value": 20000},
					},
					"status": map[string]any{
						"confirmed":    true,
						"block_height": 100,
						"block_time":   issuedAt.Add(1 * time.Minute).Unix(),
					},
				},
				{
					"txid": "new-confirmed",
					"vout": []map[string]any{
						{"scriptpubkey_address": "tb1qexample", "value": 30000},
						{"scriptpubkey_address": "tb1qother", "value": 999},
					},
					"status": map[string]any{
						"confirmed":    true,
						"block_height": 98,
						"block_time":   issuedAt.Add(2 * time.Minute).Unix(),
					},
				},
			})
			return
		case "/address/tb1qexample/txs/chain/new-confirmed":
			_ = json.NewEncoder(w).Encode([]map[string]any{})
			return
		case "/address/tb1qexample/txs/mempool":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{
					"txid": "mempool-receipt",
					"vout": []map[string]any{
						{"scriptpubkey_address": "tb1qexample", "value": 40000},
					},
					"status": map[string]any{
						"confirmed": false,
					},
				},
			})
			return
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	observer, err := NewBitcoinEsploraReceiptObserver(
		map[valueobjects.NetworkID]*BitcoinEsploraObserverConfig{
			valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4): {
				Endpoint: server.URL,
				Timeout:  3 * time.Second,
			},
		},
	)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}

	output, err := observer.ObserveAddress(context.Background(), outport.ObservePaymentAddressInput{
		Network:               valueobjects.NetworkID("testnet4"),
		Address:               "tb1qexample",
		IssuedAt:              issuedAt,
		RequiredConfirmations: 2,
		LatestBlockHeight:     100,
	})
	if err != nil {
		t.Fatalf("observe address error: %v", err)
	}
	if output.ObservedTotalMinor != 90000 {
		t.Fatalf("unexpected observed total: got %d", output.ObservedTotalMinor)
	}
	if output.ConfirmedTotalMinor != 30000 {
		t.Fatalf("unexpected confirmed total: got %d", output.ConfirmedTotalMinor)
	}
	if output.UnconfirmedTotalMinor != 60000 {
		t.Fatalf("unexpected unconfirmed total: got %d", output.UnconfirmedTotalMinor)
	}
	if output.LatestBlockHeight != 100 {
		t.Fatalf("unexpected latest block height: got %d", output.LatestBlockHeight)
	}
}

func TestBitcoinEsploraReceiptObserverMissingNetworkEndpoint(t *testing.T) {
	observer, err := NewBitcoinEsploraReceiptObserver(
		map[valueobjects.NetworkID]*BitcoinEsploraObserverConfig{
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet): {Endpoint: "http://127.0.0.1:18443"},
		},
	)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}

	_, err = observer.ObserveAddress(context.Background(), outport.ObservePaymentAddressInput{
		Network:               valueobjects.NetworkID("testnet4"),
		Address:               "tb1qexample",
		IssuedAt:              time.Date(2026, 3, 5, 13, 0, 0, 0, time.UTC),
		RequiredConfirmations: 1,
		LatestBlockHeight:     100,
	})
	if err == nil {
		t.Fatal("expected network endpoint error but got nil")
	}
}

func TestBitcoinEsploraReceiptObserverEndpointError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer server.Close()

	observer, err := NewBitcoinEsploraReceiptObserver(
		map[valueobjects.NetworkID]*BitcoinEsploraObserverConfig{
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet): {Endpoint: server.URL},
		},
	)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}

	_, err = observer.FetchLatestBlockHeight(context.Background(), valueobjects.NetworkID("mainnet"))
	if err == nil {
		t.Fatal("expected endpoint error but got nil")
	}
}

func TestBitcoinEsploraReceiptObserverValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	observer, err := NewBitcoinEsploraReceiptObserver(
		map[valueobjects.NetworkID]*BitcoinEsploraObserverConfig{
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet): {Endpoint: server.URL},
		},
	)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}

	_, err = observer.ObserveAddress(context.Background(), outport.ObservePaymentAddressInput{
		Network:               valueobjects.NetworkID("mainnet"),
		Address:               "bc1qexample",
		RequiredConfirmations: 1,
	})
	if err == nil {
		t.Fatal("expected missing issued at error")
	}
}

func TestBitcoinEsploraReceiptObserverFetchLatestBlockHeight(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/blocks/tip/height" {
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte("321"))
	}))
	defer server.Close()

	observer, err := NewBitcoinEsploraReceiptObserver(
		map[valueobjects.NetworkID]*BitcoinEsploraObserverConfig{
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet): {Endpoint: server.URL},
		},
	)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}

	latestBlockHeight, err := observer.FetchLatestBlockHeight(context.Background(), valueobjects.NetworkID("mainnet"))
	if err != nil {
		t.Fatalf("FetchLatestBlockHeight returned error: %v", err)
	}
	if latestBlockHeight != 321 {
		t.Fatalf("unexpected latest block height: got %d", latestBlockHeight)
	}
}

func TestNewBitcoinEsploraReceiptObserverValidation(t *testing.T) {
	_, err := NewBitcoinEsploraReceiptObserver(nil)
	if err == nil {
		t.Fatal("expected constructor validation error but got nil")
	}
}

func TestNewBitcoinEsploraReceiptObserverUnknownConfiguredNetwork(t *testing.T) {
	_, err := NewBitcoinEsploraReceiptObserver(
		map[valueobjects.NetworkID]*BitcoinEsploraObserverConfig{
			valueobjects.NetworkID("unknown"): {Endpoint: "https://example.com/api"},
		},
	)
	if err == nil {
		t.Fatal("expected unknown network constructor validation error but got nil")
	}
}
