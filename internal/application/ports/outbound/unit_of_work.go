package outbound

import (
	"context"
	"errors"
)

var (
	ErrUnitOfWorkNotConfigured = errors.New("unit of work is not configured")
	ErrUnitOfWorkFailed        = errors.New("unit of work failed")
)

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
