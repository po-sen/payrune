package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/value_objects"
)

const maxNonHardenedIndex int64 = 0x7fffffff

var errAllocationNotReserved = errors.New("address allocation is not reserved")

type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type PaymentAddressAllocationRepository struct {
	executor Executor
}

func NewPaymentAddressAllocationRepository(executor Executor) *PaymentAddressAllocationRepository {
	return &PaymentAddressAllocationRepository{executor: executor}
}

func (r *PaymentAddressAllocationRepository) Complete(
	ctx context.Context,
	allocation entities.PaymentAddressAllocation,
) error {
	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE address_policy_allocations
		 SET chain = $2,
		     network = $3,
		     scheme = $4,
		     address = $5,
		     derivation_path = $6,
		     allocation_status = 'issued',
		     issued_at = NOW(),
		     failure_reason = NULL
		 WHERE id = $1 AND allocation_status = 'reserved'`,
		allocation.PaymentAddressID,
		string(allocation.Chain),
		string(allocation.Network),
		string(allocation.Scheme),
		strings.TrimSpace(allocation.Address),
		nullIfEmpty(allocation.DerivationPath),
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errAllocationNotReserved
	}

	return nil
}

func (r *PaymentAddressAllocationRepository) MarkDerivationFailed(
	ctx context.Context,
	allocation entities.PaymentAddressAllocation,
) error {
	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE address_policy_allocations
		 SET allocation_status = 'derivation_failed',
		     failure_reason = $2
		 WHERE id = $1 AND allocation_status = 'reserved'`,
		allocation.PaymentAddressID,
		nullIfEmpty(strings.TrimSpace(allocation.FailureReason)),
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errAllocationNotReserved
	}

	return nil
}

func (r *PaymentAddressAllocationRepository) ReopenFailedReservation(
	ctx context.Context,
	input outport.ReservePaymentAddressAllocationInput,
) (entities.PaymentAddressAllocation, bool, error) {
	customerReference := strings.TrimSpace(input.CustomerReference)

	var paymentAddressID int64
	var derivationIndex int64
	err := r.executor.QueryRowContext(
		ctx,
		`SELECT id, derivation_index
		 FROM address_policy_allocations
		 WHERE address_policy_id = $1
		   AND xpub_fingerprint_algo = $2
		   AND xpub_fingerprint = $3
		   AND allocation_status = 'derivation_failed'
		 ORDER BY reserved_at ASC, id ASC
		 LIMIT 1
		 FOR UPDATE SKIP LOCKED`,
		input.Policy.AddressPolicyID,
		input.Policy.XPubFingerprintAlgo,
		input.Policy.XPubFingerprint,
	).Scan(&paymentAddressID, &derivationIndex)
	if errors.Is(err, sql.ErrNoRows) {
		return entities.PaymentAddressAllocation{}, false, nil
	}
	if err != nil {
		return entities.PaymentAddressAllocation{}, false, err
	}
	if derivationIndex < 0 || derivationIndex > maxNonHardenedIndex {
		return entities.PaymentAddressAllocation{}, false, outport.ErrAddressIndexExhausted
	}

	if _, err := r.executor.ExecContext(
		ctx,
		`UPDATE address_policy_allocations
		 SET allocation_status = 'reserved',
		     failure_reason = NULL,
		     expected_amount_minor = $2,
		     customer_reference = $3,
		     chain = NULL,
		     network = NULL,
		     scheme = NULL,
		     address = NULL,
		     derivation_path = NULL,
		     reserved_at = NOW(),
		     issued_at = NULL
		 WHERE id = $1`,
		paymentAddressID,
		input.ExpectedAmountMinor,
		nullIfEmpty(customerReference),
	); err != nil {
		return entities.PaymentAddressAllocation{}, false, err
	}

	return entities.PaymentAddressAllocation{
		PaymentAddressID:    paymentAddressID,
		AddressPolicyID:     input.Policy.AddressPolicyID,
		DerivationIndex:     uint32(derivationIndex),
		ExpectedAmountMinor: input.ExpectedAmountMinor,
		CustomerReference:   customerReference,
		Status:              value_objects.PaymentAddressAllocationStatusReserved,
	}, true, nil
}

func (r *PaymentAddressAllocationRepository) ReserveFresh(
	ctx context.Context,
	input outport.ReservePaymentAddressAllocationInput,
) (entities.PaymentAddressAllocation, error) {
	customerReference := strings.TrimSpace(input.CustomerReference)

	if _, err := r.executor.ExecContext(
		ctx,
		`INSERT INTO address_policy_cursors (
			   address_policy_id,
			   xpub_fingerprint_algo,
			   xpub_fingerprint,
			   next_index
			 )
		 VALUES ($1, $2, $3, 0)
		 ON CONFLICT (address_policy_id, xpub_fingerprint_algo, xpub_fingerprint) DO NOTHING`,
		input.Policy.AddressPolicyID,
		input.Policy.XPubFingerprintAlgo,
		input.Policy.XPubFingerprint,
	); err != nil {
		return entities.PaymentAddressAllocation{}, err
	}

	var nextIndex int64
	err := r.executor.QueryRowContext(
		ctx,
		`SELECT next_index
		 FROM address_policy_cursors
		 WHERE address_policy_id = $1
		   AND xpub_fingerprint_algo = $2
		   AND xpub_fingerprint = $3
		 FOR UPDATE`,
		input.Policy.AddressPolicyID,
		input.Policy.XPubFingerprintAlgo,
		input.Policy.XPubFingerprint,
	).Scan(&nextIndex)
	if err != nil {
		return entities.PaymentAddressAllocation{}, err
	}
	if nextIndex > maxNonHardenedIndex {
		return entities.PaymentAddressAllocation{}, outport.ErrAddressIndexExhausted
	}

	var paymentAddressID int64
	err = r.executor.QueryRowContext(
		ctx,
		`INSERT INTO address_policy_allocations (
			   address_policy_id,
			   xpub_fingerprint_algo,
			   xpub_fingerprint,
			   derivation_index,
			   expected_amount_minor,
			   customer_reference,
			   allocation_status
			 )
		 VALUES ($1, $2, $3, $4, $5, $6, 'reserved')
		 RETURNING id`,
		input.Policy.AddressPolicyID,
		input.Policy.XPubFingerprintAlgo,
		input.Policy.XPubFingerprint,
		nextIndex,
		input.ExpectedAmountMinor,
		nullIfEmpty(customerReference),
	).Scan(&paymentAddressID)
	if err != nil {
		return entities.PaymentAddressAllocation{}, err
	}

	if _, err := r.executor.ExecContext(
		ctx,
		`UPDATE address_policy_cursors
		 SET next_index = next_index + 1,
		     updated_at = NOW()
		 WHERE address_policy_id = $1
		   AND xpub_fingerprint_algo = $2
		   AND xpub_fingerprint = $3`,
		input.Policy.AddressPolicyID,
		input.Policy.XPubFingerprintAlgo,
		input.Policy.XPubFingerprint,
	); err != nil {
		return entities.PaymentAddressAllocation{}, err
	}

	return entities.PaymentAddressAllocation{
		PaymentAddressID:    paymentAddressID,
		AddressPolicyID:     input.Policy.AddressPolicyID,
		DerivationIndex:     uint32(nextIndex),
		ExpectedAmountMinor: input.ExpectedAmountMinor,
		CustomerReference:   customerReference,
		Status:              value_objects.PaymentAddressAllocationStatusReserved,
	}, nil
}

func nullIfEmpty(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}
