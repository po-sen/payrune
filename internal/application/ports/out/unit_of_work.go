package out

import "context"

type TxScope struct {
	PaymentAddressAllocation               PaymentAddressAllocationStore
	PaymentReceiptTracking                 PaymentReceiptTrackingStore
	PaymentReceiptStatusNotificationOutbox PaymentReceiptStatusNotificationOutbox
}

type UnitOfWork interface {
	WithinTransaction(
		ctx context.Context,
		fn func(txScope TxScope) error,
	) error
}
