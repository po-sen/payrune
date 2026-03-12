package postgres

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

func newFindPaymentAddressStatusInput(paymentAddressID int64) outport.FindPaymentAddressStatusInput {
	return outport.FindPaymentAddressStatusInput{
		Chain:            valueobjects.SupportedChainBitcoin,
		PaymentAddressID: paymentAddressID,
	}
}

func TestPaymentAddressStatusFinderFindByIDInvalidID(t *testing.T) {
	finder := NewPaymentAddressStatusFinder(&stubNotificationExecutor{})

	record, found, err := finder.FindByID(context.Background(), newFindPaymentAddressStatusInput(0))
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if found {
		t.Fatal("expected found=false")
	}
	if record != (outport.PaymentAddressStatusRecord{}) {
		t.Fatalf("unexpected record: %+v", record)
	}
}

func TestPaymentAddressStatusFinderFindByIDSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	finder := NewPaymentAddressStatusFinder(db)
	issuedAt := time.Date(2026, 3, 8, 11, 0, 0, 0, time.UTC)
	firstObservedAt := issuedAt.Add(5 * time.Minute)

	rows := sqlmock.NewRows([]string{
		"id",
		"address_policy_id",
		"expected_amount_minor",
		"customer_reference",
		"chain",
		"network",
		"scheme",
		"address",
		"issued_at",
		"tracking_id",
		"required_confirmations",
		"receipt_status",
		"observed_total_minor",
		"confirmed_total_minor",
		"unconfirmed_total_minor",
		"last_observed_block_height",
		"first_observed_at",
		"paid_at",
		"confirmed_at",
		"expires_at",
		"last_error",
	}).AddRow(
		int64(101),
		"bitcoin-mainnet-native-segwit",
		int64(120000),
		"order-20260308-001",
		"bitcoin",
		"mainnet",
		"nativeSegwit",
		"bc1qstatus",
		issuedAt,
		int64(55),
		int32(1),
		"paid_unconfirmed_reverted",
		int64(80000),
		int64(40000),
		int64(40000),
		int64(123),
		firstObservedAt,
		nil,
		nil,
		nil,
		"",
	)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs("bitcoin", int64(101)).
		WillReturnRows(rows)

	record, found, err := finder.FindByID(context.Background(), newFindPaymentAddressStatusInput(101))
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if !found {
		t.Fatal("expected found=true")
	}
	if record.PaymentAddressID != 101 {
		t.Fatalf("unexpected payment address id: got %d", record.PaymentAddressID)
	}
	if record.PaymentStatus != valueobjects.PaymentReceiptStatusPaidUnconfirmedReverted {
		t.Fatalf("unexpected payment status: got %q", record.PaymentStatus)
	}
	if record.FirstObservedAt == nil || !record.FirstObservedAt.Equal(firstObservedAt) {
		t.Fatalf("unexpected first observed at: got %v", record.FirstObservedAt)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPaymentAddressStatusFinderFindByIDNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	finder := NewPaymentAddressStatusFinder(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs("bitcoin", int64(404)).
		WillReturnError(sql.ErrNoRows)

	record, found, err := finder.FindByID(context.Background(), newFindPaymentAddressStatusInput(404))
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if found {
		t.Fatal("expected found=false")
	}
	if record != (outport.PaymentAddressStatusRecord{}) {
		t.Fatalf("unexpected record: %+v", record)
	}
}

func TestPaymentAddressStatusFinderFindByIDIncompleteTracking(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	finder := NewPaymentAddressStatusFinder(db)
	issuedAt := time.Date(2026, 3, 8, 11, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"id",
		"address_policy_id",
		"expected_amount_minor",
		"customer_reference",
		"chain",
		"network",
		"scheme",
		"address",
		"issued_at",
		"tracking_id",
		"required_confirmations",
		"receipt_status",
		"observed_total_minor",
		"confirmed_total_minor",
		"unconfirmed_total_minor",
		"last_observed_block_height",
		"first_observed_at",
		"paid_at",
		"confirmed_at",
		"expires_at",
		"last_error",
	}).AddRow(
		int64(101),
		"bitcoin-mainnet-native-segwit",
		int64(120000),
		"order-20260308-001",
		"bitcoin",
		"mainnet",
		"nativeSegwit",
		"bc1qstatus",
		issuedAt,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs("bitcoin", int64(101)).
		WillReturnRows(rows)

	_, _, err = finder.FindByID(context.Background(), newFindPaymentAddressStatusInput(101))
	if !errors.Is(err, outport.ErrPaymentAddressStatusIncomplete) {
		t.Fatalf("expected ErrPaymentAddressStatusIncomplete, got %v", err)
	}
}
