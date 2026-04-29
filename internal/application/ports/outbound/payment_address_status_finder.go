package outbound

import (
	"context"
	"errors"
	"time"
)

var (
	ErrPaymentAddressStatusFindFailed                      = errors.New("payment address status finder failed")
	ErrPaymentAddressStatusIncomplete                      = errors.New("payment address status is incomplete")
	ErrPaymentAddressStatusPersistedAddressPolicyIDInvalid = errors.New("persisted payment address policy id is invalid")
	ErrPaymentAddressStatusPersistedChainInvalid           = errors.New("persisted payment address chain is invalid")
	ErrPaymentAddressStatusPersistedNetworkInvalid         = errors.New("persisted payment address network is invalid")
	ErrPaymentAddressStatusPersistedReceiptStatusInvalid   = errors.New("persisted payment receipt status is invalid")
)

type FindPaymentAddressStatusInput struct {
	Chain            string
	PaymentAddressID int64
}

type PaymentAddressStatusRecord struct {
	PaymentAddressID        int64
	AddressPolicyID         string
	ExpectedAmountMinor     int64
	CustomerReference       string
	Chain                   string
	Network                 string
	Scheme                  string
	Address                 string
	PaymentStatus           string
	ObservedTotalMinor      int64
	ConfirmedTotalMinor     int64
	UnconfirmedTotalMinor   int64
	RequiredConfirmations   int32
	LastObservedBlockHeight int64
	IssuedAt                time.Time
	FirstObservedAt         *time.Time
	PaidAt                  *time.Time
	ConfirmedAt             *time.Time
	ExpiresAt               *time.Time
	LastFailureReason       string
}

type PaymentAddressStatusFinder interface {
	FindByID(
		ctx context.Context,
		input FindPaymentAddressStatusInput,
	) (PaymentAddressStatusRecord, bool, error)
}
