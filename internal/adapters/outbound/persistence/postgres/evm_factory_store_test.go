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

func TestEVMFactoryStoreReplaceActiveValidation(t *testing.T) {
	store := NewEVMFactoryStore(nil)

	_, err := store.ReplaceActive(context.Background(), outport.ReplaceActiveEVMFactoryInput{}, time.Now().UTC())
	if err == nil || err.Error() != "executor is not configured" {
		t.Fatalf("unexpected error: got %v", err)
	}
}

func TestEVMFactoryStoreReplaceActiveSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewEVMFactoryStore(db)
	now := time.Date(2026, 3, 16, 2, 30, 0, 0, time.UTC)
	deployedAt := time.Date(2026, 3, 16, 2, 0, 0, 0, time.UTC)
	input := outport.ReplaceActiveEVMFactoryInput{
		Network:               valueobjects.NetworkID("sepolia"),
		FactoryAddress:        "0x1111111111111111111111111111111111111111",
		CollectorAddress:      "0x2222222222222222222222222222222222222222",
		VaultCreationCodeHash: "0x1234",
		DeploymentTxHash:      "0xabc",
		DeployedAt:            deployedAt,
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE evm_factories")).
		WithArgs("sepolia", input.FactoryAddress, now).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rows := sqlmock.NewRows([]string{
		"id",
		"network",
		"factory_address",
		"collector_address",
		"vault_creation_code_hash",
		"status",
		"deployment_tx_hash",
		"deployed_at",
		"created_at",
		"updated_at",
	}).AddRow(
		int64(9),
		"sepolia",
		input.FactoryAddress,
		input.CollectorAddress,
		input.VaultCreationCodeHash,
		"active",
		input.DeploymentTxHash,
		deployedAt,
		now,
		now,
	)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO evm_factories")).
		WithArgs(
			"sepolia",
			input.FactoryAddress,
			input.CollectorAddress,
			input.VaultCreationCodeHash,
			sql.NullString{String: input.DeploymentTxHash, Valid: true},
			sql.NullTime{Time: deployedAt, Valid: true},
			now,
		).
		WillReturnRows(rows)

	record, err := store.ReplaceActive(context.Background(), input, now)
	if err != nil {
		t.Fatalf("ReplaceActive returned error: %v", err)
	}
	if record.ID != 9 {
		t.Fatalf("unexpected id: got %d", record.ID)
	}
	if record.Status != outport.EVMFactoryStatusActive {
		t.Fatalf("unexpected status: got %q", record.Status)
	}
	if record.Network != valueobjects.NetworkID("sepolia") {
		t.Fatalf("unexpected network: got %q", record.Network)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestEVMFactoryStoreListActiveSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewEVMFactoryStore(db)
	now := time.Date(2026, 3, 16, 3, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"id",
		"network",
		"factory_address",
		"collector_address",
		"vault_creation_code_hash",
		"status",
		"deployment_tx_hash",
		"deployed_at",
		"created_at",
		"updated_at",
	}).AddRow(
		int64(1),
		"mainnet",
		"0x1111111111111111111111111111111111111111",
		"0x2222222222222222222222222222222222222222",
		"0x1234",
		"active",
		"",
		nil,
		now,
		now,
	)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WillReturnRows(rows)

	records, err := store.ListActive(context.Background())
	if err != nil {
		t.Fatalf("ListActive returned error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("unexpected record count: got %d", len(records))
	}
	if records[0].Network != valueobjects.NetworkID("mainnet") {
		t.Fatalf("unexpected network: got %q", records[0].Network)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestEVMFactoryStoreFindActiveByNetworkNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewEVMFactoryStore(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs("sepolia").
		WillReturnError(sql.ErrNoRows)

	_, found, err := store.FindActiveByNetwork(context.Background(), valueobjects.NetworkID("sepolia"))
	if err != nil {
		t.Fatalf("FindActiveByNetwork returned error: %v", err)
	}
	if found {
		t.Fatal("expected found=false")
	}
}

func TestEVMFactoryStoreFindActiveByNetworkValidation(t *testing.T) {
	store := NewEVMFactoryStore(&stubEVMFactoryExecutor{})

	_, found, err := store.FindActiveByNetwork(context.Background(), valueobjects.NetworkID("not valid!"))
	if err != nil {
		t.Fatalf("FindActiveByNetwork returned error: %v", err)
	}
	if found {
		t.Fatal("expected found=false")
	}
}

type stubEVMFactoryExecutor struct{}

func (stubEVMFactoryExecutor) ExecContext(context.Context, string, ...any) (sql.Result, error) {
	return nil, errors.New("unexpected exec")
}

func (stubEVMFactoryExecutor) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	return nil, errors.New("unexpected query")
}

func (stubEVMFactoryExecutor) QueryRowContext(context.Context, string, ...any) *sql.Row {
	return &sql.Row{}
}
