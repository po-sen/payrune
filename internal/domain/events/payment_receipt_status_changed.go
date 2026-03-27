package events

import (
	"time"

	"payrune/internal/domain/valueobjects"
)

type PaymentReceiptStatusChanged struct {
	PaymentAddressID      int64
	PreviousStatus        valueobjects.PaymentReceiptStatus
	CurrentStatus         valueobjects.PaymentReceiptStatus
	ObservedTotalMinor    int64
	ConfirmedTotalMinor   int64
	UnconfirmedTotalMinor int64
	StatusChangedAt       time.Time
}

func NewPaymentReceiptStatusChanged(
	paymentAddressID int64,
	previousStatus valueobjects.PaymentReceiptStatus,
	currentStatus valueobjects.PaymentReceiptStatus,
	observedTotalMinor int64,
	confirmedTotalMinor int64,
	unconfirmedTotalMinor int64,
	statusChangedAt time.Time,
) (PaymentReceiptStatusChanged, error) {
	if paymentAddressID <= 0 {
		return PaymentReceiptStatusChanged{}, ErrPaymentReceiptStatusChangedPaymentAddressIDInvalid
	}
	if previousStatus == "" {
		return PaymentReceiptStatusChanged{}, ErrPaymentReceiptStatusChangedPreviousStatusRequired
	}
	if currentStatus == "" {
		return PaymentReceiptStatusChanged{}, ErrPaymentReceiptStatusChangedCurrentStatusRequired
	}
	if previousStatus == currentStatus {
		return PaymentReceiptStatusChanged{}, ErrPaymentReceiptStatusChangedStatusChangeRequired
	}
	if observedTotalMinor < 0 {
		return PaymentReceiptStatusChanged{}, ErrPaymentReceiptStatusChangedObservedTotalMinorInvalid
	}
	if confirmedTotalMinor < 0 {
		return PaymentReceiptStatusChanged{}, ErrPaymentReceiptStatusChangedConfirmedTotalMinorInvalid
	}
	if unconfirmedTotalMinor < 0 {
		return PaymentReceiptStatusChanged{}, ErrPaymentReceiptStatusChangedUnconfirmedTotalMinorInvalid
	}
	if statusChangedAt.IsZero() {
		return PaymentReceiptStatusChanged{}, ErrPaymentReceiptStatusChangedAtRequired
	}

	return PaymentReceiptStatusChanged{
		PaymentAddressID:      paymentAddressID,
		PreviousStatus:        previousStatus,
		CurrentStatus:         currentStatus,
		ObservedTotalMinor:    observedTotalMinor,
		ConfirmedTotalMinor:   confirmedTotalMinor,
		UnconfirmedTotalMinor: unconfirmedTotalMinor,
		StatusChangedAt:       statusChangedAt.UTC(),
	}, nil
}
