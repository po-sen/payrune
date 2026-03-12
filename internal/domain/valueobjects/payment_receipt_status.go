package valueobjects

import "strings"

type PaymentReceiptStatus string

const (
	PaymentReceiptStatusWatching                PaymentReceiptStatus = "watching"
	PaymentReceiptStatusPartiallyPaid           PaymentReceiptStatus = "partially_paid"
	PaymentReceiptStatusPaidUnconfirmed         PaymentReceiptStatus = "paid_unconfirmed"
	PaymentReceiptStatusPaidUnconfirmedReverted PaymentReceiptStatus = "paid_unconfirmed_reverted"
	PaymentReceiptStatusPaidConfirmed           PaymentReceiptStatus = "paid_confirmed"
	PaymentReceiptStatusFailedExpired           PaymentReceiptStatus = "failed_expired"
)

var paymentReceiptStatuses = map[string]PaymentReceiptStatus{
	"watching":                  PaymentReceiptStatusWatching,
	"partially_paid":            PaymentReceiptStatusPartiallyPaid,
	"paid_unconfirmed":          PaymentReceiptStatusPaidUnconfirmed,
	"paid_unconfirmed_reverted": PaymentReceiptStatusPaidUnconfirmedReverted,
	"paid_confirmed":            PaymentReceiptStatusPaidConfirmed,
	"failed_expired":            PaymentReceiptStatusFailedExpired,
}

func ParsePaymentReceiptStatus(raw string) (PaymentReceiptStatus, bool) {
	status, ok := paymentReceiptStatuses[strings.ToLower(strings.TrimSpace(raw))]
	return status, ok
}
