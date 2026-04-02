package valueobjects

import "errors"

var (
	ErrAddressPolicyIDInvalid                                = errors.New("address policy id is invalid")
	ErrPaymentReceiptObservationObservedTotalMinorInvalid    = errors.New("observed total minor must be non-negative")
	ErrPaymentReceiptObservationConfirmedTotalMinorInvalid   = errors.New("confirmed total minor must be non-negative")
	ErrPaymentReceiptObservationUnconfirmedTotalMinorInvalid = errors.New("unconfirmed total minor must be non-negative")
	ErrPaymentReceiptObservationLatestBlockHeightInvalid     = errors.New("latest block height must be non-negative")
	ErrPaymentReceiptObservationTotalMismatch                = errors.New("observed total minor must equal confirmed plus unconfirmed total")
	ErrPaymentReceiptObservationConfirmedExceedsObserved     = errors.New("confirmed total minor cannot exceed observed total")
	ErrPaymentReceiptObservationUnconfirmedExceedsObserved   = errors.New("unconfirmed total minor cannot exceed observed total")
)
