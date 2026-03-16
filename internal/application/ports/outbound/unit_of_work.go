package outbound

import "context"

type TxScope struct {
	EVMFactoryRegistry                     EVMFactoryStore
	EVMPaymentVaults                       EVMPaymentVaultStore
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
