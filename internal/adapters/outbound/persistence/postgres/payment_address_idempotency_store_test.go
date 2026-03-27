package postgres

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

func newFindPaymentAddressIdempotencyInput(idempotencyKey string) outport.FindPaymentAddressIdempotencyInput {
	return outport.FindPaymentAddressIdempotencyInput{
		Chain:          valueobjects.SupportedChainBitcoin,
		IdempotencyKey: idempotencyKey,
	}
}

func newClaimPaymentAddressIdempotencyInput(idempotencyKey string) outport.ClaimPaymentAddressIdempotencyInput {
	return outport.ClaimPaymentAddressIdempotencyInput{
		Chain:               valueobjects.SupportedChainBitcoin,
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

func TestPaymentAddressIdempotencyStoreFindByKeyRejectsInvalidPersistedChain(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressIdempotencyStore(db)

	rows := sqlmock.NewRows([]string{
		"chain",
		"address_policy_id",
		"expected_amount_minor",
		"customer_reference",
		"payment_address_id",
	}).AddRow(
		"bad/chain",
		"bitcoin-mainnet-native-segwit",
		int64(125000),
		"order-idem",
		int64(77),
	)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT chain,")).
		WithArgs("bitcoin", "idem-invalid").
		WillReturnRows(rows)

	_, _, err = store.FindByKey(context.Background(), newFindPaymentAddressIdempotencyInput("idem-invalid"))
	if !errors.Is(err, outport.ErrPaymentAddressIdempotencyPersistedChainInvalid) {
		t.Fatalf("unexpected invalid chain error: %v", err)
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

func TestPaymentAddressIdempotencyStoreClaimValidation(t *testing.T) {
	store := NewPaymentAddressIdempotencyStore(&stubNotificationExecutor{})
	validInput := newClaimPaymentAddressIdempotencyInput("idem-claim")

	testCases := []struct {
		name    string
		input   outport.ClaimPaymentAddressIdempotencyInput
		wantErr error
	}{
		{
			name: "missing chain",
			input: outport.ClaimPaymentAddressIdempotencyInput{
				IdempotencyKey:      validInput.IdempotencyKey,
				AddressPolicyID:     validInput.AddressPolicyID,
				ExpectedAmountMinor: validInput.ExpectedAmountMinor,
				CustomerReference:   validInput.CustomerReference,
			},
			wantErr: outport.ErrPaymentAddressIdempotencyChainRequired,
		},
		{
			name: "missing key",
			input: outport.ClaimPaymentAddressIdempotencyInput{
				Chain:               validInput.Chain,
				IdempotencyKey:      "   ",
				AddressPolicyID:     validInput.AddressPolicyID,
				ExpectedAmountMinor: validInput.ExpectedAmountMinor,
				CustomerReference:   validInput.CustomerReference,
			},
			wantErr: outport.ErrPaymentAddressIdempotencyKeyRequired,
		},
		{
			name: "missing policy",
			input: outport.ClaimPaymentAddressIdempotencyInput{
				Chain:               validInput.Chain,
				IdempotencyKey:      validInput.IdempotencyKey,
				AddressPolicyID:     " ",
				ExpectedAmountMinor: validInput.ExpectedAmountMinor,
				CustomerReference:   validInput.CustomerReference,
			},
			wantErr: outport.ErrPaymentAddressIdempotencyAddressPolicyIDRequired,
		},
		{
			name: "invalid amount",
			input: outport.ClaimPaymentAddressIdempotencyInput{
				Chain:               validInput.Chain,
				IdempotencyKey:      validInput.IdempotencyKey,
				AddressPolicyID:     validInput.AddressPolicyID,
				ExpectedAmountMinor: 0,
				CustomerReference:   validInput.CustomerReference,
			},
			wantErr: outport.ErrPaymentAddressIdempotencyExpectedAmountInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := store.Claim(context.Background(), tc.input)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("unexpected error: got %v want %v", err, tc.wantErr)
			}
		})
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
		Chain:            valueobjects.SupportedChainBitcoin,
		IdempotencyKey:   " idem-complete ",
		PaymentAddressID: 99,
	})
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
}

func TestPaymentAddressIdempotencyStoreCompleteValidation(t *testing.T) {
	store := NewPaymentAddressIdempotencyStore(&stubNotificationExecutor{})

	testCases := []struct {
		name    string
		input   outport.CompletePaymentAddressIdempotencyInput
		wantErr error
	}{
		{
			name: "missing chain",
			input: outport.CompletePaymentAddressIdempotencyInput{
				IdempotencyKey:   "idem-complete",
				PaymentAddressID: 99,
			},
			wantErr: outport.ErrPaymentAddressIdempotencyChainRequired,
		},
		{
			name: "missing key",
			input: outport.CompletePaymentAddressIdempotencyInput{
				Chain:            valueobjects.SupportedChainBitcoin,
				IdempotencyKey:   " ",
				PaymentAddressID: 99,
			},
			wantErr: outport.ErrPaymentAddressIdempotencyKeyRequired,
		},
		{
			name: "invalid payment address id",
			input: outport.CompletePaymentAddressIdempotencyInput{
				Chain:            valueobjects.SupportedChainBitcoin,
				IdempotencyKey:   "idem-complete",
				PaymentAddressID: 0,
			},
			wantErr: outport.ErrPaymentAddressIdempotencyPaymentAddressIDInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := store.Complete(context.Background(), tc.input)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("unexpected error: got %v want %v", err, tc.wantErr)
			}
		})
	}
}

func TestPaymentAddressIdempotencyStoreCompleteClaimNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressIdempotencyStore(db)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE payment_address_idempotency_keys")).
		WithArgs("bitcoin", "idem-complete", int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = store.Complete(context.Background(), outport.CompletePaymentAddressIdempotencyInput{
		Chain:            valueobjects.SupportedChainBitcoin,
		IdempotencyKey:   "idem-complete",
		PaymentAddressID: 99,
	})
	if !errors.Is(err, outport.ErrPaymentAddressIdempotencyClaimNotFound) {
		t.Fatalf("unexpected claim not found error: %v", err)
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
		Chain:          valueobjects.SupportedChainBitcoin,
		IdempotencyKey: " idem-release ",
	})
	if err != nil {
		t.Fatalf("Release returned error: %v", err)
	}
}

func TestPaymentAddressIdempotencyStoreReleaseValidation(t *testing.T) {
	store := NewPaymentAddressIdempotencyStore(&stubNotificationExecutor{})

	testCases := []struct {
		name    string
		input   outport.ReleasePaymentAddressIdempotencyInput
		wantErr error
	}{
		{
			name: "missing chain",
			input: outport.ReleasePaymentAddressIdempotencyInput{
				IdempotencyKey: "idem-release",
			},
			wantErr: outport.ErrPaymentAddressIdempotencyChainRequired,
		},
		{
			name: "missing key",
			input: outport.ReleasePaymentAddressIdempotencyInput{
				Chain:          valueobjects.SupportedChainBitcoin,
				IdempotencyKey: " ",
			},
			wantErr: outport.ErrPaymentAddressIdempotencyKeyRequired,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := store.Release(context.Background(), tc.input)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("unexpected error: got %v want %v", err, tc.wantErr)
			}
		})
	}
}

func TestPaymentAddressIdempotencyStoreReleaseClaimNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressIdempotencyStore(db)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM payment_address_idempotency_keys")).
		WithArgs("bitcoin", "idem-release").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = store.Release(context.Background(), outport.ReleasePaymentAddressIdempotencyInput{
		Chain:          valueobjects.SupportedChainBitcoin,
		IdempotencyKey: "idem-release",
	})
	if !errors.Is(err, outport.ErrPaymentAddressIdempotencyClaimNotFound) {
		t.Fatalf("unexpected claim not found error: %v", err)
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
