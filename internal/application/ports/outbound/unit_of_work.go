package outbound

import "context"

type TxScope struct {
	PaymentAddressAllocation               PaymentAddressAllocationStore
	PaymentAddressIdempotency              PaymentAddressIdempotencyStore
	PaymentReceiptTracking                 PaymentReceiptTrackingStore
	PaymentReceiptStatusNotificationOutbox PaymentReceiptStatusNotificationOutbox
}

type UnitOfWork interface {
	WithinTransaction(
		ctx context.Context,
		fn func(txScope TxScope) error,
	) error
}
