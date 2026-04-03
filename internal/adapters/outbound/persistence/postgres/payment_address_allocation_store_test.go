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
	"payrune/internal/domain/entities"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
)

func newAllocationStoreTestPolicy() policies.AddressIssuancePolicy {
	return policies.AddressIssuancePolicy{
		AddressPolicyID: "bitcoin-mainnet-native-segwit",
		Chain:           valueobjects.SupportedChainBitcoin,
		Network:         valueobjects.NetworkIDMainnet,
		Scheme:          valueobjects.AddressSchemeNativeSegwit,
		IssuanceConfig: valueobjects.AddressIssuanceConfig{
			AddressSpaceRef: "xpub-main",
		},
	}.Normalize()
}

func newReservePaymentAddressAllocationInput(customerReference string) outport.ReservePaymentAddressAllocationInput {
	return outport.ReservePaymentAddressAllocationInput{
		IssuancePolicy:      newAllocationStoreTestPolicy(),
		ExpectedAmountMinor: 125000,
		CustomerReference:   customerReference,
	}
}

func newFindIssuedPaymentAddressAllocationByIDInput(paymentAddressID int64) outport.FindIssuedPaymentAddressAllocationByIDInput {
	return outport.FindIssuedPaymentAddressAllocationByIDInput{
		PaymentAddressID: paymentAddressID,
	}
}

func TestPaymentAddressAllocationStoreCompleteValidation(t *testing.T) {
	store := NewPaymentAddressAllocationStore(&stubNotificationExecutor{})

	err := store.Complete(context.Background(), outport.CompletePaymentAddressAllocationInput{})
	if !errors.Is(err, outport.ErrPaymentAddressAllocationIssuedAtRequired) {
		t.Fatalf("unexpected error: got %v", err)
	}
}

func TestPaymentAddressAllocationStoreCompleteSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressAllocationStore(db)
	issuedAt := time.Date(2026, 3, 7, 9, 0, 0, 0, time.UTC)
	allocation := entities.PaymentAddressAllocation{
		PaymentAddressID: 44,
		AddressPolicyID:  "bitcoin-mainnet-native-segwit",
		Chain:            valueobjects.SupportedChainBitcoin,
		Network:          valueobjects.NetworkIDMainnet,
		Scheme:           valueobjects.AddressSchemeNativeSegwit,
		Address:          " bc1qallocated ",
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE address_policy_allocations")).
		WithArgs(
			int64(44),
			"bitcoin",
			"mainnet",
			"nativeSegwit",
			"bc1qallocated",
			`{"material_type":"bitcoin_hd"}`,
			issuedAt.UTC(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := store.Complete(context.Background(), outport.CompletePaymentAddressAllocationInput{
		Allocation:        allocation,
		SweepMaterialJSON: `{"material_type":"bitcoin_hd"}`,
		IssuedAt:          issuedAt,
	}); err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPaymentAddressAllocationStoreCompleteNotReserved(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressAllocationStore(db)
	issuedAt := time.Now().UTC()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE address_policy_allocations")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = store.Complete(context.Background(), outport.CompletePaymentAddressAllocationInput{
		Allocation: entities.PaymentAddressAllocation{
			PaymentAddressID: 44,
			AddressPolicyID:  "bitcoin-mainnet-native-segwit",
			Chain:            valueobjects.SupportedChainBitcoin,
			Network:          valueobjects.NetworkIDMainnet,
			Scheme:           valueobjects.AddressSchemeNativeSegwit,
			Address:          "bc1qallocated",
		},
		SweepMaterialJSON: `{"material_type":"bitcoin_hd"}`,
		IssuedAt:          issuedAt,
	})
	if !errors.Is(err, outport.ErrPaymentAddressAllocationNotReserved) {
		t.Fatalf("expected ErrPaymentAddressAllocationNotReserved, got %v", err)
	}
}

func TestPaymentAddressAllocationStoreCompleteRejectsInvalidSweepMaterialInput(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressAllocationStore(db)

	err = store.Complete(context.Background(), outport.CompletePaymentAddressAllocationInput{
		Allocation: entities.PaymentAddressAllocation{
			PaymentAddressID: 44,
			AddressPolicyID:  "bitcoin-mainnet-native-segwit",
			Chain:            valueobjects.SupportedChainBitcoin,
			Network:          valueobjects.NetworkIDMainnet,
			Scheme:           valueobjects.AddressSchemeNativeSegwit,
			Address:          "bc1qallocated",
		},
		SweepMaterialJSON: " ",
		IssuedAt:          time.Date(2026, 3, 7, 9, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, outport.ErrPaymentAddressAllocationStoreFailed) {
		t.Fatalf("expected ErrPaymentAddressAllocationStoreFailed, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected SQL call: %v", err)
	}
}

func TestPaymentAddressAllocationStoreFindIssuedByIDInvalidID(t *testing.T) {
	store := NewPaymentAddressAllocationStore(&stubNotificationExecutor{})

	allocation, found, err := store.FindIssuedByID(
		context.Background(),
		newFindIssuedPaymentAddressAllocationByIDInput(0),
	)
	if err != nil {
		t.Fatalf("FindIssuedByID returned error: %v", err)
	}
	if found {
		t.Fatal("expected found=false")
	}
	if allocation != (entities.PaymentAddressAllocation{}) {
		t.Fatalf("unexpected allocation: %+v", allocation)
	}
}

func TestPaymentAddressAllocationStoreFindIssuedByIDSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressAllocationStore(db)

	rows := sqlmock.NewRows([]string{
		"id",
		"address_policy_id",
		"slot_index",
		"expected_amount_minor",
		"customer_reference",
		"chain",
		"network",
		"scheme",
		"address",
		"failure_reason",
	}).AddRow(
		int64(199),
		"bitcoin-mainnet-native-segwit",
		int64(21),
		int64(125000),
		"order-lookup",
		"bitcoin",
		"mainnet",
		"nativeSegwit",
		"bc1qlookup",
		"",
	)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT a.id,")).
		WithArgs(int64(199)).
		WillReturnRows(rows)

	allocation, found, err := store.FindIssuedByID(
		context.Background(),
		newFindIssuedPaymentAddressAllocationByIDInput(199),
	)
	if err != nil {
		t.Fatalf("FindIssuedByID returned error: %v", err)
	}
	if !found {
		t.Fatal("expected found=true")
	}
	if allocation.PaymentAddressID != 199 {
		t.Fatalf("unexpected payment address id: got %d", allocation.PaymentAddressID)
	}
	if allocation.SlotIndex != 21 {
		t.Fatalf("unexpected slot index: got %d", allocation.SlotIndex)
	}
}

func TestPaymentAddressAllocationStoreFindIssuedByIDRejectsInvalidPersistedChain(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressAllocationStore(db)

	rows := sqlmock.NewRows([]string{
		"id",
		"address_policy_id",
		"slot_index",
		"expected_amount_minor",
		"customer_reference",
		"chain",
		"network",
		"scheme",
		"address",
		"failure_reason",
	}).AddRow(
		int64(199),
		"bitcoin-mainnet-native-segwit",
		int64(21),
		int64(125000),
		"order-lookup",
		"bad/chain",
		"mainnet",
		"nativeSegwit",
		"bc1qlookup",
		"",
	)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT a.id,")).
		WithArgs(int64(199)).
		WillReturnRows(rows)

	_, _, err = store.FindIssuedByID(context.Background(), newFindIssuedPaymentAddressAllocationByIDInput(199))
	if !errors.Is(err, outport.ErrPaymentAddressAllocationPersistedChainInvalid) {
		t.Fatalf("unexpected invalid chain error: %v", err)
	}
}

func TestPaymentAddressAllocationStoreFindIssuedByIDRejectsInvalidPersistedAddressPolicyID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressAllocationStore(db)

	rows := sqlmock.NewRows([]string{
		"id",
		"address_policy_id",
		"slot_index",
		"expected_amount_minor",
		"customer_reference",
		"chain",
		"network",
		"scheme",
		"address",
		"failure_reason",
	}).AddRow(
		int64(199),
		"bitcoin/mainnet",
		int64(21),
		int64(125000),
		"order-lookup",
		"bitcoin",
		"mainnet",
		"nativeSegwit",
		"bc1qlookup",
		"",
	)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT a.id,")).
		WithArgs(int64(199)).
		WillReturnRows(rows)

	_, _, err = store.FindIssuedByID(context.Background(), newFindIssuedPaymentAddressAllocationByIDInput(199))
	if !errors.Is(err, outport.ErrPaymentAddressAllocationPersistedAddressPolicyIDInvalid) {
		t.Fatalf("unexpected invalid address policy id error: %v", err)
	}
}

func TestPaymentAddressAllocationStoreMarkDerivationFailedSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressAllocationStore(db)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE address_policy_allocations")).
		WithArgs(int64(44), "derivation_failed").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = store.MarkDerivationFailed(context.Background(), entities.PaymentAddressAllocation{
		PaymentAddressID:        44,
		DerivationFailureReason: valueobjects.PaymentAddressAllocationDerivationFailureReasonDerivationFailed,
	})
	if err != nil {
		t.Fatalf("MarkDerivationFailed returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPaymentAddressAllocationStoreMarkDerivationFailedNotReserved(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressAllocationStore(db)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE address_policy_allocations")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = store.MarkDerivationFailed(context.Background(), entities.PaymentAddressAllocation{PaymentAddressID: 44})
	if !errors.Is(err, outport.ErrPaymentAddressAllocationNotReserved) {
		t.Fatalf("expected ErrPaymentAddressAllocationNotReserved, got %v", err)
	}
}

func TestPaymentAddressAllocationStoreReopenFailedReservationNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressAllocationStore(db)
	input := newReservePaymentAddressAllocationInput(" order-1 ")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, slot_index")).
		WithArgs(
			input.IssuancePolicy.AddressPolicyID,
			input.IssuancePolicy.IssuanceConfig.AddressSpaceRef,
		).
		WillReturnError(sql.ErrNoRows)

	allocation, reopened, err := store.ReopenFailedReservation(context.Background(), input)
	if err != nil {
		t.Fatalf("ReopenFailedReservation returned error: %v", err)
	}
	if reopened {
		t.Fatal("expected reopened=false")
	}
	if allocation != (entities.PaymentAddressAllocation{}) {
		t.Fatalf("unexpected allocation: %+v", allocation)
	}
}

func TestPaymentAddressAllocationStoreReopenFailedReservationSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressAllocationStore(db)
	input := newReservePaymentAddressAllocationInput(" order-1 ")

	rows := sqlmock.NewRows([]string{"id", "slot_index"}).AddRow(int64(99), int64(11))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, slot_index")).
		WithArgs(
			input.IssuancePolicy.AddressPolicyID,
			input.IssuancePolicy.IssuanceConfig.AddressSpaceRef,
		).
		WillReturnRows(rows)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE address_policy_allocations")).
		WithArgs(int64(99), int64(125000), "order-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	allocation, reopened, err := store.ReopenFailedReservation(context.Background(), input)
	if err != nil {
		t.Fatalf("ReopenFailedReservation returned error: %v", err)
	}
	if !reopened {
		t.Fatal("expected reopened=true")
	}
	if allocation.SlotIndex != 11 {
		t.Fatalf("unexpected slot index: got %d", allocation.SlotIndex)
	}
}

func TestPaymentAddressAllocationStoreReserveFreshSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressAllocationStore(db)
	input := newReservePaymentAddressAllocationInput(" order-2 ")

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO address_policy_cursors")).
		WithArgs(
			input.IssuancePolicy.AddressPolicyID,
			input.IssuancePolicy.IssuanceConfig.AddressSpaceRef,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT next_index")).
		WithArgs(
			input.IssuancePolicy.AddressPolicyID,
			input.IssuancePolicy.IssuanceConfig.AddressSpaceRef,
		).
		WillReturnRows(sqlmock.NewRows([]string{"next_index"}).AddRow(int64(21)))

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO address_policy_allocations")).
		WithArgs(
			input.IssuancePolicy.AddressPolicyID,
			input.IssuancePolicy.IssuanceConfig.AddressSpaceRef,
			int64(21),
			int64(125000),
			"order-2",
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(144)))

	mock.ExpectExec(regexp.QuoteMeta("UPDATE address_policy_cursors")).
		WithArgs(
			input.IssuancePolicy.AddressPolicyID,
			input.IssuancePolicy.IssuanceConfig.AddressSpaceRef,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	allocation, err := store.ReserveFresh(context.Background(), input)
	if err != nil {
		t.Fatalf("ReserveFresh returned error: %v", err)
	}
	if allocation.PaymentAddressID != 144 {
		t.Fatalf("unexpected payment address id: got %d", allocation.PaymentAddressID)
	}
	if allocation.SlotIndex != 21 {
		t.Fatalf("unexpected slot index: got %d", allocation.SlotIndex)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPaymentAddressAllocationStoreReserveFreshRejectsOverflowIndex(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressAllocationStore(db)
	input := newReservePaymentAddressAllocationInput("order-2")

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO address_policy_cursors")).
		WithArgs(
			input.IssuancePolicy.AddressPolicyID,
			input.IssuancePolicy.IssuanceConfig.AddressSpaceRef,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT next_index")).
		WithArgs(
			input.IssuancePolicy.AddressPolicyID,
			input.IssuancePolicy.IssuanceConfig.AddressSpaceRef,
		).
		WillReturnRows(sqlmock.NewRows([]string{"next_index"}).AddRow(maxSlotIndex + 1))

	_, err = store.ReserveFresh(context.Background(), input)
	if !errors.Is(err, outport.ErrAddressIndexExhausted) {
		t.Fatalf("expected ErrAddressIndexExhausted, got %v", err)
	}
}
