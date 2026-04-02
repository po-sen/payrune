package valueobjects

import "testing"

func TestParsePaymentAddressAllocationDerivationFailureReason(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want PaymentAddressAllocationDerivationFailureReason
		ok   bool
	}{
		{
			name: "canonical code",
			raw:  "derivation_failed",
			want: PaymentAddressAllocationDerivationFailureReasonDerivationFailed,
			ok:   true,
		},
		{
			name: "canonical code with whitespace",
			raw:  " derivation_failed ",
			want: PaymentAddressAllocationDerivationFailureReasonDerivationFailed,
			ok:   true,
		},
		{
			name: "legacy alias rejected",
			raw:  " derive failed ",
			want: "",
			ok:   false,
		},
		{
			name: "unknown raw detail rejected",
			raw:  "xpub parse exploded",
			want: "",
			ok:   false,
		},
		{
			name: "empty",
			raw:  "   ",
			want: "",
			ok:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := ParsePaymentAddressAllocationDerivationFailureReason(tc.raw)
			if ok != tc.ok {
				t.Fatalf("unexpected ok: got %v want %v", ok, tc.ok)
			}
			if got != tc.want {
				t.Fatalf("unexpected reason: got %q want %q", got, tc.want)
			}
		})
	}
}

func TestPaymentAddressAllocationDerivationFailureReasonMessage(t *testing.T) {
	t.Parallel()

	reason := PaymentAddressAllocationDerivationFailureReasonDerivationFailed
	if got := reason.Message(); got != "payment address derivation failed" {
		t.Fatalf("unexpected message: got %q", got)
	}
}
