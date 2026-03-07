package postgres

import (
	"context"
	"database/sql"
	"errors"

	outport "payrune/internal/application/ports/out"
)

type TxScopeBuilder func(tx *sql.Tx) outport.TxScope

type UnitOfWork struct {
	db           *sql.DB
	buildTxScope TxScopeBuilder
}

func NewTxScope(tx *sql.Tx) outport.TxScope {
	return outport.TxScope{
		PaymentAddressAllocation:               NewPaymentAddressAllocationStore(tx),
		PaymentReceiptTracking:                 NewPaymentReceiptTrackingStore(tx),
		PaymentReceiptStatusNotificationOutbox: NewPaymentReceiptStatusNotificationOutboxStore(tx),
	}
}

func NewUnitOfWork(db *sql.DB, buildTxScope TxScopeBuilder) *UnitOfWork {
	return &UnitOfWork{
		db:           db,
		buildTxScope: buildTxScope,
	}
}

func (u *UnitOfWork) WithinTransaction(
	ctx context.Context,
	fn func(txScope outport.TxScope) error,
) error {
	if u.buildTxScope == nil {
		return errors.New("tx scope builder is not configured")
	}

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	txScope := u.buildTxScope(tx)
	if err := fn(txScope); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
