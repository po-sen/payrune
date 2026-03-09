package out

import (
	"context"
	"time"

	"payrune/internal/domain/value_objects"
)

type ObservePaymentAddressInput struct {
	Network               value_objects.NetworkID
	Address               string
	IssuedAt              time.Time
	RequiredConfirmations int32
	LatestBlockHeight     int64
	SinceBlockHeight      int64
}

type ObserveChainPaymentAddressInput struct {
	Chain                 value_objects.ChainID
	Network               value_objects.NetworkID
	Address               string
	IssuedAt              time.Time
	RequiredConfirmations int32
	LatestBlockHeight     int64
	SinceBlockHeight      int64
}

type ObservePaymentAddressOutput struct {
	ObservedTotalMinor    int64
	ConfirmedTotalMinor   int64
	UnconfirmedTotalMinor int64
	LatestBlockHeight     int64
}

type ChainReceiptObserver interface {
	FetchLatestBlockHeight(ctx context.Context, network value_objects.NetworkID) (int64, error)
	ObserveAddress(ctx context.Context, input ObservePaymentAddressInput) (ObservePaymentAddressOutput, error)
}

type BlockchainReceiptObserver interface {
	FetchLatestBlockHeight(ctx context.Context, chain value_objects.ChainID, network value_objects.NetworkID) (int64, error)
	ObserveAddress(ctx context.Context, input ObserveChainPaymentAddressInput) (ObservePaymentAddressOutput, error)
}
