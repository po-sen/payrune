package blockchain

import (
	"context"
	"errors"
	"fmt"

	outport "payrune/internal/application/ports/outbound"
)

type MultiChainReceiptObserver struct {
	observers map[string]outport.ChainReceiptObserver
}

func NewMultiChainReceiptObserver(
	observers map[string]outport.ChainReceiptObserver,
) (*MultiChainReceiptObserver, error) {
	if len(observers) == 0 {
		return nil, errors.New("at least one chain observer is required")
	}

	normalized := make(map[string]outport.ChainReceiptObserver, len(observers))
	for chain, observer := range observers {
		if observer == nil {
			return nil, fmt.Errorf("observer is not configured for chain: %s", chain)
		}
		normalizedChain, ok := outport.NormalizeChainID(chain)
		if !ok {
			return nil, fmt.Errorf("observer chain key is invalid: %s", chain)
		}
		normalized[normalizedChain] = observer
	}

	return &MultiChainReceiptObserver{observers: normalized}, nil
}

func (o *MultiChainReceiptObserver) ObserveAddress(
	ctx context.Context,
	input outport.ObserveChainPaymentAddressInput,
) (outport.ObservePaymentAddressOutput, error) {
	observer, normalizedNetwork, err := o.resolveObserver(input.Chain, input.Network)
	if err != nil {
		return outport.ObservePaymentAddressOutput{}, err
	}

	output, err := observer.ObserveAddress(ctx, outport.ObservePaymentAddressInput{
		AssetReference:        input.AssetReference,
		Network:               normalizedNetwork,
		Address:               input.Address,
		IssuedAt:              input.IssuedAt,
		ObservedTotalMinor:    input.ObservedTotalMinor,
		ConfirmedTotalMinor:   input.ConfirmedTotalMinor,
		UnconfirmedTotalMinor: input.UnconfirmedTotalMinor,
		RequiredConfirmations: input.RequiredConfirmations,
		LatestBlockHeight:     input.LatestBlockHeight,
		SinceBlockHeight:      input.SinceBlockHeight,
	})
	if err != nil {
		switch {
		case errors.Is(err, outport.ErrBlockchainReceiptObserverNotConfigured):
			return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverNotConfigured
		case errors.Is(err, outport.ErrBlockchainReceiptObserverInputInvalid):
			return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverInputInvalid
		default:
			return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverFailed
		}
	}
	return output, nil
}

func (o *MultiChainReceiptObserver) FetchLatestBlockHeight(
	ctx context.Context,
	chain string,
	network string,
) (int64, error) {
	observer, normalizedNetwork, err := o.resolveObserver(chain, network)
	if err != nil {
		return 0, err
	}
	latestBlockHeight, err := observer.FetchLatestBlockHeight(ctx, normalizedNetwork)
	if err != nil {
		switch {
		case errors.Is(err, outport.ErrBlockchainReceiptObserverNotConfigured):
			return 0, outport.ErrBlockchainReceiptObserverNotConfigured
		case errors.Is(err, outport.ErrBlockchainReceiptObserverInputInvalid):
			return 0, outport.ErrBlockchainReceiptObserverInputInvalid
		default:
			return 0, outport.ErrBlockchainReceiptObserverFailed
		}
	}
	return latestBlockHeight, nil
}

func (o *MultiChainReceiptObserver) resolveObserver(
	chain string,
	network string,
) (outport.ChainReceiptObserver, string, error) {
	normalizedChain, ok := outport.NormalizeChainID(chain)
	if !ok {
		return nil, "", outport.ErrBlockchainReceiptObserverInputInvalid
	}
	normalizedNetwork, ok := outport.NormalizeNetworkID(network)
	if !ok {
		return nil, "", outport.ErrBlockchainReceiptObserverInputInvalid
	}

	observer, found := o.observers[normalizedChain]
	if !found {
		return nil, "", outport.ErrBlockchainReceiptObserverNotConfigured
	}
	return observer, normalizedNetwork, nil
}
