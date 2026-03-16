package postgres

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

func TestEVMPaymentVaultStoreFindSweepCandidatesSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewEVMPaymentVaultStore(db)
	issuedAt := time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"payment_address_id",
		"network",
		"factory_id",
		"factory_address",
		"collector_address",
		"asset_code",
		"asset_type",
		"token_address",
		"salt_hex",
		"predicted_address",
		"deploy_status",
		"sweep_status",
		"issued_at",
	}).AddRow(
		int64(101),
		"sepolia",
		int64(7),
		"0xfactory",
		"0xcollector",
		"usdt",
		"erc20",
		"0xtoken",
		"0xsalt",
		"0xvault",
		"predicted",
		"pending",
		issuedAt,
	)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WillReturnRows(rows)

	records, err := store.FindSweepCandidates(context.Background(), outport.FindEVMSweepCandidatesInput{
		Network:   valueobjects.NetworkID("sepolia"),
		AssetCode: "usdt",
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("FindSweepCandidates returned error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("unexpected record count: got %d", len(records))
	}
	if records[0].PaymentAddressID != 101 {
		t.Fatalf("unexpected payment address id: got %d", records[0].PaymentAddressID)
	}
	if records[0].AssetType != "erc20" {
		t.Fatalf("unexpected asset type: got %q", records[0].AssetType)
	}
}

func TestEVMPaymentVaultStoreMarkSweepSubmittedSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewEVMPaymentVaultStore(db)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE evm_payment_vaults")).
		WithArgs(sqlmock.AnyArg(), "0xtx").
		WillReturnResult(sqlmock.NewResult(0, 2))

	if err := store.MarkSweepSubmitted(context.Background(), outport.MarkEVMSweepSubmittedInput{
		PaymentAddressIDs: []int64{1, 2},
		TxHash:            "0xtx",
	}); err != nil {
		t.Fatalf("MarkSweepSubmitted returned error: %v", err)
	}
}

func TestEVMPaymentVaultStoreMarkSweepFailedRequiresError(t *testing.T) {
	store := NewEVMPaymentVaultStore(&stubNotificationExecutor{})

	err := store.MarkSweepFailed(context.Background(), outport.MarkEVMSweepResultInput{
		PaymentAddressIDs: []int64{1},
		TxHash:            "0xtx",
	})
	if err == nil || err.Error() != "last error is required" {
		t.Fatalf("unexpected error: got %v", err)
	}
}
