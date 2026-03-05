package blockchain

import (
	"context"
	"errors"
	"fmt"

	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/value_objects"
)

type ChainRouterReceiptObserver struct {
	observers map[value_objects.ChainID]outport.ChainReceiptObserver
}

func NewChainRouterReceiptObserver(
	observers map[value_objects.ChainID]outport.ChainReceiptObserver,
) (*ChainRouterReceiptObserver, error) {
	if len(observers) == 0 {
		return nil, errors.New("at least one chain observer is required")
	}

	normalized := make(map[value_objects.ChainID]outport.ChainReceiptObserver, len(observers))
	for chain, observer := range observers {
		if observer == nil {
			return nil, fmt.Errorf("observer is not configured for chain: %s", chain)
		}
		normalizedChain, ok := value_objects.ParseChainID(string(chain))
		if !ok {
			return nil, fmt.Errorf("observer chain key is invalid: %s", chain)
		}
		normalized[normalizedChain] = observer
	}

	return &ChainRouterReceiptObserver{observers: normalized}, nil
}

func (o *ChainRouterReceiptObserver) ObserveAddress(
	ctx context.Context,
	input outport.ObserveChainPaymentAddressInput,
) (outport.ObservePaymentAddressOutput, error) {
	normalizedChain, ok := value_objects.ParseChainID(string(input.Chain))
	if !ok {
		return outport.ObservePaymentAddressOutput{}, errors.New("chain is invalid")
	}
	normalizedNetwork, ok := value_objects.ParseNetworkID(string(input.Network))
	if !ok {
		return outport.ObservePaymentAddressOutput{}, errors.New("network is invalid")
	}

	observer, found := o.observers[normalizedChain]
	if !found {
		return outport.ObservePaymentAddressOutput{}, fmt.Errorf("receipt observer is not configured for chain: %s", normalizedChain)
	}

	return observer.ObserveAddress(ctx, outport.ObservePaymentAddressInput{
		Network:               normalizedNetwork,
		Address:               input.Address,
		IssuedAt:              input.IssuedAt,
		RequiredConfirmations: input.RequiredConfirmations,
		SinceBlockHeight:      input.SinceBlockHeight,
	})
}
