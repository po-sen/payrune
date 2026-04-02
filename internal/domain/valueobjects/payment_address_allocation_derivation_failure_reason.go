package valueobjects

import "strings"

type PaymentAddressAllocationDerivationFailureReason string

const (
	PaymentAddressAllocationDerivationFailureReasonDerivationFailed PaymentAddressAllocationDerivationFailureReason = "derivation_failed"
)

var paymentAddressAllocationDerivationFailureReasons = map[string]PaymentAddressAllocationDerivationFailureReason{
	"derivation_failed": PaymentAddressAllocationDerivationFailureReasonDerivationFailed,
}

func ParsePaymentAddressAllocationDerivationFailureReason(raw string) (PaymentAddressAllocationDerivationFailureReason, bool) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return "", false
	}
	if reason, ok := paymentAddressAllocationDerivationFailureReasons[normalized]; ok {
		return reason, true
	}
	return "", false
}

func (r PaymentAddressAllocationDerivationFailureReason) IsZero() bool {
	return strings.TrimSpace(string(r)) == ""
}

func (r PaymentAddressAllocationDerivationFailureReason) Message() string {
	switch r {
	case PaymentAddressAllocationDerivationFailureReasonDerivationFailed:
		return "payment address derivation failed"
	default:
		return ""
	}
}
