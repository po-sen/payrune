package valueobjects

import "testing"

func TestParsePaymentReceiptStatus(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   PaymentReceiptStatus
		wantOK bool
	}{
		{
			name:   "watching",
			input:  "watching",
			want:   PaymentReceiptStatusWatching,
			wantOK: true,
		},
		{
			name:   "partially paid",
			input:  "partially_paid",
			want:   PaymentReceiptStatusPartiallyPaid,
			wantOK: true,
		},
		{
			name:   "paid unconfirmed",
			input:  "paid_unconfirmed",
			want:   PaymentReceiptStatusPaidUnconfirmed,
			wantOK: true,
		},
		{
			name:   "paid unconfirmed reverted",
			input:  "paid_unconfirmed_reverted",
			want:   PaymentReceiptStatusPaidUnconfirmedReverted,
			wantOK: true,
		},
		{
			name:   "paid confirmed",
			input:  "paid_confirmed",
			want:   PaymentReceiptStatusPaidConfirmed,
			wantOK: true,
		},
		{
			name:   "failed expired",
			input:  "failed_expired",
			want:   PaymentReceiptStatusFailedExpired,
			wantOK: true,
		},
		{
			name:   "mixed case",
			input:  "  Paid_Confirmed ",
			want:   PaymentReceiptStatusPaidConfirmed,
			wantOK: true,
		},
		{
			name:   "unknown",
			input:  "double_spend_suspected",
			want:   "",
			wantOK: false,
		},
		{
			name:   "another unknown",
			input:  "settled",
			want:   "",
			wantOK: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParsePaymentReceiptStatus(tc.input)
			if ok != tc.wantOK {
				t.Fatalf("unexpected ok: got %v, want %v", ok, tc.wantOK)
			}
			if got != tc.want {
				t.Fatalf("unexpected status: got %q, want %q", got, tc.want)
			}
		})
	}
}
