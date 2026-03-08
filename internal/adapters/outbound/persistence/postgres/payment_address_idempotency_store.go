package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"

	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/value_objects"
)

const paymentAddressIdempotencyPrimaryKey = "pk_payment_address_idempotency_keys"

var errPaymentAddressIdempotencyClaimNotFound = errors.New("payment address idempotency claim was not found")

type PaymentAddressIdempotencyStore struct {
	executor Executor
}

func NewPaymentAddressIdempotencyStore(executor Executor) *PaymentAddressIdempotencyStore {
	return &PaymentAddressIdempotencyStore{executor: executor}
}

func (r *PaymentAddressIdempotencyStore) FindByKey(
	ctx context.Context,
	input outport.FindPaymentAddressIdempotencyInput,
) (outport.PaymentAddressIdempotencyRecord, bool, error) {
	idempotencyKey := strings.TrimSpace(input.IdempotencyKey)
	if idempotencyKey == "" {
		return outport.PaymentAddressIdempotencyRecord{}, false, nil
	}

	var (
		rawChain            string
		addressPolicyID     string
		expectedAmountMinor int64
		customerReference   string
		paymentAddressID    sql.NullInt64
	)

	err := r.executor.QueryRowContext(
		ctx,
		`SELECT chain,
		        address_policy_id,
		        expected_amount_minor,
		        COALESCE(customer_reference, ''),
		        payment_address_id
		   FROM payment_address_idempotency_keys
		  WHERE chain = $1
		    AND idempotency_key = $2
		  LIMIT 1`,
		string(input.Chain),
		idempotencyKey,
	).Scan(
		&rawChain,
		&addressPolicyID,
		&expectedAmountMinor,
		&customerReference,
		&paymentAddressID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return outport.PaymentAddressIdempotencyRecord{}, false, nil
	}
	if err != nil {
		return outport.PaymentAddressIdempotencyRecord{}, false, err
	}

	chain, ok := value_objects.ParseSupportedChain(rawChain)
	if !ok {
		return outport.PaymentAddressIdempotencyRecord{}, false, errors.New("persisted idempotency chain is invalid")
	}

	return outport.PaymentAddressIdempotencyRecord{
		Chain:               chain,
		IdempotencyKey:      idempotencyKey,
		AddressPolicyID:     strings.TrimSpace(addressPolicyID),
		ExpectedAmountMinor: expectedAmountMinor,
		CustomerReference:   customerReference,
		PaymentAddressID:    paymentAddressID.Int64,
	}, true, nil
}

func (r *PaymentAddressIdempotencyStore) Claim(
	ctx context.Context,
	input outport.ClaimPaymentAddressIdempotencyInput,
) (outport.PaymentAddressIdempotencyRecord, error) {
	idempotencyKey := strings.TrimSpace(input.IdempotencyKey)
	addressPolicyID := strings.TrimSpace(input.AddressPolicyID)
	customerReference := strings.TrimSpace(input.CustomerReference)

	if input.Chain == "" {
		return outport.PaymentAddressIdempotencyRecord{}, errors.New("chain is required")
	}
	if idempotencyKey == "" {
		return outport.PaymentAddressIdempotencyRecord{}, errors.New("idempotency key is required")
	}
	if addressPolicyID == "" {
		return outport.PaymentAddressIdempotencyRecord{}, errors.New("address policy id is required")
	}
	if input.ExpectedAmountMinor <= 0 {
		return outport.PaymentAddressIdempotencyRecord{}, errors.New("expected amount minor must be greater than zero")
	}

	_, err := r.executor.ExecContext(
		ctx,
		`INSERT INTO payment_address_idempotency_keys (
			   chain,
			   idempotency_key,
			   address_policy_id,
			   expected_amount_minor,
			   customer_reference
			 )
		 VALUES ($1, $2, $3, $4, $5)`,
		string(input.Chain),
		idempotencyKey,
		addressPolicyID,
		input.ExpectedAmountMinor,
		nullIfEmpty(customerReference),
	)
	if err != nil {
		if isPaymentAddressIdempotencyDuplicateKey(err) {
			return outport.PaymentAddressIdempotencyRecord{}, outport.ErrPaymentAddressIdempotencyKeyExists
		}
		return outport.PaymentAddressIdempotencyRecord{}, err
	}

	return outport.PaymentAddressIdempotencyRecord{
		Chain:               input.Chain,
		IdempotencyKey:      idempotencyKey,
		AddressPolicyID:     addressPolicyID,
		ExpectedAmountMinor: input.ExpectedAmountMinor,
		CustomerReference:   customerReference,
	}, nil
}

func (r *PaymentAddressIdempotencyStore) Complete(
	ctx context.Context,
	input outport.CompletePaymentAddressIdempotencyInput,
) error {
	idempotencyKey := strings.TrimSpace(input.IdempotencyKey)
	if input.Chain == "" {
		return errors.New("chain is required")
	}
	if idempotencyKey == "" {
		return errors.New("idempotency key is required")
	}
	if input.PaymentAddressID <= 0 {
		return errors.New("payment address id must be greater than zero")
	}

	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE payment_address_idempotency_keys
		 SET payment_address_id = $3,
		     updated_at = NOW()
		 WHERE chain = $1
		   AND idempotency_key = $2
		   AND payment_address_id IS NULL`,
		string(input.Chain),
		idempotencyKey,
		input.PaymentAddressID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errPaymentAddressIdempotencyClaimNotFound
	}
	return nil
}

func (r *PaymentAddressIdempotencyStore) Release(
	ctx context.Context,
	input outport.ReleasePaymentAddressIdempotencyInput,
) error {
	idempotencyKey := strings.TrimSpace(input.IdempotencyKey)
	if input.Chain == "" {
		return errors.New("chain is required")
	}
	if idempotencyKey == "" {
		return errors.New("idempotency key is required")
	}

	result, err := r.executor.ExecContext(
		ctx,
		`DELETE FROM payment_address_idempotency_keys
		  WHERE chain = $1
		    AND idempotency_key = $2
		    AND payment_address_id IS NULL`,
		string(input.Chain),
		idempotencyKey,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errPaymentAddressIdempotencyClaimNotFound
	}
	return nil
}

func isPaymentAddressIdempotencyDuplicateKey(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return false
	}
	return string(pqErr.Code) == "23505" && pqErr.Constraint == paymentAddressIdempotencyPrimaryKey
}
