package postgres

import (
	"context"
	"database/sql"
	"errors"

	outport "payrune/internal/application/ports/out"
)

type TxRepositoriesBuilder func(tx *sql.Tx) outport.TxRepositories

type UnitOfWork struct {
	db                  *sql.DB
	buildTxRepositories TxRepositoriesBuilder
}

func NewTxRepositories(tx *sql.Tx) outport.TxRepositories {
	return outport.TxRepositories{
		PaymentAddressAllocation:         NewPaymentAddressAllocationRepository(tx),
		PaymentReceiptTracking:           NewPaymentReceiptTrackingRepository(tx),
		PaymentReceiptStatusNotification: NewPaymentReceiptStatusNotificationRepository(tx),
	}
}

func NewUnitOfWork(db *sql.DB, buildTxRepositories TxRepositoriesBuilder) *UnitOfWork {
	return &UnitOfWork{
		db:                  db,
		buildTxRepositories: buildTxRepositories,
	}
}

func (u *UnitOfWork) WithinTransaction(
	ctx context.Context,
	fn func(txRepositories outport.TxRepositories) error,
) error {
	if u.buildTxRepositories == nil {
		return errors.New("tx repositories builder is not configured")
	}

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	txRepositories := u.buildTxRepositories(tx)
	if err := fn(txRepositories); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
