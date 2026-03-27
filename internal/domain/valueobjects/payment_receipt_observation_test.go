package valueobjects

import (
	"errors"
	"testing"
)

func TestPaymentReceiptObservationValidate(t *testing.T) {
	tests := []struct {
		name        string
		observation PaymentReceiptObservation
		wantErr     error
	}{
		{
			name: "valid",
			observation: PaymentReceiptObservation{
				ObservedTotalMinor:    100,
				ConfirmedTotalMinor:   60,
				UnconfirmedTotalMinor: 40,
				LatestBlockHeight:     10,
			},
			wantErr: nil,
		},
		{
			name: "negative observed",
			observation: PaymentReceiptObservation{
				ObservedTotalMinor:    -1,
				ConfirmedTotalMinor:   0,
				UnconfirmedTotalMinor: 0,
				LatestBlockHeight:     0,
			},
			wantErr: ErrPaymentReceiptObservationObservedTotalMinorInvalid,
		},
		{
			name: "sum mismatch",
			observation: PaymentReceiptObservation{
				ObservedTotalMinor:    100,
				ConfirmedTotalMinor:   80,
				UnconfirmedTotalMinor: 10,
				LatestBlockHeight:     0,
			},
			wantErr: ErrPaymentReceiptObservationTotalMismatch,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.observation.Validate()
			if tc.wantErr == nil && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}
