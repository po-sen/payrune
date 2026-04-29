package cloudflarepostgres

import (
	"testing"

	outport "payrune/internal/application/ports/outbound"
)

func TestNormalizePaymentReceiptTrackingFailureReason(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "canonical code",
			raw:  "observation_failed",
			want: outport.PaymentReceiptTrackingFailureReasonObservationFailed,
		},
		{
			name: "legacy alias",
			raw:  "payment window expired",
			want: outport.PaymentReceiptTrackingFailureReasonPaymentWindowExpired,
		},
		{
			name: "unknown raw detail falls back",
			raw:  "dial tcp timeout",
			want: outport.PaymentReceiptTrackingFailureReasonProcessingFailed,
		},
		{
			name: "blank remains zero",
			raw:  " ",
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizePaymentReceiptTrackingFailureReason(tc.raw); got != tc.want {
				t.Fatalf("unexpected reason: got %q want %q", got, tc.want)
			}
		})
	}
}

func TestNormalizePaymentReceiptNotificationDeliveryFailureReason(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "canonical code",
			raw:  "delivery_failed",
			want: outport.PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed,
		},
		{
			name: "legacy alias",
			raw:  "receipt webhook delivery failed",
			want: outport.PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed,
		},
		{
			name: "unknown raw detail falls back",
			raw:  "webhook returned status 429",
			want: outport.PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed,
		},
		{
			name: "blank remains zero",
			raw:  "",
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizePaymentReceiptNotificationDeliveryFailureReason(tc.raw); got != tc.want {
				t.Fatalf("unexpected reason: got %q want %q", got, tc.want)
			}
		})
	}
}

func TestNormalizePaymentAddressAllocationDerivationFailureReason(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "canonical code",
			raw:  "derivation_failed",
			want: outport.PaymentAddressAllocationFailureDerivationFailed,
		},
		{
			name: "legacy alias",
			raw:  "derive failed",
			want: outport.PaymentAddressAllocationFailureDerivationFailed,
		},
		{
			name: "unknown raw detail falls back",
			raw:  "xpub parse exploded",
			want: outport.PaymentAddressAllocationFailureDerivationFailed,
		},
		{
			name: "blank remains zero",
			raw:  " ",
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizePaymentAddressAllocationDerivationFailureReason(tc.raw); got != tc.want {
				t.Fatalf("unexpected reason: got %q want %q", got, tc.want)
			}
		})
	}
}
