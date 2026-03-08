package postgres

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"

	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/value_objects"
)

func newFindPaymentAddressIdempotencyInput(idempotencyKey string) outport.FindPaymentAddressIdempotencyInput {
	return outport.FindPaymentAddressIdempotencyInput{
		Chain:          value_objects.SupportedChainBitcoin,
		IdempotencyKey: idempotencyKey,
	}
}

func newClaimPaymentAddressIdempotencyInput(idempotencyKey string) outport.ClaimPaymentAddressIdempotencyInput {
	return outport.ClaimPaymentAddressIdempotencyInput{
		Chain:               value_objects.SupportedChainBitcoin,
		IdempotencyKey:      idempotencyKey,
		AddressPolicyID:     "bitcoin-mainnet-native-segwit",
		ExpectedAmountMinor: 125000,
		CustomerReference:   "order-idem",
	}
}

func TestPaymentAddressIdempotencyStoreFindByKeyBlankKey(t *testing.T) {
	store := NewPaymentAddressIdempotencyStore(&stubNotificationExecutor{})

	record, found, err := store.FindByKey(context.Background(), newFindPaymentAddressIdempotencyInput("   "))
	if err != nil {
		t.Fatalf("FindByKey returned error: %v", err)
	}
	if found {
		t.Fatal("expected found=false")
	}
	if record != (outport.PaymentAddressIdempotencyRecord{}) {
		t.Fatalf("unexpected record: %+v", record)
	}
}

func TestPaymentAddressIdempotencyStoreFindByKeySuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressIdempotencyStore(db)
	input := newFindPaymentAddressIdempotencyInput(" idem-lookup ")

	rows := sqlmock.NewRows([]string{
		"chain",
		"address_policy_id",
		"expected_amount_minor",
		"customer_reference",
		"payment_address_id",
	}).AddRow(
		"bitcoin",
		"bitcoin-mainnet-native-segwit",
		int64(125000),
		"order-idem",
		int64(77),
	)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT chain,")).
		WithArgs("bitcoin", "idem-lookup").
		WillReturnRows(rows)

	record, found, err := store.FindByKey(context.Background(), input)
	if err != nil {
		t.Fatalf("FindByKey returned error: %v", err)
	}
	if !found {
		t.Fatal("expected found=true")
	}
	if record.PaymentAddressID != 77 {
		t.Fatalf("unexpected payment address id: got %d", record.PaymentAddressID)
	}
	if record.AddressPolicyID != "bitcoin-mainnet-native-segwit" {
		t.Fatalf("unexpected address policy id: got %q", record.AddressPolicyID)
	}
	if record.CustomerReference != "order-idem" {
		t.Fatalf("unexpected customer reference: got %q", record.CustomerReference)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPaymentAddressIdempotencyStoreClaimSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressIdempotencyStore(db)
	input := newClaimPaymentAddressIdempotencyInput(" idem-claim ")

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO payment_address_idempotency_keys")).
		WithArgs(
			"bitcoin",
			"idem-claim",
			"bitcoin-mainnet-native-segwit",
			int64(125000),
			"order-idem",
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	record, err := store.Claim(context.Background(), input)
	if err != nil {
		t.Fatalf("Claim returned error: %v", err)
	}
	if record.IdempotencyKey != "idem-claim" {
		t.Fatalf("unexpected idempotency key: got %q", record.IdempotencyKey)
	}
	if record.PaymentAddressID != 0 {
		t.Fatalf("unexpected payment address id: got %d", record.PaymentAddressID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPaymentAddressIdempotencyStoreClaimMapsDuplicateKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressIdempotencyStore(db)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO payment_address_idempotency_keys")).
		WillReturnError(&pq.Error{Code: "23505", Constraint: paymentAddressIdempotencyPrimaryKey})

	_, err = store.Claim(context.Background(), newClaimPaymentAddressIdempotencyInput("idem-dup"))
	if !errors.Is(err, outport.ErrPaymentAddressIdempotencyKeyExists) {
		t.Fatalf("expected ErrPaymentAddressIdempotencyKeyExists, got %v", err)
	}
}

func TestPaymentAddressIdempotencyStoreCompleteSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressIdempotencyStore(db)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE payment_address_idempotency_keys")).
		WithArgs("bitcoin", "idem-complete", int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = store.Complete(context.Background(), outport.CompletePaymentAddressIdempotencyInput{
		Chain:            value_objects.SupportedChainBitcoin,
		IdempotencyKey:   " idem-complete ",
		PaymentAddressID: 99,
	})
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
}

func TestPaymentAddressIdempotencyStoreReleaseSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressIdempotencyStore(db)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM payment_address_idempotency_keys")).
		WithArgs("bitcoin", "idem-release").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = store.Release(context.Background(), outport.ReleasePaymentAddressIdempotencyInput{
		Chain:          value_objects.SupportedChainBitcoin,
		IdempotencyKey: " idem-release ",
	})
	if err != nil {
		t.Fatalf("Release returned error: %v", err)
	}
}

func TestPaymentAddressIdempotencyStoreFindByKeyNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressIdempotencyStore(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT chain,")).
		WithArgs("bitcoin", "idem-missing").
		WillReturnError(sql.ErrNoRows)

	record, found, err := store.FindByKey(context.Background(), newFindPaymentAddressIdempotencyInput("idem-missing"))
	if err != nil {
		t.Fatalf("FindByKey returned error: %v", err)
	}
	if found {
		t.Fatal("expected found=false")
	}
	if record != (outport.PaymentAddressIdempotencyRecord{}) {
		t.Fatalf("unexpected record: %+v", record)
	}
}
