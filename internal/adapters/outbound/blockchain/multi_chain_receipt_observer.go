package blockchain

import (
	"context"
	"errors"
	"fmt"

	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/value_objects"
)

type MultiChainReceiptObserver struct {
	observers map[value_objects.ChainID]outport.ChainReceiptObserver
}

func NewMultiChainReceiptObserver(
	observers map[value_objects.ChainID]outport.ChainReceiptObserver,
) (*MultiChainReceiptObserver, error) {
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

	return observer.ObserveAddress(ctx, outport.ObservePaymentAddressInput{
		Network:               normalizedNetwork,
		Address:               input.Address,
		IssuedAt:              input.IssuedAt,
		RequiredConfirmations: input.RequiredConfirmations,
		LatestBlockHeight:     input.LatestBlockHeight,
		SinceBlockHeight:      input.SinceBlockHeight,
	})
}

func (o *MultiChainReceiptObserver) FetchLatestBlockHeight(
	ctx context.Context,
	chain value_objects.ChainID,
	network value_objects.NetworkID,
) (int64, error) {
	observer, normalizedNetwork, err := o.resolveObserver(chain, network)
	if err != nil {
		return 0, err
	}
	return observer.FetchLatestBlockHeight(ctx, normalizedNetwork)
}

func (o *MultiChainReceiptObserver) resolveObserver(
	chain value_objects.ChainID,
	network value_objects.NetworkID,
) (outport.ChainReceiptObserver, value_objects.NetworkID, error) {
	normalizedChain, ok := value_objects.ParseChainID(string(chain))
	if !ok {
		return nil, "", errors.New("chain is invalid")
	}
	normalizedNetwork, ok := value_objects.ParseNetworkID(string(network))
	if !ok {
		return nil, "", errors.New("network is invalid")
	}

	observer, found := o.observers[normalizedChain]
	if !found {
		return nil, "", fmt.Errorf("receipt observer is not configured for chain: %s", normalizedChain)
	}
	return observer, normalizedNetwork, nil
}
