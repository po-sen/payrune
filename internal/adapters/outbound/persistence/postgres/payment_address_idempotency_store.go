package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

const paymentAddressIdempotencyPrimaryKey = "pk_payment_address_idempotency_keys"

type PaymentAddressIdempotencyStore struct {
	executor executor
}

func NewPaymentAddressIdempotencyStore(executor executor) *PaymentAddressIdempotencyStore {
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
		return outport.PaymentAddressIdempotencyRecord{}, false, outport.ErrPaymentAddressIdempotencyStoreFailed
	}

	chain, ok := valueobjects.ParseSupportedChain(rawChain)
	if !ok {
		return outport.PaymentAddressIdempotencyRecord{}, false, outport.ErrPaymentAddressIdempotencyPersistedChainInvalid
	}
	parsedAddressPolicyID, err := valueobjects.NewAddressPolicyID(addressPolicyID)
	if err != nil {
		return outport.PaymentAddressIdempotencyRecord{}, false, outport.ErrPaymentAddressIdempotencyPersistedAddressPolicyIDInvalid
	}

	return outport.PaymentAddressIdempotencyRecord{
		Chain:               chain,
		IdempotencyKey:      idempotencyKey,
		AddressPolicyID:     parsedAddressPolicyID,
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
	addressPolicyID := input.AddressPolicyID.Normalize()
	customerReference := strings.TrimSpace(input.CustomerReference)

	if input.Chain == "" {
		return outport.PaymentAddressIdempotencyRecord{}, outport.ErrPaymentAddressIdempotencyChainRequired
	}
	if idempotencyKey == "" {
		return outport.PaymentAddressIdempotencyRecord{}, outport.ErrPaymentAddressIdempotencyKeyRequired
	}
	if addressPolicyID.IsZero() {
		return outport.PaymentAddressIdempotencyRecord{}, outport.ErrPaymentAddressIdempotencyAddressPolicyIDRequired
	}
	if input.ExpectedAmountMinor <= 0 {
		return outport.PaymentAddressIdempotencyRecord{}, outport.ErrPaymentAddressIdempotencyExpectedAmountInvalid
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
		string(addressPolicyID),
		input.ExpectedAmountMinor,
		nullIfEmpty(customerReference),
	)
	if err != nil {
		if isPaymentAddressIdempotencyDuplicateKey(err) {
			return outport.PaymentAddressIdempotencyRecord{}, outport.ErrPaymentAddressIdempotencyKeyExists
		}
		return outport.PaymentAddressIdempotencyRecord{}, outport.ErrPaymentAddressIdempotencyStoreFailed
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
		return outport.ErrPaymentAddressIdempotencyChainRequired
	}
	if idempotencyKey == "" {
		return outport.ErrPaymentAddressIdempotencyKeyRequired
	}
	if input.PaymentAddressID <= 0 {
		return outport.ErrPaymentAddressIdempotencyPaymentAddressIDInvalid
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
		return outport.ErrPaymentAddressIdempotencyStoreFailed
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return outport.ErrPaymentAddressIdempotencyStoreFailed
	}
	if rowsAffected == 0 {
		return outport.ErrPaymentAddressIdempotencyClaimNotFound
	}
	return nil
}

func (r *PaymentAddressIdempotencyStore) Release(
	ctx context.Context,
	input outport.ReleasePaymentAddressIdempotencyInput,
) error {
	idempotencyKey := strings.TrimSpace(input.IdempotencyKey)
	if input.Chain == "" {
		return outport.ErrPaymentAddressIdempotencyChainRequired
	}
	if idempotencyKey == "" {
		return outport.ErrPaymentAddressIdempotencyKeyRequired
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
		return outport.ErrPaymentAddressIdempotencyStoreFailed
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return outport.ErrPaymentAddressIdempotencyStoreFailed
	}
	if rowsAffected == 0 {
		return outport.ErrPaymentAddressIdempotencyClaimNotFound
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
