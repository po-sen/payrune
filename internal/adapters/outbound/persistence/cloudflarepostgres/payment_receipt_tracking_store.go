package cloudflarepostgres

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

type PaymentReceiptTrackingStore struct {
	executor Executor
}

func NewPaymentReceiptTrackingStore(executor Executor) *PaymentReceiptTrackingStore {
	return &PaymentReceiptTrackingStore{executor: executor}
}

func (r *PaymentReceiptTrackingStore) Create(
	ctx context.Context,
	tracking entities.PaymentReceiptTracking,
	nextPollAt time.Time,
) error {
	if nextPollAt.IsZero() {
		return errors.New("next poll at is required")
	}

	result, err := r.executor.ExecContext(
		ctx,
		`INSERT INTO payment_receipt_trackings (
		     payment_address_id,
		     address_policy_id,
		     chain,
		     network,
		     address,
		     issued_at,
		     expires_at,
		     expected_amount_minor,
		     required_confirmations,
		     receipt_status,
		     observed_total_minor,
		     confirmed_total_minor,
		     unconfirmed_total_minor,
		     last_observed_block_height,
		     first_observed_at,
		     paid_at,
		     confirmed_at,
		     last_error,
		     next_poll_at
		   )
		   VALUES (
		     $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
		     $11, $12, $13, $14, $15, $16, $17, $18, $19
		   )
		   ON CONFLICT (payment_address_id) DO NOTHING`,
		tracking.PaymentAddressID,
		tracking.AddressPolicyID,
		string(tracking.Chain),
		string(tracking.Network),
		tracking.Address,
		tracking.IssuedAt.UTC(),
		nullableTimePointer(tracking.ExpiresAt),
		tracking.ExpectedAmountMinor,
		tracking.RequiredConfirmations,
		string(tracking.Status),
		tracking.ObservedTotalMinor,
		tracking.ConfirmedTotalMinor,
		tracking.UnconfirmedTotalMinor,
		tracking.LastObservedBlockHeight,
		nullableTimePointer(tracking.FirstObservedAt),
		nullableTimePointer(tracking.PaidAt),
		nullableTimePointer(tracking.ConfirmedAt),
		nullIfEmpty(tracking.LastError),
		nextPollAt.UTC(),
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("payment receipt tracking already exists")
	}
	return nil
}

func (r *PaymentReceiptTrackingStore) ClaimDue(
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
	if len(input.Statuses) == 0 {
		return nil, errors.New("claim statuses are required")
	}

	chainFilter := strings.ToLower(strings.TrimSpace(input.Chain))
	networkFilter := strings.ToLower(strings.TrimSpace(input.Network))
	statusFilters := make([]string, 0, len(input.Statuses))
	args := []any{input.Now.UTC(), input.Limit, input.ClaimUntil.UTC()}
	for _, status := range input.Statuses {
		if status == "" {
			return nil, errors.New("claim status is required")
		}
		statusFilters = append(statusFilters, string(status))
		args = append(args, string(status))
	}
	args = append(args, chainFilter, networkFilter)

	statusClause := buildSequentialPlaceholders(4, len(statusFilters))
	chainArgIndex := 4 + len(statusFilters)
	networkArgIndex := chainArgIndex + 1

	query := fmt.Sprintf(
		`WITH due AS (
		     SELECT id
		     FROM payment_receipt_trackings
		     WHERE next_poll_at <= $1
		       AND (lease_until IS NULL OR lease_until <= $1)
		       AND receipt_status IN (%s)
		       AND ($%d = '' OR chain = $%d)
		       AND ($%d = '' OR network = $%d)
		     ORDER BY next_poll_at ASC, id ASC
		     FOR UPDATE SKIP LOCKED
		     LIMIT $2
		   )
		   UPDATE payment_receipt_trackings pr
		   SET lease_until = $3,
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
		     pr.last_observed_block_height,
		     pr.first_observed_at,
		     pr.paid_at,
		     pr.confirmed_at,
		     pr.expires_at,
		     COALESCE(pr.last_error, '')`,
		statusClause,
		chainArgIndex,
		chainArgIndex,
		networkArgIndex,
		networkArgIndex,
	)

	rows, err := r.executor.QueryContext(ctx, query, args...)
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

func (r *PaymentReceiptTrackingStore) Save(
	ctx context.Context,
	tracking entities.PaymentReceiptTracking,
	polledAt time.Time,
	nextPollAt time.Time,
) error {
	if polledAt.IsZero() {
		return errors.New("polled at is required")
	}
	if nextPollAt.IsZero() {
		return errors.New("next poll at is required")
	}

	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE payment_receipt_trackings
		 SET receipt_status = $2,
		     observed_total_minor = $3,
		     confirmed_total_minor = $4,
		     unconfirmed_total_minor = $5,
		     last_observed_block_height = $6,
		     first_observed_at = $7,
		     paid_at = $8,
		     confirmed_at = $9,
		     expires_at = $10,
		     last_error = $11,
		     last_polled_at = $12,
		     next_poll_at = $13,
		     lease_until = NULL,
		     updated_at = NOW()
		 WHERE payment_address_id = $1`,
		tracking.PaymentAddressID,
		string(tracking.Status),
		tracking.ObservedTotalMinor,
		tracking.ConfirmedTotalMinor,
		tracking.UnconfirmedTotalMinor,
		tracking.LastObservedBlockHeight,
		nullableTimePointer(tracking.FirstObservedAt),
		nullableTimePointer(tracking.PaidAt),
		nullableTimePointer(tracking.ConfirmedAt),
		nullableTimePointer(tracking.ExpiresAt),
		nullIfEmpty(tracking.LastError),
		polledAt.UTC(),
		nextPollAt.UTC(),
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
	var lastObservedBlockHeight int64
	var firstObservedAt sql.NullTime
	var paidAt sql.NullTime
	var confirmedAt sql.NullTime
	var expiresAt sql.NullTime
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
		&lastObservedBlockHeight,
		&firstObservedAt,
		&paidAt,
		&confirmedAt,
		&expiresAt,
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
	if expiresAt.Valid {
		timeValue := expiresAt.Time.UTC()
		tracking.ExpiresAt = &timeValue
	}

	return tracking, nil
}
