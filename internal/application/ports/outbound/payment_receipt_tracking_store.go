package outbound

import (
	"context"
	"errors"
	"time"
)

var (
	ErrPaymentReceiptTrackingStoreFailed                     = errors.New("payment receipt tracking store failed")
	ErrPaymentReceiptTrackingNextPollAtRequired              = errors.New("next poll at is required")
	ErrPaymentReceiptTrackingAlreadyExists                   = errors.New("payment receipt tracking already exists")
	ErrPaymentReceiptTrackingClaimNowRequired                = errors.New("claim now is required")
	ErrPaymentReceiptTrackingClaimUntilRequired              = errors.New("claim until is required")
	ErrPaymentReceiptTrackingClaimLimitInvalid               = errors.New("claim limit must be greater than zero")
	ErrPaymentReceiptTrackingClaimStatusesRequired           = errors.New("claim statuses are required")
	ErrPaymentReceiptTrackingClaimStatusRequired             = errors.New("claim status is required")
	ErrPaymentReceiptTrackingPolledAtRequired                = errors.New("polled at is required")
	ErrPaymentReceiptTrackingNotFound                        = errors.New("payment receipt tracking is not found")
	ErrPaymentReceiptTrackingPersistedAddressPolicyIDInvalid = errors.New("persisted receipt tracking address policy id is invalid")
	ErrPaymentReceiptTrackingPersistedChainInvalid           = errors.New("persisted receipt tracking chain is invalid")
	ErrPaymentReceiptTrackingPersistedNetworkInvalid         = errors.New("persisted receipt tracking network is invalid")
	ErrPaymentReceiptTrackingPersistedAssetReferenceInvalid  = errors.New("persisted receipt tracking asset reference is invalid")
	ErrPaymentReceiptTrackingPersistedStatusInvalid          = errors.New("persisted receipt tracking status is invalid")
)

type PaymentReceiptTrackingRecord struct {
	TrackingID              int64
	PaymentAddressID        int64
	AddressPolicyID         string
	Chain                   string
	Network                 string
	Address                 string
	AssetReference          string
	IssuedAt                time.Time
	ExpectedAmountMinor     int64
	RequiredConfirmations   int32
	Status                  string
	ObservedTotalMinor      int64
	ConfirmedTotalMinor     int64
	UnconfirmedTotalMinor   int64
	LastObservedBlockHeight int64
	FirstObservedAt         *time.Time
	PaidAt                  *time.Time
	ConfirmedAt             *time.Time
	ExpiresAt               *time.Time
	LastFailureReason       string
}

type ClaimPaymentReceiptTrackingsInput struct {
	Now        time.Time
	Limit      int
	ClaimUntil time.Time
	Chain      string
	Network    string
	Statuses   []string
}

type PaymentReceiptTrackingStore interface {
	Create(
		ctx context.Context,
		tracking PaymentReceiptTrackingRecord,
		nextPollAt time.Time,
	) error
	ClaimDue(
		ctx context.Context,
		input ClaimPaymentReceiptTrackingsInput,
	) ([]PaymentReceiptTrackingRecord, error)
	Save(
		ctx context.Context,
		tracking PaymentReceiptTrackingRecord,
		polledAt time.Time,
		nextPollAt time.Time,
	) error
}
