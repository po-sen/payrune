package value_objects

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
