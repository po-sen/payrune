package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"

	"github.com/lib/pq"
)

type EVMPaymentVaultStore struct {
	executor Executor
}

func NewEVMPaymentVaultStore(executor Executor) *EVMPaymentVaultStore {
	return &EVMPaymentVaultStore{executor: executor}
}

func (s *EVMPaymentVaultStore) Create(
	ctx context.Context,
	input outport.CreateEVMPaymentVaultInput,
) (outport.EVMPaymentVaultRecord, error) {
	validated, err := input.Validate()
	if err != nil {
		return outport.EVMPaymentVaultRecord{}, err
	}

	_, err = s.executor.ExecContext(
		ctx,
		`INSERT INTO evm_payment_vaults (
		     payment_address_id,
		     network,
		     factory_id,
		     factory_address,
		     collector_address,
		     token_address,
		     salt_hex,
		     predicted_address,
		     deploy_status,
		     sweep_status
		   )
		   VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'predicted', 'pending')`,
		validated.PaymentAddressID,
		string(validated.Network),
		validated.FactoryID,
		validated.FactoryAddress,
		validated.CollectorAddress,
		nullIfEmpty(validated.TokenAddress),
		validated.SaltHex,
		validated.PredictedAddress,
	)
	if err != nil {
		return outport.EVMPaymentVaultRecord{}, err
	}

	return outport.EVMPaymentVaultRecord{
		PaymentAddressID: validated.PaymentAddressID,
		Network:          validated.Network,
		FactoryID:        validated.FactoryID,
		FactoryAddress:   validated.FactoryAddress,
		CollectorAddress: validated.CollectorAddress,
		TokenAddress:     validated.TokenAddress,
		SaltHex:          validated.SaltHex,
		PredictedAddress: validated.PredictedAddress,
		DeployStatus:     "predicted",
		SweepStatus:      "pending",
	}, nil
}

func (s *EVMPaymentVaultStore) FindSweepCandidates(
	ctx context.Context,
	input outport.FindEVMSweepCandidatesInput,
) ([]outport.EVMSweepCandidateRecord, error) {
	validated, err := input.Validate()
	if err != nil {
		return nil, err
	}

	queryArgs := make([]any, 0, 4)
	conditions := []string{
		"pr.chain = 'ethereum'",
		"pr.receipt_status = 'paid_confirmed'",
		"v.sweep_status IN ('pending', 'failed')",
	}

	if validated.Network != "" {
		queryArgs = append(queryArgs, string(validated.Network))
		conditions = append(conditions, fmt.Sprintf("v.network = $%d", len(queryArgs)))
	}
	if validated.AssetCode != "" {
		queryArgs = append(queryArgs, validated.AssetCode)
		conditions = append(conditions, fmt.Sprintf("pr.asset_code = $%d", len(queryArgs)))
	}
	if len(validated.PaymentAddressIDs) > 0 {
		queryArgs = append(queryArgs, pq.Array(validated.PaymentAddressIDs))
		conditions = append(conditions, fmt.Sprintf("v.payment_address_id = ANY($%d)", len(queryArgs)))
	}
	if !validated.BeforeIssuedAt.IsZero() {
		queryArgs = append(queryArgs, validated.BeforeIssuedAt)
		conditions = append(conditions, fmt.Sprintf("pr.issued_at < $%d", len(queryArgs)))
	}

	limitClause := ""
	if validated.Limit > 0 {
		queryArgs = append(queryArgs, validated.Limit)
		limitClause = fmt.Sprintf(" LIMIT $%d", len(queryArgs))
	}

	rows, err := s.executor.QueryContext(
		ctx,
		`SELECT
		     v.payment_address_id,
		     v.network,
		     v.factory_id,
		     v.factory_address,
		     v.collector_address,
		     pr.asset_code,
		     pr.asset_type,
		     COALESCE(v.token_address, ''),
		     v.salt_hex,
		     v.predicted_address,
		     v.deploy_status,
		     v.sweep_status,
		     pr.issued_at
		   FROM evm_payment_vaults v
		   INNER JOIN payment_receipt_trackings pr
		     ON pr.payment_address_id = v.payment_address_id
		  WHERE `+strings.Join(conditions, " AND ")+`
		  ORDER BY pr.issued_at ASC, v.payment_address_id ASC`+limitClause,
		queryArgs...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]outport.EVMSweepCandidateRecord, 0)
	for rows.Next() {
		record, err := scanEVMSweepCandidateRecord(rows)
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

func (s *EVMPaymentVaultStore) MarkSweepSubmitted(
	ctx context.Context,
	input outport.MarkEVMSweepSubmittedInput,
) error {
	validated, err := input.Validate()
	if err != nil {
		return err
	}

	_, err = s.executor.ExecContext(
		ctx,
		`UPDATE evm_payment_vaults
		    SET sweep_status = 'submitted',
		        last_sweep_tx_hash = $2,
		        last_sweep_error = NULL,
		        updated_at = NOW()
		  WHERE payment_address_id = ANY($1)`,
		pq.Array(validated.PaymentAddressIDs),
		validated.TxHash,
	)
	return err
}

func (s *EVMPaymentVaultStore) MarkSweepSucceeded(
	ctx context.Context,
	input outport.MarkEVMSweepResultInput,
) error {
	validated, err := input.Validate()
	if err != nil {
		return err
	}
	if strings.TrimSpace(validated.TxHash) == "" {
		return errors.New("tx hash is required")
	}

	_, err = s.executor.ExecContext(
		ctx,
		`UPDATE evm_payment_vaults
		    SET deploy_status = 'deployed',
		        sweep_status = 'succeeded',
		        last_sweep_tx_hash = $2,
		        last_sweep_error = NULL,
		        updated_at = NOW()
		  WHERE payment_address_id = ANY($1)`,
		pq.Array(validated.PaymentAddressIDs),
		validated.TxHash,
	)
	return err
}

func (s *EVMPaymentVaultStore) MarkSweepFailed(
	ctx context.Context,
	input outport.MarkEVMSweepResultInput,
) error {
	validated, err := input.Validate()
	if err != nil {
		return err
	}
	if validated.LastError == "" {
		return errors.New("last error is required")
	}

	_, err = s.executor.ExecContext(
		ctx,
		`UPDATE evm_payment_vaults
		    SET last_sweep_tx_hash = NULLIF($2, ''),
		        last_sweep_error = $3,
		        sweep_status = 'failed',
		        updated_at = NOW()
		  WHERE payment_address_id = ANY($1)`,
		pq.Array(validated.PaymentAddressIDs),
		validated.TxHash,
		validated.LastError,
	)
	return err
}

func scanEVMSweepCandidateRecord(scanner interface {
	Scan(dest ...any) error
}) (outport.EVMSweepCandidateRecord, error) {
	var (
		record       outport.EVMSweepCandidateRecord
		rawNetwork   string
		tokenAddress string
		issuedAt     sql.NullTime
	)

	if err := scanner.Scan(
		&record.PaymentAddressID,
		&rawNetwork,
		&record.FactoryID,
		&record.FactoryAddress,
		&record.CollectorAddress,
		&record.AssetCode,
		&record.AssetType,
		&tokenAddress,
		&record.SaltHex,
		&record.PredictedAddress,
		&record.DeployStatus,
		&record.SweepStatus,
		&issuedAt,
	); err != nil {
		return outport.EVMSweepCandidateRecord{}, err
	}

	network, ok := valueobjects.ParseNetworkID(rawNetwork)
	if !ok {
		return outport.EVMSweepCandidateRecord{}, fmt.Errorf("invalid evm sweep candidate network: %s", rawNetwork)
	}
	if !issuedAt.Valid {
		return outport.EVMSweepCandidateRecord{}, errors.New("evm sweep candidate issued at is required")
	}

	record.Network = network
	record.FactoryAddress = strings.TrimSpace(record.FactoryAddress)
	record.CollectorAddress = strings.TrimSpace(record.CollectorAddress)
	record.AssetCode = strings.TrimSpace(record.AssetCode)
	record.AssetType = strings.TrimSpace(record.AssetType)
	record.TokenAddress = strings.TrimSpace(tokenAddress)
	record.SaltHex = strings.TrimSpace(strings.ToLower(record.SaltHex))
	record.PredictedAddress = strings.TrimSpace(record.PredictedAddress)
	record.DeployStatus = strings.TrimSpace(record.DeployStatus)
	record.SweepStatus = strings.TrimSpace(record.SweepStatus)
	record.IssuedAt = issuedAt.Time.UTC()

	return record, nil
}
