package value_objects

import "strings"

type PaymentReceiptStatus string

const (
	PaymentReceiptStatusWatching             PaymentReceiptStatus = "watching"
	PaymentReceiptStatusPartiallyPaid        PaymentReceiptStatus = "partially_paid"
	PaymentReceiptStatusPaidUnconfirmed      PaymentReceiptStatus = "paid_unconfirmed"
	PaymentReceiptStatusPaidConfirmed        PaymentReceiptStatus = "paid_confirmed"
	PaymentReceiptStatusDoubleSpendSuspected PaymentReceiptStatus = "double_spend_suspected"
)

var paymentReceiptStatuses = map[string]PaymentReceiptStatus{
	"watching":               PaymentReceiptStatusWatching,
	"partially_paid":         PaymentReceiptStatusPartiallyPaid,
	"paid_unconfirmed":       PaymentReceiptStatusPaidUnconfirmed,
	"paid_confirmed":         PaymentReceiptStatusPaidConfirmed,
	"double_spend_suspected": PaymentReceiptStatusDoubleSpendSuspected,
}

func ParsePaymentReceiptStatus(raw string) (PaymentReceiptStatus, bool) {
	status, ok := paymentReceiptStatuses[strings.ToLower(strings.TrimSpace(raw))]
	return status, ok
}
