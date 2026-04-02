package bitcoin

import (
	"context"
	"strings"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

type CloudflareBitcoinEsploraReceiptObserver struct {
	bridgeID string
	bridge   cloudflareEsploraBridge
}

func NewCloudflareBitcoinEsploraReceiptObserver(
	bridgeID string,
	bridge cloudflareEsploraBridge,
) *CloudflareBitcoinEsploraReceiptObserver {
	return &CloudflareBitcoinEsploraReceiptObserver{
		bridgeID: strings.TrimSpace(bridgeID),
		bridge:   bridge,
	}
}

func (o *CloudflareBitcoinEsploraReceiptObserver) ObserveAddress(
	ctx context.Context,
	input outport.ObservePaymentAddressInput,
) (outport.ObservePaymentAddressOutput, error) {
	address := strings.TrimSpace(input.Address)
	if address == "" {
		return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverInputInvalid
	}
	if input.IssuedAt.IsZero() {
		return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverInputInvalid
	}
	if input.RequiredConfirmations <= 0 {
		return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverInputInvalid
	}
	if input.LatestBlockHeight <= 0 {
		return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverInputInvalid
	}
	if input.SinceBlockHeight < 0 {
		return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverInputInvalid
	}

	if _, err := o.validateNetwork(input.Network); err != nil {
		return outport.ObservePaymentAddressOutput{}, err
	}
	if o.bridge == nil {
		return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverNotConfigured
	}

	chainTransactions, err := o.bridge.FetchAddressChainTransactions(ctx, o.bridgeID, input.Network, address)
	if err != nil {
		return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverFailed
	}
	mempoolTransactions, err := o.bridge.FetchAddressMempoolTransactions(ctx, o.bridgeID, input.Network, address)
	if err != nil {
		return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverFailed
	}

	confirmedTotalMinor, unconfirmedTotalMinor, err := aggregateInboundTotals(
		address,
		input.IssuedAt.UTC(),
		int64(input.RequiredConfirmations),
		input.LatestBlockHeight,
		chainTransactions,
		mempoolTransactions,
	)
	if err != nil {
		return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverFailed
	}

	return outport.ObservePaymentAddressOutput{
		ObservedTotalMinor:    confirmedTotalMinor + unconfirmedTotalMinor,
		ConfirmedTotalMinor:   confirmedTotalMinor,
		UnconfirmedTotalMinor: unconfirmedTotalMinor,
		LatestBlockHeight:     input.LatestBlockHeight,
	}, nil
}

func (o *CloudflareBitcoinEsploraReceiptObserver) FetchLatestBlockHeight(
	ctx context.Context,
	network valueobjects.NetworkID,
) (int64, error) {
	if _, err := o.validateNetwork(network); err != nil {
		return 0, err
	}
	if o.bridge == nil {
		return 0, outport.ErrBlockchainReceiptObserverNotConfigured
	}
	latestBlockHeight, err := o.bridge.FetchLatestBlockHeight(ctx, o.bridgeID, network)
	if err != nil {
		return 0, outport.ErrBlockchainReceiptObserverFailed
	}
	return latestBlockHeight, nil
}

func (o *CloudflareBitcoinEsploraReceiptObserver) validateNetwork(
	network valueobjects.NetworkID,
) (network, error) {
	if strings.TrimSpace(o.bridgeID) == "" {
		return "", outport.ErrBlockchainReceiptObserverInputInvalid
	}

	bitcoinNetwork, ok := parseNetwork(network)
	if !ok {
		return "", outport.ErrBlockchainReceiptObserverInputInvalid
	}
	return bitcoinNetwork, nil
}

var _ outport.ChainReceiptObserver = (*CloudflareBitcoinEsploraReceiptObserver)(nil)
