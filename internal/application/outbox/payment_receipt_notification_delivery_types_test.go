package outbox

import "testing"

func TestParsePaymentReceiptNotificationDeliveryStatus(t *testing.T) {
	tests := []struct {
		input  string
		want   PaymentReceiptNotificationDeliveryStatus
		wantOK bool
	}{
		{input: " pending ", want: PaymentReceiptNotificationDeliveryStatusPending, wantOK: true},
		{input: "SENT", want: PaymentReceiptNotificationDeliveryStatusSent, wantOK: true},
		{input: "failed", want: PaymentReceiptNotificationDeliveryStatusFailed, wantOK: true},
		{input: "unknown", wantOK: false},
	}

	for _, tc := range tests {
		got, ok := ParsePaymentReceiptNotificationDeliveryStatus(tc.input)
		if ok != tc.wantOK {
			t.Fatalf("unexpected ok for %q: got %v want %v", tc.input, ok, tc.wantOK)
		}
		if got != tc.want {
			t.Fatalf("unexpected status for %q: got %q want %q", tc.input, got, tc.want)
		}
	}
}

func TestParsePaymentReceiptNotificationDeliveryFailureReason(t *testing.T) {
	tests := []struct {
		raw    string
		want   PaymentReceiptNotificationDeliveryFailureReason
		wantOK bool
	}{
		{raw: "delivery_failed", want: PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed, wantOK: true},
		{raw: " DELIVERY_FAILED ", want: PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed, wantOK: true},
		{raw: "", wantOK: false},
		{raw: "unknown", wantOK: false},
	}

	for _, tc := range tests {
		got, ok := ParsePaymentReceiptNotificationDeliveryFailureReason(tc.raw)
		if ok != tc.wantOK {
			t.Fatalf("unexpected ok for %q: got %v want %v", tc.raw, ok, tc.wantOK)
		}
		if got != tc.want {
			t.Fatalf("unexpected reason for %q: got %q want %q", tc.raw, got, tc.want)
		}
	}
}

func TestPaymentReceiptNotificationDeliveryFailureReasonMessage(t *testing.T) {
	reason := PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed
	if reason.Message() != "receipt webhook delivery failed" {
		t.Fatalf("unexpected message: got %q", reason.Message())
	}
}
