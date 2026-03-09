package out

import (
	"context"
	"errors"
	"time"

	"payrune/internal/domain/value_objects"
)

var ErrPaymentAddressStatusIncomplete = errors.New("payment address status is incomplete")

type FindPaymentAddressStatusInput struct {
	Chain            value_objects.SupportedChain
	PaymentAddressID int64
}

type PaymentAddressStatusRecord struct {
	PaymentAddressID        int64
	AddressPolicyID         string
	ExpectedAmountMinor     int64
	CustomerReference       string
	Chain                   value_objects.SupportedChain
	Network                 value_objects.NetworkID
	Scheme                  string
	Address                 string
	PaymentStatus           value_objects.PaymentReceiptStatus
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
