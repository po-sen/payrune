package outbox

import "strings"

type PaymentReceiptNotificationDeliveryStatus string

const (
	PaymentReceiptNotificationDeliveryStatusPending PaymentReceiptNotificationDeliveryStatus = "pending"
	PaymentReceiptNotificationDeliveryStatusSent    PaymentReceiptNotificationDeliveryStatus = "sent"
	PaymentReceiptNotificationDeliveryStatusFailed  PaymentReceiptNotificationDeliveryStatus = "failed"
)

var paymentReceiptNotificationDeliveryStatuses = map[string]PaymentReceiptNotificationDeliveryStatus{
	"pending": PaymentReceiptNotificationDeliveryStatusPending,
	"sent":    PaymentReceiptNotificationDeliveryStatusSent,
	"failed":  PaymentReceiptNotificationDeliveryStatusFailed,
}

func ParsePaymentReceiptNotificationDeliveryStatus(raw string) (PaymentReceiptNotificationDeliveryStatus, bool) {
	status, ok := paymentReceiptNotificationDeliveryStatuses[strings.ToLower(strings.TrimSpace(raw))]
	return status, ok
}

type PaymentReceiptNotificationDeliveryFailureReason string

const (
	PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed PaymentReceiptNotificationDeliveryFailureReason = "delivery_failed"
)

var paymentReceiptNotificationDeliveryFailureReasons = map[string]PaymentReceiptNotificationDeliveryFailureReason{
	"delivery_failed": PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed,
}

func ParsePaymentReceiptNotificationDeliveryFailureReason(raw string) (PaymentReceiptNotificationDeliveryFailureReason, bool) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return "", false
	}
	if reason, ok := paymentReceiptNotificationDeliveryFailureReasons[normalized]; ok {
		return reason, true
	}
	return "", false
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
