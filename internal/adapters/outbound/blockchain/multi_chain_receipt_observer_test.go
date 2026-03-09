package blockchain

import (
	"context"
	"errors"
	"testing"

	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/value_objects"
)

type fakeChainObserver struct {
	output               outport.ObservePaymentAddressOutput
	err                  error
	latestBlockHeight    int64
	latestBlockHeightErr error
	lastInput            outport.ObservePaymentAddressInput
	lastNetwork          value_objects.NetworkID
	calls                int
	fetchCalls           int
}

func (f *fakeChainObserver) ObserveAddress(
	_ context.Context,
	input outport.ObservePaymentAddressInput,
) (outport.ObservePaymentAddressOutput, error) {
	f.calls++
	f.lastInput = input
	if f.err != nil {
		return outport.ObservePaymentAddressOutput{}, f.err
	}
	return f.output, nil
}

func (f *fakeChainObserver) FetchLatestBlockHeight(
	_ context.Context,
	network value_objects.NetworkID,
) (int64, error) {
	f.fetchCalls++
	f.lastNetwork = network
	if f.latestBlockHeightErr != nil {
		return 0, f.latestBlockHeightErr
	}
	if f.latestBlockHeight > 0 {
		return f.latestBlockHeight, nil
	}
	return 1, nil
}

func TestNewMultiChainReceiptObserverValidation(t *testing.T) {
	_, err := NewMultiChainReceiptObserver(nil)
	if err == nil {
		t.Fatal("expected error for empty observer map")
	}

	_, err = NewMultiChainReceiptObserver(map[value_objects.ChainID]outport.ChainReceiptObserver{
		value_objects.ChainIDBitcoin: nil,
	})
	if err == nil {
		t.Fatal("expected error for nil observer")
	}

	_, err = NewMultiChainReceiptObserver(map[value_objects.ChainID]outport.ChainReceiptObserver{
		value_objects.ChainID("eth/mainnet"): &fakeChainObserver{},
	})
	if err == nil {
		t.Fatal("expected error for invalid chain key")
	}
}

func TestMultiChainReceiptObserverObserveAddress(t *testing.T) {
	bitcoinObserver := &fakeChainObserver{
		output: outport.ObservePaymentAddressOutput{
			ObservedTotalMinor: 123,
		},
	}
	router, err := NewMultiChainReceiptObserver(map[value_objects.ChainID]outport.ChainReceiptObserver{
		value_objects.ChainIDBitcoin: bitcoinObserver,
	})
	if err != nil {
		t.Fatalf("setup router: %v", err)
	}

	output, err := router.ObserveAddress(context.Background(), outport.ObserveChainPaymentAddressInput{
		Chain:                 value_objects.ChainID("BitCoin"),
		Network:               value_objects.NetworkID("testnet4"),
		Address:               "tb1qexample",
		RequiredConfirmations: 1,
		LatestBlockHeight:     222,
	})
	if err != nil {
		t.Fatalf("ObserveAddress returned error: %v", err)
	}
	if output.ObservedTotalMinor != 123 {
		t.Fatalf("unexpected observed total: got %d", output.ObservedTotalMinor)
	}
	if bitcoinObserver.calls != 1 {
		t.Fatalf("unexpected observer calls: got %d", bitcoinObserver.calls)
	}
	if bitcoinObserver.lastInput.Network != value_objects.NetworkID("testnet4") {
		t.Fatalf("unexpected normalized network: got %q", bitcoinObserver.lastInput.Network)
	}
	if bitcoinObserver.lastInput.LatestBlockHeight != 222 {
		t.Fatalf("unexpected forwarded latest block height: got %d", bitcoinObserver.lastInput.LatestBlockHeight)
	}
}

func TestMultiChainReceiptObserverFetchLatestBlockHeight(t *testing.T) {
	bitcoinObserver := &fakeChainObserver{latestBlockHeight: 321}
	router, err := NewMultiChainReceiptObserver(map[value_objects.ChainID]outport.ChainReceiptObserver{
		value_objects.ChainIDBitcoin: bitcoinObserver,
	})
	if err != nil {
		t.Fatalf("setup router: %v", err)
	}

	latestBlockHeight, err := router.FetchLatestBlockHeight(
		context.Background(),
		value_objects.ChainID("BitCoin"),
		value_objects.NetworkID("testnet4"),
	)
	if err != nil {
		t.Fatalf("FetchLatestBlockHeight returned error: %v", err)
	}
	if latestBlockHeight != 321 {
		t.Fatalf("unexpected latest block height: got %d", latestBlockHeight)
	}
	if bitcoinObserver.fetchCalls != 1 {
		t.Fatalf("unexpected fetch call count: got %d", bitcoinObserver.fetchCalls)
	}
	if bitcoinObserver.lastNetwork != value_objects.NetworkID("testnet4") {
		t.Fatalf("unexpected normalized network: got %q", bitcoinObserver.lastNetwork)
	}
}

func TestMultiChainReceiptObserverObserveAddressValidation(t *testing.T) {
	bitcoinObserver := &fakeChainObserver{}
	router, err := NewMultiChainReceiptObserver(map[value_objects.ChainID]outport.ChainReceiptObserver{
		value_objects.ChainIDBitcoin: bitcoinObserver,
	})
	if err != nil {
		t.Fatalf("setup router: %v", err)
	}

	_, err = router.ObserveAddress(context.Background(), outport.ObserveChainPaymentAddressInput{
		Chain:                 value_objects.ChainID("eth/mainnet"),
		Network:               value_objects.NetworkID("testnet4"),
		Address:               "tb1qexample",
		RequiredConfirmations: 1,
	})
	if err == nil {
		t.Fatal("expected invalid chain error")
	}

	_, err = router.ObserveAddress(context.Background(), outport.ObserveChainPaymentAddressInput{
		Chain:                 value_objects.ChainIDBitcoin,
		Network:               value_objects.NetworkID("main/net"),
		Address:               "tb1qexample",
		RequiredConfirmations: 1,
	})
	if err == nil {
		t.Fatal("expected invalid network error")
	}

	_, err = router.ObserveAddress(context.Background(), outport.ObserveChainPaymentAddressInput{
		Chain:                 value_objects.ChainID("ethereum"),
		Network:               value_objects.NetworkID("mainnet"),
		Address:               "0x123",
		RequiredConfirmations: 1,
	})
	if err == nil {
		t.Fatal("expected missing observer error")
	}
}

func TestMultiChainReceiptObserverObserveAddressPassThroughError(t *testing.T) {
	bitcoinObserver := &fakeChainObserver{err: errors.New("boom")}
	router, err := NewMultiChainReceiptObserver(map[value_objects.ChainID]outport.ChainReceiptObserver{
		value_objects.ChainIDBitcoin: bitcoinObserver,
	})
	if err != nil {
		t.Fatalf("setup router: %v", err)
	}

	_, err = router.ObserveAddress(context.Background(), outport.ObserveChainPaymentAddressInput{
		Chain:                 value_objects.ChainIDBitcoin,
		Network:               value_objects.NetworkID("mainnet"),
		Address:               "bc1qexample",
		RequiredConfirmations: 1,
	})
	if err == nil {
		t.Fatal("expected downstream observer error")
	}
}
