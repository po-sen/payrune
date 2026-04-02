package policies

import "payrune/internal/domain/valueobjects"

func PollablePaymentReceiptStatuses() []valueobjects.PaymentReceiptStatus {
	return []valueobjects.PaymentReceiptStatus{
		valueobjects.PaymentReceiptStatusWatching,
		valueobjects.PaymentReceiptStatusPartiallyPaid,
		valueobjects.PaymentReceiptStatusPaidUnconfirmed,
		valueobjects.PaymentReceiptStatusPaidUnconfirmedReverted,
	}
}
