package postgres

import (
	"context"
	"database/sql"
	"errors"

	outport "payrune/internal/application/ports/out"
)

type UnitOfWork struct {
	db *sql.DB
}

func NewUnitOfWork(db *sql.DB) *UnitOfWork {
	return &UnitOfWork{
		db: db,
	}
}

func (u *UnitOfWork) WithinTransaction(
	ctx context.Context,
	fn func(txScope outport.TxScope) error,
) error {
	if u.db == nil {
		return errors.New("database is not configured")
	}

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	txScope := outport.TxScope{
		PaymentAddressAllocation:               NewPaymentAddressAllocationStore(tx),
		PaymentAddressIdempotency:              NewPaymentAddressIdempotencyStore(tx),
		PaymentReceiptTracking:                 NewPaymentReceiptTrackingStore(tx),
		PaymentReceiptStatusNotificationOutbox: NewPaymentReceiptStatusNotificationOutboxStore(tx),
	}
	if err := fn(txScope); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
