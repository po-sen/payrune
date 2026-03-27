package cloudflarepostgres

import (
	"context"

	outport "payrune/internal/application/ports/outbound"
	cloudflarepostgresinfra "payrune/internal/infrastructure/cloudflarepostgres"
)

type UnitOfWork struct {
	bridgeID string
	bridge   cloudflarepostgresinfra.Bridge
}

func NewUnitOfWork(bridgeID string, bridge cloudflarepostgresinfra.Bridge) *UnitOfWork {
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
		return outport.ErrUnitOfWorkNotConfigured
	}

	txID, err := u.bridge.BeginTx(ctx, u.bridgeID)
	if err != nil {
		return outport.ErrUnitOfWorkFailed
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
			return outport.ErrUnitOfWorkFailed
		}
		return err
	}

	if err := u.bridge.CommitTx(ctx, u.bridgeID, txID); err != nil {
		return outport.ErrUnitOfWorkFailed
	}
	return nil
}
