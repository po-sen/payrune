package valueobjects

import "strings"

type PaymentReceiptNotificationDeliveryFailureReason string

const (
	PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed PaymentReceiptNotificationDeliveryFailureReason = "delivery_failed"
)

var paymentReceiptNotificationDeliveryFailureReasons = map[string]PaymentReceiptNotificationDeliveryFailureReason{
	"delivery_failed": PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed,
}

var paymentReceiptNotificationDeliveryFailureReasonLegacyAliases = map[string]PaymentReceiptNotificationDeliveryFailureReason{
	"receipt webhook delivery failed": PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed,
	"timeout":                         PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed,
	"webhook returned status 500":     PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed,
}

func ParsePaymentReceiptNotificationDeliveryFailureReason(raw string) (PaymentReceiptNotificationDeliveryFailureReason, bool) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return "", false
	}
	if reason, ok := paymentReceiptNotificationDeliveryFailureReasons[normalized]; ok {
		return reason, true
	}
	if reason, ok := paymentReceiptNotificationDeliveryFailureReasonLegacyAliases[normalized]; ok {
		return reason, true
	}
	return PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed, true
}

func (r PaymentReceiptNotificationDeliveryFailureReason) IsZero() bool {
	return strings.TrimSpace(string(r)) == ""
}

func (r PaymentReceiptNotificationDeliveryFailureReason) Message() string {
	switch r {
	case PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed:
		return "receipt webhook delivery failed"
	default:
		return ""
	}
}
