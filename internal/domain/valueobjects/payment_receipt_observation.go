package valueobjects

type PaymentReceiptObservation struct {
	ObservedTotalMinor    int64
	ConfirmedTotalMinor   int64
	UnconfirmedTotalMinor int64
	LatestBlockHeight     int64
}

func (o PaymentReceiptObservation) Validate() error {
	if o.ObservedTotalMinor < 0 {
		return ErrPaymentReceiptObservationObservedTotalMinorInvalid
	}
	if o.ConfirmedTotalMinor < 0 {
		return ErrPaymentReceiptObservationConfirmedTotalMinorInvalid
	}
	if o.UnconfirmedTotalMinor < 0 {
		return ErrPaymentReceiptObservationUnconfirmedTotalMinorInvalid
	}
	if o.LatestBlockHeight < 0 {
		return ErrPaymentReceiptObservationLatestBlockHeightInvalid
	}
	if o.ConfirmedTotalMinor+o.UnconfirmedTotalMinor != o.ObservedTotalMinor {
		return ErrPaymentReceiptObservationTotalMismatch
	}
	if o.ConfirmedTotalMinor > o.ObservedTotalMinor {
		return ErrPaymentReceiptObservationConfirmedExceedsObserved
	}
	if o.UnconfirmedTotalMinor > o.ObservedTotalMinor {
		return ErrPaymentReceiptObservationUnconfirmedExceedsObserved
	}
	return nil
}
