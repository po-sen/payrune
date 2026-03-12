package valueobjects

import "errors"

type PaymentReceiptObservation struct {
	ObservedTotalMinor    int64
	ConfirmedTotalMinor   int64
	UnconfirmedTotalMinor int64
	LatestBlockHeight     int64
}

func (o PaymentReceiptObservation) Validate() error {
	if o.ObservedTotalMinor < 0 {
		return errors.New("observed total minor must be non-negative")
	}
	if o.ConfirmedTotalMinor < 0 {
		return errors.New("confirmed total minor must be non-negative")
	}
	if o.UnconfirmedTotalMinor < 0 {
		return errors.New("unconfirmed total minor must be non-negative")
	}
	if o.LatestBlockHeight < 0 {
		return errors.New("latest block height must be non-negative")
	}
	if o.ConfirmedTotalMinor+o.UnconfirmedTotalMinor != o.ObservedTotalMinor {
		return errors.New("observed total minor must equal confirmed plus unconfirmed total")
	}
	if o.ConfirmedTotalMinor > o.ObservedTotalMinor {
		return errors.New("confirmed total minor cannot exceed observed total")
	}
	if o.UnconfirmedTotalMinor > o.ObservedTotalMinor {
		return errors.New("unconfirmed total minor cannot exceed observed total")
	}
	return nil
}
