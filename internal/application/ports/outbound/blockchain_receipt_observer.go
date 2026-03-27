package outbound

import (
	"context"
	"errors"
	"time"

	"payrune/internal/domain/valueobjects"
)

var (
	ErrBlockchainReceiptObserverNotConfigured = errors.New("blockchain receipt observer is not configured")
	ErrBlockchainReceiptObserverInputInvalid  = errors.New("blockchain receipt observer input is invalid")
	ErrBlockchainReceiptObserverFailed        = errors.New("blockchain receipt observer failed")
)

type ObservePaymentAddressInput struct {
	Network               valueobjects.NetworkID
	Address               string
	IssuedAt              time.Time
	RequiredConfirmations int32
	LatestBlockHeight     int64
	SinceBlockHeight      int64
}

type ObserveChainPaymentAddressInput struct {
	Chain                 valueobjects.ChainID
	Network               valueobjects.NetworkID
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
	FetchLatestBlockHeight(ctx context.Context, network valueobjects.NetworkID) (int64, error)
	ObserveAddress(ctx context.Context, input ObservePaymentAddressInput) (ObservePaymentAddressOutput, error)
}

type BlockchainReceiptObserver interface {
	FetchLatestBlockHeight(ctx context.Context, chain valueobjects.ChainID, network valueobjects.NetworkID) (int64, error)
	ObserveAddress(ctx context.Context, input ObserveChainPaymentAddressInput) (ObservePaymentAddressOutput, error)
}
