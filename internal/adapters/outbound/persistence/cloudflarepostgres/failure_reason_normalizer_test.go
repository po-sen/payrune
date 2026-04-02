package cloudflarepostgres

import (
	"testing"

	"payrune/internal/domain/valueobjects"
)

func TestNormalizePaymentReceiptTrackingFailureReason(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want valueobjects.PaymentReceiptTrackingFailureReason
	}{
		{
			name: "canonical code",
			raw:  "observation_failed",
			want: valueobjects.PaymentReceiptTrackingFailureReasonObservationFailed,
		},
		{
			name: "legacy alias",
			raw:  "payment window expired",
			want: valueobjects.PaymentReceiptTrackingFailureReasonPaymentWindowExpired,
		},
		{
			name: "unknown raw detail falls back",
			raw:  "dial tcp timeout",
			want: valueobjects.PaymentReceiptTrackingFailureReasonProcessingFailed,
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
		want valueobjects.PaymentReceiptNotificationDeliveryFailureReason
	}{
		{
			name: "canonical code",
			raw:  "delivery_failed",
			want: valueobjects.PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed,
		},
		{
			name: "legacy alias",
			raw:  "receipt webhook delivery failed",
			want: valueobjects.PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed,
		},
		{
			name: "unknown raw detail falls back",
			raw:  "webhook returned status 429",
			want: valueobjects.PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed,
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
		want valueobjects.PaymentAddressAllocationDerivationFailureReason
	}{
		{
			name: "canonical code",
			raw:  "derivation_failed",
			want: valueobjects.PaymentAddressAllocationDerivationFailureReasonDerivationFailed,
		},
		{
			name: "legacy alias",
			raw:  "derive failed",
			want: valueobjects.PaymentAddressAllocationDerivationFailureReasonDerivationFailed,
		},
		{
			name: "unknown raw detail falls back",
			raw:  "xpub parse exploded",
			want: valueobjects.PaymentAddressAllocationDerivationFailureReasonDerivationFailed,
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
