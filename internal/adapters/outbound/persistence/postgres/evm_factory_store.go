package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

type EVMFactoryStore struct {
	executor Executor
}

func NewEVMFactoryStore(executor Executor) *EVMFactoryStore {
	return &EVMFactoryStore{executor: executor}
}

func (s *EVMFactoryStore) ReplaceActive(
	ctx context.Context,
	input outport.ReplaceActiveEVMFactoryInput,
	now time.Time,
) (outport.EVMFactoryRecord, error) {
	if s.executor == nil {
		return outport.EVMFactoryRecord{}, errors.New("executor is not configured")
	}
	normalized, err := input.Validate()
	if err != nil {
		return outport.EVMFactoryRecord{}, err
	}
	if now.IsZero() {
		return outport.EVMFactoryRecord{}, errors.New("now is required")
	}
	now = now.UTC()

	if _, err := s.executor.ExecContext(
		ctx,
		`UPDATE evm_factories
		    SET status = 'retired',
		        updated_at = $3
		  WHERE network = $1
		    AND status = 'active'
		    AND factory_address <> $2`,
		string(normalized.Network),
		normalized.FactoryAddress,
		now,
	); err != nil {
		return outport.EVMFactoryRecord{}, err
	}

	var record outport.EVMFactoryRecord
	var deploymentTxHash sql.NullString
	var deployedAt sql.NullTime
	if err := s.executor.QueryRowContext(
		ctx,
		`INSERT INTO evm_factories (
		     network,
		     factory_address,
		     collector_address,
		     vault_creation_code_hash,
		     status,
		     deployment_tx_hash,
		     deployed_at,
		     created_at,
		     updated_at
		   )
		   VALUES ($1, $2, $3, $4, 'active', $5, $6, $7, $7)
		   ON CONFLICT (factory_address)
		   DO UPDATE SET
		     network = EXCLUDED.network,
		     collector_address = EXCLUDED.collector_address,
		     vault_creation_code_hash = EXCLUDED.vault_creation_code_hash,
		     status = 'active',
		     deployment_tx_hash = EXCLUDED.deployment_tx_hash,
		     deployed_at = EXCLUDED.deployed_at,
		     updated_at = EXCLUDED.updated_at
		   RETURNING
		     id,
		     network,
		     factory_address,
		     collector_address,
		     vault_creation_code_hash,
		     status,
		     deployment_tx_hash,
		     deployed_at,
		     created_at,
		     updated_at`,
		string(normalized.Network),
		normalized.FactoryAddress,
		normalized.CollectorAddress,
		normalized.VaultCreationCodeHash,
		nullString(normalized.DeploymentTxHash),
		nullTime(normalized.DeployedAt),
		now,
	).Scan(
		&record.ID,
		&record.Network,
		&record.FactoryAddress,
		&record.CollectorAddress,
		&record.VaultCreationCodeHash,
		&record.Status,
		&deploymentTxHash,
		&deployedAt,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return outport.EVMFactoryRecord{}, err
	}
	record.DeploymentTxHash = deploymentTxHash.String
	record.DeployedAt = deployedAt.Time.UTC()
	record.CreatedAt = record.CreatedAt.UTC()
	record.UpdatedAt = record.UpdatedAt.UTC()

	return record, nil
}

func (s *EVMFactoryStore) ListActive(ctx context.Context) ([]outport.EVMFactoryRecord, error) {
	if s.executor == nil {
		return nil, errors.New("executor is not configured")
	}

	rows, err := s.executor.QueryContext(
		ctx,
		`SELECT
		     id,
		     network,
		     factory_address,
		     collector_address,
		     vault_creation_code_hash,
		     status,
		     deployment_tx_hash,
		     deployed_at,
		     created_at,
		     updated_at
		   FROM evm_factories
		  WHERE status = 'active'
		  ORDER BY network ASC, id ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]outport.EVMFactoryRecord, 0)
	for rows.Next() {
		record, err := scanEVMFactoryRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (s *EVMFactoryStore) FindActiveByNetwork(
	ctx context.Context,
	network valueobjects.NetworkID,
) (outport.EVMFactoryRecord, bool, error) {
	if s.executor == nil {
		return outport.EVMFactoryRecord{}, false, errors.New("executor is not configured")
	}
	normalizedNetwork, ok := valueobjects.ParseNetworkID(string(network))
	if !ok {
		return outport.EVMFactoryRecord{}, false, nil
	}

	row := s.executor.QueryRowContext(
		ctx,
		`SELECT
		     id,
		     network,
		     factory_address,
		     collector_address,
		     vault_creation_code_hash,
		     status,
		     deployment_tx_hash,
		     deployed_at,
		     created_at,
		     updated_at
		   FROM evm_factories
		  WHERE network = $1
		    AND status = 'active'
		  ORDER BY id DESC
		  LIMIT 1`,
		string(normalizedNetwork),
	)

	record, err := scanEVMFactoryRecord(row)
	if errors.Is(err, sql.ErrNoRows) {
		return outport.EVMFactoryRecord{}, false, nil
	}
	if err != nil {
		return outport.EVMFactoryRecord{}, false, err
	}
	return record, true, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanEVMFactoryRecord(scanner rowScanner) (outport.EVMFactoryRecord, error) {
	var record outport.EVMFactoryRecord
	var deploymentTxHash sql.NullString
	var deployedAt sql.NullTime
	if err := scanner.Scan(
		&record.ID,
		&record.Network,
		&record.FactoryAddress,
		&record.CollectorAddress,
		&record.VaultCreationCodeHash,
		&record.Status,
		&deploymentTxHash,
		&deployedAt,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return outport.EVMFactoryRecord{}, err
	}
	record.DeploymentTxHash = deploymentTxHash.String
	if deployedAt.Valid {
		record.DeployedAt = deployedAt.Time.UTC()
	}
	record.CreatedAt = record.CreatedAt.UTC()
	record.UpdatedAt = record.UpdatedAt.UTC()
	return record, nil
}

func nullString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func nullTime(value time.Time) sql.NullTime {
	if value.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: value.UTC(), Valid: true}
}
