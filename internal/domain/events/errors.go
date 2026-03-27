package events

import "errors"

var (
	ErrPaymentReceiptStatusChangedPaymentAddressIDInvalid      = errors.New("payment address id must be greater than zero")
	ErrPaymentReceiptStatusChangedPreviousStatusRequired       = errors.New("previous status is required")
	ErrPaymentReceiptStatusChangedCurrentStatusRequired        = errors.New("current status is required")
	ErrPaymentReceiptStatusChangedStatusChangeRequired         = errors.New("status change is required")
	ErrPaymentReceiptStatusChangedObservedTotalMinorInvalid    = errors.New("observed total minor must be greater than or equal to zero")
	ErrPaymentReceiptStatusChangedConfirmedTotalMinorInvalid   = errors.New("confirmed total minor must be greater than or equal to zero")
	ErrPaymentReceiptStatusChangedUnconfirmedTotalMinorInvalid = errors.New("unconfirmed total minor must be greater than or equal to zero")
	ErrPaymentReceiptStatusChangedAtRequired                   = errors.New("status changed at is required")
)
