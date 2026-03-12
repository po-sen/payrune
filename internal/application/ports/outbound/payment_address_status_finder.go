package outbound

import (
	"context"
	"errors"
	"time"

	"payrune/internal/domain/valueobjects"
)

var ErrPaymentAddressStatusIncomplete = errors.New("payment address status is incomplete")

type FindPaymentAddressStatusInput struct {
	Chain            valueobjects.SupportedChain
	PaymentAddressID int64
}

type PaymentAddressStatusRecord struct {
	PaymentAddressID        int64
	AddressPolicyID         string
	ExpectedAmountMinor     int64
	CustomerReference       string
	Chain                   valueobjects.SupportedChain
	Network                 valueobjects.NetworkID
	Scheme                  string
	Address                 string
	PaymentStatus           valueobjects.PaymentReceiptStatus
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
	LastError               string
}

type PaymentAddressStatusFinder interface {
	FindByID(
		ctx context.Context,
		input FindPaymentAddressStatusInput,
	) (PaymentAddressStatusRecord, bool, error)
}
