package value_objects

import "testing"

func TestPaymentReceiptObservationValidate(t *testing.T) {
	tests := []struct {
		name        string
		observation PaymentReceiptObservation
		wantErr     bool
	}{
		{
			name: "valid",
			observation: PaymentReceiptObservation{
				ObservedTotalMinor:    100,
				ConfirmedTotalMinor:   60,
				UnconfirmedTotalMinor: 40,
				ConflictTotalMinor:    0,
				LatestBlockHeight:     10,
			},
			wantErr: false,
		},
		{
			name: "negative observed",
			observation: PaymentReceiptObservation{
				ObservedTotalMinor:    -1,
				ConfirmedTotalMinor:   0,
				UnconfirmedTotalMinor: 0,
				ConflictTotalMinor:    0,
				LatestBlockHeight:     0,
			},
			wantErr: true,
		},
		{
			name: "sum mismatch",
			observation: PaymentReceiptObservation{
				ObservedTotalMinor:    100,
				ConfirmedTotalMinor:   80,
				UnconfirmedTotalMinor: 10,
				ConflictTotalMinor:    0,
				LatestBlockHeight:     0,
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.observation.Validate()
			if tc.wantErr && err == nil {
				t.Fatal("expected error but got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}
