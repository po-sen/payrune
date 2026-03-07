package events

import (
	"errors"
	"time"

	"payrune/internal/domain/value_objects"
)

type PaymentReceiptStatusChanged struct {
	PaymentAddressID      int64
	PreviousStatus        value_objects.PaymentReceiptStatus
	CurrentStatus         value_objects.PaymentReceiptStatus
	ObservedTotalMinor    int64
	ConfirmedTotalMinor   int64
	UnconfirmedTotalMinor int64
	ConflictTotalMinor    int64
	StatusChangedAt       time.Time
}

func NewPaymentReceiptStatusChanged(
	paymentAddressID int64,
	previousStatus value_objects.PaymentReceiptStatus,
	currentStatus value_objects.PaymentReceiptStatus,
	observedTotalMinor int64,
	confirmedTotalMinor int64,
	unconfirmedTotalMinor int64,
	conflictTotalMinor int64,
	statusChangedAt time.Time,
) (PaymentReceiptStatusChanged, error) {
	if paymentAddressID <= 0 {
		return PaymentReceiptStatusChanged{}, errors.New("payment address id must be greater than zero")
	}
	if previousStatus == "" {
		return PaymentReceiptStatusChanged{}, errors.New("previous status is required")
	}
	if currentStatus == "" {
		return PaymentReceiptStatusChanged{}, errors.New("current status is required")
	}
	if previousStatus == currentStatus {
		return PaymentReceiptStatusChanged{}, errors.New("status change is required")
	}
	if observedTotalMinor < 0 {
		return PaymentReceiptStatusChanged{}, errors.New("observed total minor must be greater than or equal to zero")
	}
	if confirmedTotalMinor < 0 {
		return PaymentReceiptStatusChanged{}, errors.New("confirmed total minor must be greater than or equal to zero")
	}
	if unconfirmedTotalMinor < 0 {
		return PaymentReceiptStatusChanged{}, errors.New("unconfirmed total minor must be greater than or equal to zero")
	}
	if conflictTotalMinor < 0 {
		return PaymentReceiptStatusChanged{}, errors.New("conflict total minor must be greater than or equal to zero")
	}
	if statusChangedAt.IsZero() {
		return PaymentReceiptStatusChanged{}, errors.New("status changed at is required")
	}

	return PaymentReceiptStatusChanged{
		PaymentAddressID:      paymentAddressID,
		PreviousStatus:        previousStatus,
		CurrentStatus:         currentStatus,
		ObservedTotalMinor:    observedTotalMinor,
		ConfirmedTotalMinor:   confirmedTotalMinor,
		UnconfirmedTotalMinor: unconfirmedTotalMinor,
		ConflictTotalMinor:    conflictTotalMinor,
		StatusChangedAt:       statusChangedAt.UTC(),
	}, nil
}
