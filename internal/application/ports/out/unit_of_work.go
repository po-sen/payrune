package out

import "context"

type TxRepositories struct {
	PaymentAddressAllocation PaymentAddressAllocationRepository
	PaymentReceiptTracking   PaymentReceiptTrackingRepository
}

type UnitOfWork interface {
	WithinTransaction(
		ctx context.Context,
		fn func(txRepositories TxRepositories) error,
	) error
}
