package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/value_objects"
)

type PaymentReceiptTrackingRepository struct {
	executor Executor
}

func NewPaymentReceiptTrackingRepository(executor Executor) *PaymentReceiptTrackingRepository {
	return &PaymentReceiptTrackingRepository{executor: executor}
}

func (r *PaymentReceiptTrackingRepository) RegisterMissingIssued(
	ctx context.Context,
	now time.Time,
	defaultRequiredConfirmations int32,
	chain string,
	network string,
) (int, error) {
	if defaultRequiredConfirmations <= 0 {
		return 0, errors.New("default required confirmations must be greater than zero")
	}
	if now.IsZero() {
		return 0, errors.New("now is required")
	}
	chainFilter := strings.ToLower(strings.TrimSpace(chain))
	networkFilter := strings.ToLower(strings.TrimSpace(network))

	result, err := r.executor.ExecContext(
		ctx,
		`INSERT INTO payment_receipt_trackings (
		     payment_address_id,
		     address_policy_id,
		     chain,
		     network,
		     address,
		     issued_at,
		     expected_amount_minor,
		     required_confirmations,
		     receipt_status,
		     next_poll_at
		   )
		   SELECT a.id,
		          a.address_policy_id,
		          a.chain,
		          a.network,
		          a.address,
		          a.issued_at,
		          a.expected_amount_minor,
		          $1,
		          'watching',
		          $2
		   FROM address_policy_allocations a
		   WHERE a.allocation_status = 'issued'
		     AND a.network IS NOT NULL
		     AND a.address IS NOT NULL
		     AND ($3 = '' OR a.chain = $3)
		     AND ($4 = '' OR a.network = $4)
		   ON CONFLICT (payment_address_id) DO NOTHING`,
		defaultRequiredConfirmations,
		now,
		chainFilter,
		networkFilter,
	)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(rowsAffected), nil
}

func (r *PaymentReceiptTrackingRepository) ClaimDue(
	ctx context.Context,
	input outport.ClaimPaymentReceiptTrackingsInput,
) ([]entities.PaymentReceiptTracking, error) {
	if input.Now.IsZero() {
		return nil, errors.New("claim now is required")
	}
	if input.ClaimUntil.IsZero() {
		return nil, errors.New("claim until is required")
	}
	if input.Limit <= 0 {
		return nil, errors.New("claim limit must be greater than zero")
	}
	chainFilter := strings.ToLower(strings.TrimSpace(input.Chain))
	networkFilter := strings.ToLower(strings.TrimSpace(input.Network))

	rows, err := r.executor.QueryContext(
		ctx,
		`WITH due AS (
		     SELECT id
		     FROM payment_receipt_trackings
		     WHERE next_poll_at <= $1
		       AND receipt_status IN ('watching', 'partially_paid', 'paid_unconfirmed', 'double_spend_suspected')
		       AND ($4 = '' OR chain = $4)
		       AND ($5 = '' OR network = $5)
		     ORDER BY next_poll_at ASC, id ASC
		     FOR UPDATE SKIP LOCKED
		     LIMIT $2
		   )
		   UPDATE payment_receipt_trackings pr
		   SET next_poll_at = $3,
		       updated_at = NOW()
		   FROM due
		   WHERE pr.id = due.id
		   RETURNING
		     pr.id,
		     pr.payment_address_id,
		     pr.address_policy_id,
		     pr.chain,
		     pr.network,
		     pr.address,
		     pr.issued_at,
		     pr.expected_amount_minor,
		     pr.required_confirmations,
		     pr.receipt_status,
		     pr.observed_total_minor,
		     pr.confirmed_total_minor,
		     pr.unconfirmed_total_minor,
		     pr.conflict_total_minor,
		     pr.last_observed_block_height,
		     pr.first_observed_at,
		     pr.paid_at,
		     pr.confirmed_at,
		     COALESCE(pr.last_error, '')`,
		input.Now,
		input.Limit,
		input.ClaimUntil,
		chainFilter,
		networkFilter,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	trackings := make([]entities.PaymentReceiptTracking, 0, input.Limit)
	for rows.Next() {
		tracking, err := scanPaymentReceiptTracking(rows)
		if err != nil {
			return nil, err
		}
		trackings = append(trackings, tracking)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return trackings, nil
}

func (r *PaymentReceiptTrackingRepository) SaveObservation(
	ctx context.Context,
	tracking entities.PaymentReceiptTracking,
	now time.Time,
	nextPollAt time.Time,
) error {
	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE payment_receipt_trackings
		 SET receipt_status = $2,
		     observed_total_minor = $3,
		     confirmed_total_minor = $4,
		     unconfirmed_total_minor = $5,
		     conflict_total_minor = $6,
		     last_observed_block_height = $7,
		     first_observed_at = $8,
		     paid_at = $9,
		     confirmed_at = $10,
		     last_error = NULL,
		     last_polled_at = $11,
		     next_poll_at = $12,
		     updated_at = NOW()
		 WHERE payment_address_id = $1`,
		tracking.PaymentAddressID,
		string(tracking.Status),
		tracking.ObservedTotalMinor,
		tracking.ConfirmedTotalMinor,
		tracking.UnconfirmedTotalMinor,
		tracking.ConflictTotalMinor,
		tracking.LastObservedBlockHeight,
		nullableTimePointer(tracking.FirstObservedAt),
		nullableTimePointer(tracking.PaidAt),
		nullableTimePointer(tracking.ConfirmedAt),
		now,
		nextPollAt,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("payment receipt tracking is not found")
	}

	return nil
}

func (r *PaymentReceiptTrackingRepository) SavePollingError(
	ctx context.Context,
	paymentAddressID int64,
	errorReason string,
	now time.Time,
	nextPollAt time.Time,
) error {
	if paymentAddressID <= 0 {
		return errors.New("payment address id must be greater than zero")
	}

	normalizedError := strings.TrimSpace(errorReason)
	if normalizedError == "" {
		return errors.New("polling error reason is required")
	}

	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE payment_receipt_trackings
		 SET last_error = $2,
		     last_polled_at = $3,
		     next_poll_at = $4,
		     updated_at = NOW()
		 WHERE payment_address_id = $1`,
		paymentAddressID,
		normalizedError,
		now,
		nextPollAt,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("payment receipt tracking is not found")
	}

	return nil
}

func scanPaymentReceiptTracking(scanner interface {
	Scan(dest ...any) error
}) (entities.PaymentReceiptTracking, error) {
	var trackingID int64
	var paymentAddressID int64
	var addressPolicyID string
	var chainRaw string
	var networkRaw string
	var address string
	var issuedAt sql.NullTime
	var expectedAmountMinor int64
	var requiredConfirmations int32
	var receiptStatusRaw string
	var observedTotalMinor int64
	var confirmedTotalMinor int64
	var unconfirmedTotalMinor int64
	var conflictTotalMinor int64
	var lastObservedBlockHeight int64
	var firstObservedAt sql.NullTime
	var paidAt sql.NullTime
	var confirmedAt sql.NullTime
	var lastError string

	if err := scanner.Scan(
		&trackingID,
		&paymentAddressID,
		&addressPolicyID,
		&chainRaw,
		&networkRaw,
		&address,
		&issuedAt,
		&expectedAmountMinor,
		&requiredConfirmations,
		&receiptStatusRaw,
		&observedTotalMinor,
		&confirmedTotalMinor,
		&unconfirmedTotalMinor,
		&conflictTotalMinor,
		&lastObservedBlockHeight,
		&firstObservedAt,
		&paidAt,
		&confirmedAt,
		&lastError,
	); err != nil {
		return entities.PaymentReceiptTracking{}, err
	}

	chain, ok := value_objects.ParseChainID(chainRaw)
	if !ok {
		return entities.PaymentReceiptTracking{}, fmt.Errorf("invalid chain in receipt tracking: %s", chainRaw)
	}
	network, ok := value_objects.ParseNetworkID(networkRaw)
	if !ok {
		return entities.PaymentReceiptTracking{}, fmt.Errorf("invalid network in receipt tracking: %s", networkRaw)
	}
	status, ok := value_objects.ParsePaymentReceiptStatus(receiptStatusRaw)
	if !ok {
		return entities.PaymentReceiptTracking{}, fmt.Errorf("unsupported receipt status: %s", receiptStatusRaw)
	}

	tracking := entities.PaymentReceiptTracking{
		TrackingID:              trackingID,
		PaymentAddressID:        paymentAddressID,
		AddressPolicyID:         addressPolicyID,
		Chain:                   chain,
		Network:                 network,
		Address:                 address,
		ExpectedAmountMinor:     expectedAmountMinor,
		RequiredConfirmations:   requiredConfirmations,
		Status:                  status,
		ObservedTotalMinor:      observedTotalMinor,
		ConfirmedTotalMinor:     confirmedTotalMinor,
		UnconfirmedTotalMinor:   unconfirmedTotalMinor,
		ConflictTotalMinor:      conflictTotalMinor,
		LastObservedBlockHeight: lastObservedBlockHeight,
		LastError:               lastError,
	}
	if issuedAt.Valid {
		tracking.IssuedAt = issuedAt.Time.UTC()
	}
	if firstObservedAt.Valid {
		timeValue := firstObservedAt.Time.UTC()
		tracking.FirstObservedAt = &timeValue
	}
	if paidAt.Valid {
		timeValue := paidAt.Time.UTC()
		tracking.PaidAt = &timeValue
	}
	if confirmedAt.Valid {
		timeValue := confirmedAt.Time.UTC()
		tracking.ConfirmedAt = &timeValue
	}

	return tracking, nil
}

func nullableTimePointer(value *time.Time) any {
	if value == nil {
		return nil
	}
	t := value.UTC()
	return t
}
