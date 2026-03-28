package valueobjects

import "strings"

type PaymentAddressAllocationDerivationFailureReason string

const (
	PaymentAddressAllocationDerivationFailureReasonDerivationFailed PaymentAddressAllocationDerivationFailureReason = "derivation_failed"
)

var paymentAddressAllocationDerivationFailureReasons = map[string]PaymentAddressAllocationDerivationFailureReason{
	"derivation_failed": PaymentAddressAllocationDerivationFailureReasonDerivationFailed,
}

var paymentAddressAllocationDerivationFailureReasonLegacyAliases = map[string]PaymentAddressAllocationDerivationFailureReason{
	"derive failed":                     PaymentAddressAllocationDerivationFailureReasonDerivationFailed,
	"derivation failed":                 PaymentAddressAllocationDerivationFailureReasonDerivationFailed,
	"address derivation failed":         PaymentAddressAllocationDerivationFailureReasonDerivationFailed,
	"payment address derivation failed": PaymentAddressAllocationDerivationFailureReasonDerivationFailed,
}

func ParsePaymentAddressAllocationDerivationFailureReason(raw string) (PaymentAddressAllocationDerivationFailureReason, bool) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return "", false
	}
	if reason, ok := paymentAddressAllocationDerivationFailureReasons[normalized]; ok {
		return reason, true
	}
	if reason, ok := paymentAddressAllocationDerivationFailureReasonLegacyAliases[normalized]; ok {
		return reason, true
	}
	return PaymentAddressAllocationDerivationFailureReasonDerivationFailed, true
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
