package cloudflarepostgres

import (
	"context"
	"errors"

	outport "payrune/internal/application/ports/out"
)

type UnitOfWork struct {
	bridgeID string
	bridge   Bridge
}

func NewUnitOfWork(bridgeID string, bridge Bridge) *UnitOfWork {
	return &UnitOfWork{
		bridgeID: bridgeID,
		bridge:   bridge,
	}
}

func (u *UnitOfWork) WithinTransaction(
	ctx context.Context,
	fn func(txScope outport.TxScope) error,
) error {
	if u.bridge == nil {
		return errors.New("cloudflare postgres bridge is not configured")
	}

	txID, err := u.bridge.BeginTx(ctx, u.bridgeID)
	if err != nil {
		return err
	}

	txExecutor := newTxExecutor(u.bridgeID, txID, u.bridge)
	txScope := outport.TxScope{
		PaymentAddressAllocation:               NewPaymentAddressAllocationStore(txExecutor),
		PaymentAddressIdempotency:              NewPaymentAddressIdempotencyStore(txExecutor),
		PaymentReceiptTracking:                 NewPaymentReceiptTrackingStore(txExecutor),
		PaymentReceiptStatusNotificationOutbox: NewPaymentReceiptStatusNotificationOutboxStore(txExecutor),
	}

	if err := fn(txScope); err != nil {
		if rollbackErr := u.bridge.RollbackTx(ctx, u.bridgeID, txID); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	return u.bridge.CommitTx(ctx, u.bridgeID, txID)
}
