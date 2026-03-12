package cloudflarepostgres

import (
	"context"
	"testing"

	outport "payrune/internal/application/ports/out"
)

type fakeBridge struct {
	beginCalls    int
	commitCalls   int
	rollbackCalls int
}

func (f *fakeBridge) BeginTx(context.Context, string) (string, error) {
	f.beginCalls++
	return "tx-1", nil
}

func (f *fakeBridge) CommitTx(context.Context, string, string) error {
	f.commitCalls++
	return nil
}

func (f *fakeBridge) RollbackTx(context.Context, string, string) error {
	f.rollbackCalls++
	return nil
}

func (f *fakeBridge) Exec(context.Context, string, string, string, []any) (int64, error) {
	return 0, nil
}

func (f *fakeBridge) Query(context.Context, string, string, string, []any) ([][]any, error) {
	return nil, nil
}

func (f *fakeBridge) QueryRow(context.Context, string, string, string, []any) ([]any, bool, error) {
	return nil, false, nil
}

func TestUnitOfWorkWithinTransactionWiresNotificationOutbox(t *testing.T) {
	bridge := &fakeBridge{}
	unitOfWork := NewUnitOfWork("bridge-123", bridge)

	err := unitOfWork.WithinTransaction(context.Background(), func(txScope outport.TxScope) error {
		if txScope.PaymentReceiptStatusNotificationOutbox == nil {
			t.Fatal("expected payment receipt status notification outbox to be wired")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WithinTransaction returned error: %v", err)
	}
	if bridge.beginCalls != 1 || bridge.commitCalls != 1 || bridge.rollbackCalls != 0 {
		t.Fatalf("unexpected bridge call counts: %+v", bridge)
	}
}
