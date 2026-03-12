package bitcoin

import (
	"context"
	"errors"
	"fmt"
	"strings"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

type CloudflareBitcoinEsploraReceiptObserver struct {
	bridgeID string
	bridge   CloudflareEsploraBridge
}

func NewCloudflareBitcoinEsploraReceiptObserver(
	bridgeID string,
	bridge CloudflareEsploraBridge,
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
		return outport.ObservePaymentAddressOutput{}, errors.New("address is required")
	}
	if input.IssuedAt.IsZero() {
		return outport.ObservePaymentAddressOutput{}, errors.New("issued at is required")
	}
	if input.RequiredConfirmations <= 0 {
		return outport.ObservePaymentAddressOutput{}, errors.New("required confirmations must be greater than zero")
	}
	if input.LatestBlockHeight <= 0 {
		return outport.ObservePaymentAddressOutput{}, errors.New("latest block height must be greater than zero")
	}
	if input.SinceBlockHeight < 0 {
		return outport.ObservePaymentAddressOutput{}, errors.New("since block height must be non-negative")
	}

	if _, err := o.validateNetwork(input.Network); err != nil {
		return outport.ObservePaymentAddressOutput{}, err
	}
	if o.bridge == nil {
		return outport.ObservePaymentAddressOutput{}, errors.New("cloudflare bitcoin esplora bridge is not configured")
	}

	chainTransactions, err := o.bridge.FetchAddressChainTransactions(ctx, o.bridgeID, input.Network, address)
	if err != nil {
		return outport.ObservePaymentAddressOutput{}, err
	}
	mempoolTransactions, err := o.bridge.FetchAddressMempoolTransactions(ctx, o.bridgeID, input.Network, address)
	if err != nil {
		return outport.ObservePaymentAddressOutput{}, err
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
		return outport.ObservePaymentAddressOutput{}, err
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
		return 0, errors.New("cloudflare bitcoin esplora bridge is not configured")
	}
	return o.bridge.FetchLatestBlockHeight(ctx, o.bridgeID, network)
}

func (o *CloudflareBitcoinEsploraReceiptObserver) validateNetwork(
	network valueobjects.NetworkID,
) (valueobjects.BitcoinNetwork, error) {
	if strings.TrimSpace(o.bridgeID) == "" {
		return "", errors.New("cloudflare bitcoin esplora bridge id is required")
	}

	bitcoinNetwork, ok := valueobjects.ParseBitcoinNetwork(string(network))
	if !ok {
		return "", fmt.Errorf("bitcoin network is not supported: %s", network)
	}
	return bitcoinNetwork, nil
}

var _ outport.ChainReceiptObserver = (*CloudflareBitcoinEsploraReceiptObserver)(nil)
