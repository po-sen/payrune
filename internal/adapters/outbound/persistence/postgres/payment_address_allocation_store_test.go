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
	"payrune/internal/domain/valueobjects"
)

func newAllocationStoreTestPolicy() entities.AddressIssuancePolicy {
	return entities.AddressIssuancePolicy{
		AddressPolicy: entities.AddressPolicy{
			AddressPolicyID: "bitcoin-mainnet-native-segwit",
			Chain:           valueobjects.SupportedChainBitcoin,
			Network:         valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:          string(valueobjects.BitcoinAddressSchemeNativeSegwit),
		},
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

	err := store.Complete(context.Background(), entities.PaymentAddressAllocation{}, time.Time{})
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
		PaymentAddressID:  44,
		Chain:             valueobjects.SupportedChainBitcoin,
		Network:           valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
		Scheme:            string(valueobjects.BitcoinAddressSchemeNativeSegwit),
		Address:           " bc1qallocated ",
		SweepMaterialJSON: ` {"material_type":"bitcoin_hd"} `,
		IssuanceRefKind:   valueobjects.IssuanceRefKindHDPathAbsolute,
		IssuanceRef:       " m/84'/0'/0'/0/11 ",
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE address_policy_allocations")).
		WithArgs(
			int64(44),
			"bitcoin",
			"mainnet",
			"nativeSegwit",
			"bc1qallocated",
			`{"material_type":"bitcoin_hd"}`,
			"hd_path_absolute",
			"m/84'/0'/0'/0/11",
			issuedAt.UTC(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := store.Complete(context.Background(), allocation, issuedAt); err != nil {
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

	err = store.Complete(context.Background(), entities.PaymentAddressAllocation{PaymentAddressID: 44}, issuedAt)
	if !errors.Is(err, outport.ErrPaymentAddressAllocationNotReserved) {
		t.Fatalf("expected ErrPaymentAddressAllocationNotReserved, got %v", err)
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
	input := newFindIssuedPaymentAddressAllocationByIDInput(199)

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
		"sweep_material_json",
		"issuance_ref_kind",
		"issuance_ref",
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
		`{"material_type":"bitcoin_hd"}`,
		"hd_path_absolute",
		"m/84'/0'/0'/0/21",
		"",
	)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,")).
		WithArgs(int64(199)).
		WillReturnRows(rows)

	allocation, found, err := store.FindIssuedByID(context.Background(), input)
	if err != nil {
		t.Fatalf("FindIssuedByID returned error: %v", err)
	}
	if !found {
		t.Fatal("expected found=true")
	}
	if allocation.PaymentAddressID != 199 {
		t.Fatalf("unexpected payment address id: got %d", allocation.PaymentAddressID)
	}
	if allocation.ExpectedAmountMinor != 125000 {
		t.Fatalf("unexpected expected amount minor: got %d", allocation.ExpectedAmountMinor)
	}
	if allocation.Address != "bc1qlookup" {
		t.Fatalf("unexpected address: got %q", allocation.Address)
	}
	if allocation.SweepMaterialJSON != `{"material_type":"bitcoin_hd"}` {
		t.Fatalf("unexpected sweep material: got %q", allocation.SweepMaterialJSON)
	}
	if allocation.IssuanceRefKind != valueobjects.IssuanceRefKindHDPathAbsolute {
		t.Fatalf("unexpected issuance ref kind: got %q", allocation.IssuanceRefKind)
	}
	if allocation.Status != valueobjects.PaymentAddressAllocationStatusIssued {
		t.Fatalf("unexpected status: got %q", allocation.Status)
	}
	if !allocation.DerivationFailureReason.IsZero() {
		t.Fatalf("unexpected derivation failure reason: got %q", allocation.DerivationFailureReason)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPaymentAddressAllocationStoreFindIssuedByIDParsesLegacyFailureReason(t *testing.T) {
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
		"sweep_material_json",
		"issuance_ref_kind",
		"issuance_ref",
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
		`{"material_type":"bitcoin_hd"}`,
		"hd_path_absolute",
		"m/84'/0'/0'/0/21",
		"xpub parse exploded",
	)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,")).
		WithArgs(int64(199)).
		WillReturnRows(rows)

	allocation, found, err := store.FindIssuedByID(context.Background(), newFindIssuedPaymentAddressAllocationByIDInput(199))
	if err != nil {
		t.Fatalf("FindIssuedByID returned error: %v", err)
	}
	if !found {
		t.Fatal("expected found=true")
	}
	if allocation.DerivationFailureReason != valueobjects.PaymentAddressAllocationDerivationFailureReasonDerivationFailed {
		t.Fatalf("unexpected derivation failure reason: got %q", allocation.DerivationFailureReason)
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
		"sweep_material_json",
		"issuance_ref_kind",
		"issuance_ref",
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
		`{"material_type":"bitcoin_hd"}`,
		"hd_path_absolute",
		"m/84'/0'/0'/0/21",
		"",
	)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,")).
		WithArgs(int64(199)).
		WillReturnRows(rows)

	_, _, err = store.FindIssuedByID(context.Background(), newFindIssuedPaymentAddressAllocationByIDInput(199))
	if !errors.Is(err, outport.ErrPaymentAddressAllocationPersistedChainInvalid) {
		t.Fatalf("unexpected invalid chain error: %v", err)
	}
}

func TestPaymentAddressAllocationStoreFindIssuedByIDRejectsInvalidPersistedNetwork(t *testing.T) {
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
		"sweep_material_json",
		"issuance_ref_kind",
		"issuance_ref",
		"failure_reason",
	}).AddRow(
		int64(199),
		"bitcoin-mainnet-native-segwit",
		int64(21),
		int64(125000),
		"order-lookup",
		"bitcoin",
		"main/net",
		"nativeSegwit",
		"bc1qlookup",
		`{"material_type":"bitcoin_hd"}`,
		"hd_path_absolute",
		"m/84'/0'/0'/0/21",
		"",
	)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,")).
		WithArgs(int64(199)).
		WillReturnRows(rows)

	_, _, err = store.FindIssuedByID(context.Background(), newFindIssuedPaymentAddressAllocationByIDInput(199))
	if !errors.Is(err, outport.ErrPaymentAddressAllocationPersistedNetworkInvalid) {
		t.Fatalf("unexpected invalid network error: %v", err)
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
			input.IssuancePolicy.AddressPolicy.AddressPolicyID,
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

func TestPaymentAddressAllocationStoreReopenFailedReservationRejectsOverflowIndex(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewPaymentAddressAllocationStore(db)
	input := newReservePaymentAddressAllocationInput("order-1")

	rows := sqlmock.NewRows([]string{"id", "slot_index"}).
		AddRow(int64(99), maxSlotIndex+1)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, slot_index")).
		WithArgs(
			input.IssuancePolicy.AddressPolicy.AddressPolicyID,
			input.IssuancePolicy.IssuanceConfig.AddressSpaceRef,
		).
		WillReturnRows(rows)

	_, reopened, err := store.ReopenFailedReservation(context.Background(), input)
	if !errors.Is(err, outport.ErrAddressIndexExhausted) {
		t.Fatalf("expected ErrAddressIndexExhausted, got %v", err)
	}
	if reopened {
		t.Fatal("expected reopened=false")
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

	rows := sqlmock.NewRows([]string{"id", "slot_index"}).
		AddRow(int64(99), int64(11))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, slot_index")).
		WithArgs(
			input.IssuancePolicy.AddressPolicy.AddressPolicyID,
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
	if allocation.PaymentAddressID != 99 {
		t.Fatalf("unexpected payment address id: got %d", allocation.PaymentAddressID)
	}
	if allocation.SlotIndex != 11 {
		t.Fatalf("unexpected derivation index: got %d", allocation.SlotIndex)
	}
	if allocation.CustomerReference != "order-1" {
		t.Fatalf("unexpected customer reference: got %q", allocation.CustomerReference)
	}
	if allocation.Status != valueobjects.PaymentAddressAllocationStatusReserved {
		t.Fatalf("unexpected status: got %q", allocation.Status)
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
			input.IssuancePolicy.AddressPolicy.AddressPolicyID,
			input.IssuancePolicy.IssuanceConfig.AddressSpaceRef,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT next_index")).
		WithArgs(
			input.IssuancePolicy.AddressPolicy.AddressPolicyID,
			input.IssuancePolicy.IssuanceConfig.AddressSpaceRef,
		).
		WillReturnRows(sqlmock.NewRows([]string{"next_index"}).AddRow(int64(21)))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO address_policy_allocations (
			   address_policy_id,
			   address_space_ref,
			   slot_index,
			   expected_amount_minor,
			   customer_reference,
			   allocation_status
			 )
		 VALUES ($1, $2, $3, $4, $5, 'reserved')
		 RETURNING id`)).
		WithArgs(
			input.IssuancePolicy.AddressPolicy.AddressPolicyID,
			input.IssuancePolicy.IssuanceConfig.AddressSpaceRef,
			int64(21),
			int64(125000),
			"order-2",
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(144)))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE address_policy_cursors")).
		WithArgs(
			input.IssuancePolicy.AddressPolicy.AddressPolicyID,
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
		t.Fatalf("unexpected derivation index: got %d", allocation.SlotIndex)
	}
	if allocation.CustomerReference != "order-2" {
		t.Fatalf("unexpected customer reference: got %q", allocation.CustomerReference)
	}
	if allocation.Status != valueobjects.PaymentAddressAllocationStatusReserved {
		t.Fatalf("unexpected status: got %q", allocation.Status)
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
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT next_index")).
		WillReturnRows(sqlmock.NewRows([]string{"next_index"}).AddRow(maxSlotIndex + 1))

	_, err = store.ReserveFresh(context.Background(), input)
	if !errors.Is(err, outport.ErrAddressIndexExhausted) {
		t.Fatalf("expected ErrAddressIndexExhausted, got %v", err)
	}
}
