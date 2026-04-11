package ethereum

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

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
