package valueobjects

import "testing"

func TestParsePaymentReceiptNotificationDeliveryStatus(t *testing.T) {
	testCases := []struct {
		input  string
		want   PaymentReceiptNotificationDeliveryStatus
		wantOK bool
	}{
		{
			input:  "pending",
			want:   PaymentReceiptNotificationDeliveryStatusPending,
			wantOK: true,
		},
		{
			input:  " SENT ",
			want:   PaymentReceiptNotificationDeliveryStatusSent,
			wantOK: true,
		},
		{
			input:  "failed",
			want:   PaymentReceiptNotificationDeliveryStatusFailed,
			wantOK: true,
		},
		{
			input:  "unknown",
			want:   "",
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got, ok := ParsePaymentReceiptNotificationDeliveryStatus(tc.input)
			if ok != tc.wantOK {
				t.Fatalf("unexpected ok: got %t want %t", ok, tc.wantOK)
			}
			if got != tc.want {
				t.Fatalf("unexpected status: got %q want %q", got, tc.want)
			}
		})
	}
}
