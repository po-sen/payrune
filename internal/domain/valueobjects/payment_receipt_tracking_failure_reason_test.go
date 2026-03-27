package valueobjects

import "testing"

func TestParsePaymentReceiptTrackingFailureReason(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want PaymentReceiptTrackingFailureReason
		ok   bool
	}{
		{
			name: "canonical code",
			raw:  "observation_failed",
			want: PaymentReceiptTrackingFailureReasonObservationFailed,
			ok:   true,
		},
		{
			name: "legacy public text alias",
			raw:  "payment window expired",
			want: PaymentReceiptTrackingFailureReasonPaymentWindowExpired,
			ok:   true,
		},
		{
			name: "legacy raw detail falls back to generic processing failure",
			raw:  "dial tcp timeout",
			want: PaymentReceiptTrackingFailureReasonProcessingFailed,
			ok:   true,
		},
		{
			name: "blank",
			raw:  "   ",
			want: "",
			ok:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParsePaymentReceiptTrackingFailureReason(tc.raw)
			if ok != tc.ok {
				t.Fatalf("unexpected ok: got %v want %v", ok, tc.ok)
			}
			if got != tc.want {
				t.Fatalf("unexpected reason: got %q want %q", got, tc.want)
			}
		})
	}
}

func TestPaymentReceiptTrackingFailureReasonMessage(t *testing.T) {
	reason := PaymentReceiptTrackingFailureReasonLatestBlockHeightUnavailable
	if got := reason.Message(); got != "latest block height unavailable" {
		t.Fatalf("unexpected message: got %q", got)
	}
}
