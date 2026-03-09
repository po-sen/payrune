package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/value_objects"
)

type stubScanner struct {
	values []any
	err    error
}

func (s stubScanner) Scan(dest ...any) error {
	if s.err != nil {
		return s.err
	}
	if len(dest) != len(s.values) {
		return fmt.Errorf("unexpected scan arg count: got %d want %d", len(dest), len(s.values))
	}

	for i := range dest {
		destValue := reflect.ValueOf(dest[i])
		if destValue.Kind() != reflect.Ptr {
			return fmt.Errorf("dest %d must be pointer", i)
		}
		target := destValue.Elem()

		source := reflect.ValueOf(s.values[i])
		if !source.IsValid() {
			target.Set(reflect.Zero(target.Type()))
			continue
		}

		if source.Type().AssignableTo(target.Type()) {
			target.Set(source)
			continue
		}
		if source.Type().ConvertibleTo(target.Type()) {
			target.Set(source.Convert(target.Type()))
			continue
		}
		return fmt.Errorf("value %d type mismatch: %s -> %s", i, source.Type(), target.Type())
	}

	return nil
}

func TestScanPaymentReceiptTrackingSupportsGenericChainNetwork(t *testing.T) {
	now := time.Date(2026, 3, 5, 15, 0, 0, 0, time.UTC)
	tracking, err := scanPaymentReceiptTracking(stubScanner{
		values: []any{
			int64(1),   // id
			int64(2),   // payment_address_id
			"policy-1", // address_policy_id
			"ethereum", // chain
			"sepolia",  // network
			"0xabc",    // address
			sql.NullTime{Valid: true, Time: now},
			int64(100),   // expected_amount_minor
			int32(2),     // required_confirmations
			"watching",   // receipt_status
			int64(10),    // observed_total_minor
			int64(5),     // confirmed_total_minor
			int64(5),     // unconfirmed_total_minor
			int64(12345), // last_observed_block_height
			sql.NullTime{Valid: true, Time: now},
			sql.NullTime{}, // paid_at
			sql.NullTime{}, // confirmed_at
			sql.NullTime{Valid: true, Time: now.Add(24 * time.Hour)},
			"", // last_error
		},
	})
	if err != nil {
		t.Fatalf("scanPaymentReceiptTracking returned error: %v", err)
	}

	if tracking.Chain != "ethereum" {
		t.Fatalf("unexpected chain: got %q", tracking.Chain)
	}
	if tracking.Network != "sepolia" {
		t.Fatalf("unexpected network: got %q", tracking.Network)
	}
	if tracking.FirstObservedAt == nil || !tracking.FirstObservedAt.Equal(now) {
		t.Fatalf("unexpected first observed at: got %+v", tracking.FirstObservedAt)
	}
	if tracking.IssuedAt.IsZero() || !tracking.IssuedAt.Equal(now) {
		t.Fatalf("unexpected issued at: got %s", tracking.IssuedAt)
	}
	if tracking.ExpiresAt == nil || !tracking.ExpiresAt.Equal(now.Add(24*time.Hour)) {
		t.Fatalf("unexpected expires at: got %+v", tracking.ExpiresAt)
	}
}

func TestScanPaymentReceiptTrackingRejectsInvalidNetwork(t *testing.T) {
	_, err := scanPaymentReceiptTracking(stubScanner{
		values: []any{
			int64(1),
			int64(2),
			"policy-1",
			"bitcoin",
			"main/net",
			"tb1qabc",
			sql.NullTime{Valid: true, Time: time.Now().UTC()},
			int64(100),
			int32(1),
			"watching",
			int64(0),
			int64(0),
			int64(0),
			int64(0),
			sql.NullTime{},
			sql.NullTime{},
			sql.NullTime{},
			sql.NullTime{},
			"",
		},
	})
	if err == nil {
		t.Fatal("expected invalid network error")
	}
}

func TestScanPaymentReceiptTrackingRejectsInvalidChain(t *testing.T) {
	_, err := scanPaymentReceiptTracking(stubScanner{
		values: []any{
			int64(1),
			int64(2),
			"policy-1",
			"eth/mainnet",
			"mainnet",
			"0xabc",
			sql.NullTime{Valid: true, Time: time.Now().UTC()},
			int64(100),
			int32(1),
			"watching",
			int64(0),
			int64(0),
			int64(0),
			int64(0),
			sql.NullTime{},
			sql.NullTime{},
			sql.NullTime{},
			sql.NullTime{},
			"",
		},
	})
	if err == nil {
		t.Fatal("expected invalid chain error")
	}
}

func TestScanPaymentReceiptTrackingRejectsInvalidStatus(t *testing.T) {
	_, err := scanPaymentReceiptTracking(stubScanner{
		values: []any{
			int64(1),
			int64(2),
			"policy-1",
			"bitcoin",
			"mainnet",
			"bc1qabc",
			sql.NullTime{Valid: true, Time: time.Now().UTC()},
			int64(100),
			int32(1),
			"unknown",
			int64(0),
			int64(0),
			int64(0),
			int64(0),
			int64(0),
			sql.NullTime{},
			sql.NullTime{},
			sql.NullTime{},
			sql.NullTime{},
			"",
		},
	})
	if err == nil {
		t.Fatal("expected invalid status error")
	}
}

func newTrackingStoreTestEntity() entities.PaymentReceiptTracking {
	issuedAt := time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC)
	expiresAt := issuedAt.Add(24 * time.Hour)
	firstObservedAt := issuedAt.Add(2 * time.Minute)
	return entities.PaymentReceiptTracking{
		PaymentAddressID:        501,
		AddressPolicyID:         "bitcoin-mainnet-native-segwit",
		Chain:                   value_objects.ChainIDBitcoin,
		Network:                 value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
		Address:                 "bc1qtracking",
		IssuedAt:                issuedAt,
		ExpiresAt:               &expiresAt,
		ExpectedAmountMinor:     100000,
		RequiredConfirmations:   2,
		Status:                  value_objects.PaymentReceiptStatusWatching,
		ObservedTotalMinor:      1000,
		ConfirmedTotalMinor:     500,
		UnconfirmedTotalMinor:   500,
		LastObservedBlockHeight: 123,
		FirstObservedAt:         &firstObservedAt,
		LastError:               "observer warning",
	}
}

func TestPaymentReceiptTrackingStoreCreateValidation(t *testing.T) {
	store := NewPaymentReceiptTrackingStore(&stubNotificationExecutor{})

	err := store.Create(context.Background(), entities.PaymentReceiptTracking{}, time.Time{})
	if err == nil || err.Error() != "next poll at is required" {
		t.Fatalf("unexpected error: got %v", err)
	}
}

func TestPaymentReceiptTrackingStoreCreateSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentReceiptTrackingStore(db)
	tracking := newTrackingStoreTestEntity()
	nextPollAt := time.Date(2026, 3, 7, 12, 15, 0, 0, time.UTC)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO payment_receipt_trackings")).
		WithArgs(
			int64(501),
			"bitcoin-mainnet-native-segwit",
			"bitcoin",
			"mainnet",
			"bc1qtracking",
			tracking.IssuedAt.UTC(),
			tracking.ExpiresAt.UTC(),
			int64(100000),
			int32(2),
			"watching",
			int64(1000),
			int64(500),
			int64(500),
			int64(123),
			tracking.FirstObservedAt.UTC(),
			nil,
			nil,
			"observer warning",
			nextPollAt.UTC(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := store.Create(context.Background(), tracking, nextPollAt); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
}

func TestPaymentReceiptTrackingStoreCreateAlreadyExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentReceiptTrackingStore(db)
	tracking := newTrackingStoreTestEntity()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO payment_receipt_trackings")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = store.Create(context.Background(), tracking, time.Now().UTC())
	if err == nil || err.Error() != "payment receipt tracking already exists" {
		t.Fatalf("unexpected error: got %v", err)
	}
}

func TestPaymentReceiptTrackingStoreClaimDueValidation(t *testing.T) {
	store := NewPaymentReceiptTrackingStore(&stubNotificationExecutor{})
	now := time.Now().UTC()

	_, err := store.ClaimDue(context.Background(), outport.ClaimPaymentReceiptTrackingsInput{
		Limit:      1,
		ClaimUntil: now,
		Statuses:   []value_objects.PaymentReceiptStatus{value_objects.PaymentReceiptStatusWatching},
	})
	if err == nil || err.Error() != "claim now is required" {
		t.Fatalf("unexpected missing-now error: %v", err)
	}

	_, err = store.ClaimDue(context.Background(), outport.ClaimPaymentReceiptTrackingsInput{
		Now:      now,
		Limit:    1,
		Statuses: []value_objects.PaymentReceiptStatus{value_objects.PaymentReceiptStatusWatching},
	})
	if err == nil || err.Error() != "claim until is required" {
		t.Fatalf("unexpected missing-claim-until error: %v", err)
	}

	_, err = store.ClaimDue(context.Background(), outport.ClaimPaymentReceiptTrackingsInput{
		Now:        now,
		ClaimUntil: now,
		Statuses:   []value_objects.PaymentReceiptStatus{value_objects.PaymentReceiptStatusWatching},
	})
	if err == nil || err.Error() != "claim limit must be greater than zero" {
		t.Fatalf("unexpected missing-limit error: %v", err)
	}

	_, err = store.ClaimDue(context.Background(), outport.ClaimPaymentReceiptTrackingsInput{
		Now:        now,
		ClaimUntil: now,
		Limit:      1,
	})
	if err == nil || err.Error() != "claim statuses are required" {
		t.Fatalf("unexpected missing-statuses error: %v", err)
	}

	_, err = store.ClaimDue(context.Background(), outport.ClaimPaymentReceiptTrackingsInput{
		Now:        now,
		ClaimUntil: now,
		Limit:      1,
		Statuses:   []value_objects.PaymentReceiptStatus{""},
	})
	if err == nil || err.Error() != "claim status is required" {
		t.Fatalf("unexpected blank-status error: %v", err)
	}
}

func TestPaymentReceiptTrackingStoreClaimDueSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentReceiptTrackingStore(db)
	now := time.Date(2026, 3, 7, 14, 0, 0, 0, time.UTC)
	claimUntil := now.Add(30 * time.Second)
	expiresAt := now.Add(2 * time.Hour)
	firstObservedAt := now.Add(-2 * time.Minute)
	paidAt := now.Add(-1 * time.Minute)
	confirmedAt := now

	rows := sqlmock.NewRows([]string{
		"id",
		"payment_address_id",
		"address_policy_id",
		"chain",
		"network",
		"address",
		"issued_at",
		"expected_amount_minor",
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
		int64(1),
		int64(501),
		"bitcoin-mainnet-native-segwit",
		"bitcoin",
		"mainnet",
		"bc1qtracking",
		now,
		int64(100000),
		int32(2),
		"paid_unconfirmed",
		int64(100000),
		int64(50000),
		int64(50000),
		int64(321),
		firstObservedAt,
		paidAt,
		confirmedAt,
		expiresAt,
		"",
	)

	mock.ExpectQuery(regexp.QuoteMeta("WITH due AS")).
		WithArgs(
			now,
			2,
			claimUntil,
			sqlmock.AnyArg(),
			"bitcoin",
			"mainnet",
		).
		WillReturnRows(rows)

	trackings, err := store.ClaimDue(context.Background(), outport.ClaimPaymentReceiptTrackingsInput{
		Now:        now,
		ClaimUntil: claimUntil,
		Limit:      2,
		Statuses: []value_objects.PaymentReceiptStatus{
			value_objects.PaymentReceiptStatusWatching,
			value_objects.PaymentReceiptStatusPaidUnconfirmed,
		},
		Chain:   " BitCoin ",
		Network: " MainNet ",
	})
	if err != nil {
		t.Fatalf("ClaimDue returned error: %v", err)
	}
	if len(trackings) != 1 {
		t.Fatalf("unexpected tracking count: got %d", len(trackings))
	}
	if trackings[0].PaymentAddressID != 501 {
		t.Fatalf("unexpected payment address id: got %d", trackings[0].PaymentAddressID)
	}
	if trackings[0].Status != value_objects.PaymentReceiptStatusPaidUnconfirmed {
		t.Fatalf("unexpected status: got %q", trackings[0].Status)
	}
}

func TestPaymentReceiptTrackingStoreSaveValidation(t *testing.T) {
	store := NewPaymentReceiptTrackingStore(&stubNotificationExecutor{})

	err := store.Save(context.Background(), entities.PaymentReceiptTracking{}, time.Time{}, time.Now().UTC())
	if err == nil || err.Error() != "polled at is required" {
		t.Fatalf("unexpected missing-polled-at error: %v", err)
	}

	err = store.Save(context.Background(), entities.PaymentReceiptTracking{}, time.Now().UTC(), time.Time{})
	if err == nil || err.Error() != "next poll at is required" {
		t.Fatalf("unexpected missing-next-poll-at error: %v", err)
	}
}

func TestPaymentReceiptTrackingStoreSaveSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentReceiptTrackingStore(db)
	tracking := newTrackingStoreTestEntity()
	paidAt := tracking.IssuedAt.Add(5 * time.Minute)
	tracking.PaidAt = &paidAt
	tracking.Status = value_objects.PaymentReceiptStatusPaidUnconfirmed
	polledAt := time.Date(2026, 3, 7, 14, 0, 0, 0, time.UTC)
	nextPollAt := polledAt.Add(15 * time.Second)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE payment_receipt_trackings")).
		WithArgs(
			int64(501),
			"paid_unconfirmed",
			int64(1000),
			int64(500),
			int64(500),
			int64(123),
			tracking.FirstObservedAt.UTC(),
			paidAt.UTC(),
			nil,
			tracking.ExpiresAt.UTC(),
			"observer warning",
			polledAt.UTC(),
			nextPollAt.UTC(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := store.Save(context.Background(), tracking, polledAt, nextPollAt); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
}

func TestPaymentReceiptTrackingStoreSaveNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentReceiptTrackingStore(db)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE payment_receipt_trackings")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = store.Save(context.Background(), newTrackingStoreTestEntity(), time.Now().UTC(), time.Now().UTC().Add(time.Minute))
	if err == nil || err.Error() != "payment receipt tracking is not found" {
		t.Fatalf("unexpected error: got %v", err)
	}
}
