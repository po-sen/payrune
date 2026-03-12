package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	outport "payrune/internal/application/ports/outbound"
)

func TestUnitOfWorkWithinTransactionValidation(t *testing.T) {
	uow := NewUnitOfWork(nil)

	err := uow.WithinTransaction(context.Background(), func(outport.TxScope) error {
		return nil
	})
	if err == nil || err.Error() != "database is not configured" {
		t.Fatalf("unexpected error: got %v", err)
	}
}

func TestUnitOfWorkWithinTransactionCommitsOnSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	uow := NewUnitOfWork(db)
	called := false

	err = uow.WithinTransaction(context.Background(), func(txScope outport.TxScope) error {
		called = true
		if txScope.PaymentAddressAllocation == nil {
			t.Fatal("expected allocation store in tx scope")
		}
		if txScope.PaymentReceiptTracking == nil {
			t.Fatal("expected tracking store in tx scope")
		}
		if txScope.PaymentReceiptStatusNotificationOutbox == nil {
			t.Fatal("expected outbox in tx scope")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WithinTransaction returned error: %v", err)
	}
	if !called {
		t.Fatal("expected callback to be called")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUnitOfWorkWithinTransactionRollsBackOnError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	uow := NewUnitOfWork(db)
	expectedErr := errors.New("boom")

	err = uow.WithinTransaction(context.Background(), func(outport.TxScope) error {
		return expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected callback error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
