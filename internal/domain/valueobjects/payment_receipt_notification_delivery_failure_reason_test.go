package valueobjects

import "testing"

func TestParsePaymentReceiptNotificationDeliveryFailureReason(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want PaymentReceiptNotificationDeliveryFailureReason
		ok   bool
	}{
		{
			name: "canonical code",
			raw:  "delivery_failed",
			want: PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed,
			ok:   true,
		},
		{
			name: "legacy public text alias",
			raw:  "receipt webhook delivery failed",
			want: PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed,
			ok:   true,
		},
		{
			name: "legacy raw detail falls back to generic delivery failure",
			raw:  "webhook returned status 429",
			want: PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed,
			ok:   true,
		},
		{
			name: "blank",
			raw:  " ",
			want: "",
			ok:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParsePaymentReceiptNotificationDeliveryFailureReason(tc.raw)
			if ok != tc.ok {
				t.Fatalf("unexpected ok: got %v want %v", ok, tc.ok)
			}
			if got != tc.want {
				t.Fatalf("unexpected reason: got %q want %q", got, tc.want)
			}
		})
	}
}

func TestPaymentReceiptNotificationDeliveryFailureReasonMessage(t *testing.T) {
	reason := PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed
	if got := reason.Message(); got != "receipt webhook delivery failed" {
		t.Fatalf("unexpected message: got %q", got)
	}
}
