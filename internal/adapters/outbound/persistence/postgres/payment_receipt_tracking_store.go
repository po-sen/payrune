package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/valueobjects"
)

type PaymentReceiptTrackingStore struct {
	executor executor
}

func NewPaymentReceiptTrackingStore(executor executor) *PaymentReceiptTrackingStore {
	return &PaymentReceiptTrackingStore{executor: executor}
}

func (r *PaymentReceiptTrackingStore) Create(
	ctx context.Context,
	tracking entities.PaymentReceiptTracking,
	nextPollAt time.Time,
) error {
	if nextPollAt.IsZero() {
		return outport.ErrPaymentReceiptTrackingNextPollAtRequired
	}

	result, err := r.executor.ExecContext(
		ctx,
		`INSERT INTO payment_receipt_trackings (
		     payment_address_id,
		     address_policy_id,
		     chain,
		     network,
		     address,
		     asset_reference,
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
		     $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		   )
		   ON CONFLICT (payment_address_id) DO NOTHING`,
		tracking.PaymentAddressID,
		tracking.AddressPolicyID,
		string(tracking.Chain),
		string(tracking.Network),
		tracking.Address,
		nullIfEmpty(strings.TrimSpace(tracking.AssetReference)),
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
		nullIfEmpty(string(tracking.LastFailureReason)),
		nextPollAt.UTC(),
	)
	if err != nil {
		return outport.ErrPaymentReceiptTrackingStoreFailed
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return outport.ErrPaymentReceiptTrackingStoreFailed
	}
	if rowsAffected == 0 {
		return outport.ErrPaymentReceiptTrackingAlreadyExists
	}
	return nil
}

func (r *PaymentReceiptTrackingStore) ClaimDue(
	ctx context.Context,
	input outport.ClaimPaymentReceiptTrackingsInput,
) ([]entities.PaymentReceiptTracking, error) {
	if input.Now.IsZero() {
		return nil, outport.ErrPaymentReceiptTrackingClaimNowRequired
	}
	if input.ClaimUntil.IsZero() {
		return nil, outport.ErrPaymentReceiptTrackingClaimUntilRequired
	}
	if input.Limit <= 0 {
		return nil, outport.ErrPaymentReceiptTrackingClaimLimitInvalid
	}
	if len(input.Statuses) == 0 {
		return nil, outport.ErrPaymentReceiptTrackingClaimStatusesRequired
	}
	chainFilter := strings.ToLower(strings.TrimSpace(input.Chain))
	networkFilter := strings.ToLower(strings.TrimSpace(input.Network))
	statusFilters := make([]string, 0, len(input.Statuses))
	for _, status := range input.Statuses {
		if status == "" {
			return nil, outport.ErrPaymentReceiptTrackingClaimStatusRequired
		}
		statusFilters = append(statusFilters, string(status))
	}

	rows, err := r.executor.QueryContext(
		ctx,
		`WITH due AS (
		     SELECT id
		     FROM payment_receipt_trackings
		     WHERE next_poll_at <= $1
		       AND (lease_until IS NULL OR lease_until <= $1)
		       AND receipt_status = ANY($4)
		       AND ($5 = '' OR chain = $5)
		       AND ($6 = '' OR network = $6)
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
		     COALESCE(pr.last_error, ''),
		     COALESCE(pr.asset_reference, '')`,
		input.Now,
		input.Limit,
		input.ClaimUntil,
		pq.Array(statusFilters),
		chainFilter,
		networkFilter,
	)
	if err != nil {
		return nil, outport.ErrPaymentReceiptTrackingStoreFailed
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
		return nil, outport.ErrPaymentReceiptTrackingStoreFailed
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
		return outport.ErrPaymentReceiptTrackingPolledAtRequired
	}
	if nextPollAt.IsZero() {
		return outport.ErrPaymentReceiptTrackingNextPollAtRequired
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
		nullIfEmpty(string(tracking.LastFailureReason)),
		polledAt.UTC(),
		nextPollAt.UTC(),
	)
	if err != nil {
		return outport.ErrPaymentReceiptTrackingStoreFailed
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return outport.ErrPaymentReceiptTrackingStoreFailed
	}
	if rowsAffected == 0 {
		return outport.ErrPaymentReceiptTrackingNotFound
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
	var rawAssetReference string

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
		&rawAssetReference,
	); err != nil {
		return entities.PaymentReceiptTracking{}, outport.ErrPaymentReceiptTrackingStoreFailed
	}

	chain, ok := valueobjects.ParseChainID(chainRaw)
	if !ok {
		return entities.PaymentReceiptTracking{}, fmt.Errorf("%w: %s", outport.ErrPaymentReceiptTrackingPersistedChainInvalid, chainRaw)
	}
	parsedAddressPolicyID, err := valueobjects.NewAddressPolicyID(addressPolicyID)
	if err != nil {
		return entities.PaymentReceiptTracking{}, fmt.Errorf(
			"%w: %s",
			outport.ErrPaymentReceiptTrackingPersistedAddressPolicyIDInvalid,
			strings.TrimSpace(addressPolicyID),
		)
	}
	network, ok := valueobjects.ParseNetworkID(networkRaw)
	if !ok {
		return entities.PaymentReceiptTracking{}, fmt.Errorf("%w: %s", outport.ErrPaymentReceiptTrackingPersistedNetworkInvalid, networkRaw)
	}
	assetReference := strings.TrimSpace(rawAssetReference)
	status, ok := valueobjects.ParsePaymentReceiptStatus(receiptStatusRaw)
	if !ok {
		return entities.PaymentReceiptTracking{}, fmt.Errorf("%w: %s", outport.ErrPaymentReceiptTrackingPersistedStatusInvalid, receiptStatusRaw)
	}
	lastFailureReason := normalizePaymentReceiptTrackingFailureReason(lastError)

	tracking := entities.PaymentReceiptTracking{
		TrackingID:              trackingID,
		PaymentAddressID:        paymentAddressID,
		AddressPolicyID:         parsedAddressPolicyID,
		Chain:                   chain,
		Network:                 network,
		Address:                 address,
		AssetReference:          assetReference,
		ExpectedAmountMinor:     expectedAmountMinor,
		RequiredConfirmations:   requiredConfirmations,
		Status:                  status,
		ObservedTotalMinor:      observedTotalMinor,
		ConfirmedTotalMinor:     confirmedTotalMinor,
		UnconfirmedTotalMinor:   unconfirmedTotalMinor,
		LastObservedBlockHeight: lastObservedBlockHeight,
		LastFailureReason:       lastFailureReason,
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

func nullableTimePointer(value *time.Time) any {
	if value == nil {
		return nil
	}
	t := value.UTC()
	return t
}
