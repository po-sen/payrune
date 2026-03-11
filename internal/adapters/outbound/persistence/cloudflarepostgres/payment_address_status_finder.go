package cloudflarepostgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/value_objects"
)

type PaymentAddressStatusFinder struct {
	executor Executor
}

func NewPaymentAddressStatusFinder(executor Executor) *PaymentAddressStatusFinder {
	return &PaymentAddressStatusFinder{executor: executor}
}

func (f *PaymentAddressStatusFinder) FindByID(
	ctx context.Context,
	input outport.FindPaymentAddressStatusInput,
) (outport.PaymentAddressStatusRecord, bool, error) {
	if input.PaymentAddressID <= 0 {
		return outport.PaymentAddressStatusRecord{}, false, nil
	}

	var (
		paymentAddressID        int64
		addressPolicyID         string
		expectedAmountMinor     int64
		customerReference       string
		rawChain                string
		rawNetwork              string
		scheme                  string
		address                 string
		issuedAt                sql.NullTime
		trackingID              sql.NullInt64
		requiredConfirmations   sql.NullInt32
		receiptStatusRaw        sql.NullString
		observedTotalMinor      sql.NullInt64
		confirmedTotalMinor     sql.NullInt64
		unconfirmedTotalMinor   sql.NullInt64
		lastObservedBlockHeight sql.NullInt64
		firstObservedAt         sql.NullTime
		paidAt                  sql.NullTime
		confirmedAt             sql.NullTime
		expiresAt               sql.NullTime
		lastError               sql.NullString
	)

	err := f.executor.QueryRowContext(
		ctx,
		`SELECT
		     a.id,
		     a.address_policy_id,
		     a.expected_amount_minor,
		     COALESCE(a.customer_reference, ''),
		     COALESCE(a.chain, ''),
		     COALESCE(a.network, ''),
		     COALESCE(a.scheme, ''),
		     COALESCE(a.address, ''),
		     a.issued_at,
		     pr.id,
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
		     pr.last_error
		   FROM address_policy_allocations a
		   LEFT JOIN payment_receipt_trackings pr
		     ON pr.payment_address_id = a.id
		  WHERE a.chain = $1
		    AND a.id = $2
		    AND a.allocation_status = 'issued'
		  LIMIT 1`,
		string(input.Chain),
		input.PaymentAddressID,
	).Scan(
		&paymentAddressID,
		&addressPolicyID,
		&expectedAmountMinor,
		&customerReference,
		&rawChain,
		&rawNetwork,
		&scheme,
		&address,
		&issuedAt,
		&trackingID,
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
	)
	if errors.Is(err, sql.ErrNoRows) {
		return outport.PaymentAddressStatusRecord{}, false, nil
	}
	if err != nil {
		return outport.PaymentAddressStatusRecord{}, false, err
	}

	if !issuedAt.Valid || !trackingID.Valid {
		return outport.PaymentAddressStatusRecord{}, false, outport.ErrPaymentAddressStatusIncomplete
	}

	chain, ok := value_objects.ParseSupportedChain(rawChain)
	if !ok {
		return outport.PaymentAddressStatusRecord{}, false, fmt.Errorf("persisted payment address chain is invalid: %s", rawChain)
	}
	network, ok := value_objects.ParseNetworkID(rawNetwork)
	if !ok {
		return outport.PaymentAddressStatusRecord{}, false, fmt.Errorf("persisted payment address network is invalid: %s", rawNetwork)
	}
	if !receiptStatusRaw.Valid {
		return outport.PaymentAddressStatusRecord{}, false, outport.ErrPaymentAddressStatusIncomplete
	}
	status, ok := value_objects.ParsePaymentReceiptStatus(receiptStatusRaw.String)
	if !ok {
		return outport.PaymentAddressStatusRecord{}, false, fmt.Errorf("persisted payment receipt status is invalid: %s", receiptStatusRaw.String)
	}
	if !requiredConfirmations.Valid || !observedTotalMinor.Valid || !confirmedTotalMinor.Valid ||
		!unconfirmedTotalMinor.Valid || !lastObservedBlockHeight.Valid {
		return outport.PaymentAddressStatusRecord{}, false, outport.ErrPaymentAddressStatusIncomplete
	}

	record := outport.PaymentAddressStatusRecord{
		PaymentAddressID:        paymentAddressID,
		AddressPolicyID:         strings.TrimSpace(addressPolicyID),
		ExpectedAmountMinor:     expectedAmountMinor,
		CustomerReference:       strings.TrimSpace(customerReference),
		Chain:                   chain,
		Network:                 network,
		Scheme:                  strings.TrimSpace(scheme),
		Address:                 strings.TrimSpace(address),
		PaymentStatus:           status,
		ObservedTotalMinor:      observedTotalMinor.Int64,
		ConfirmedTotalMinor:     confirmedTotalMinor.Int64,
		UnconfirmedTotalMinor:   unconfirmedTotalMinor.Int64,
		RequiredConfirmations:   requiredConfirmations.Int32,
		LastObservedBlockHeight: lastObservedBlockHeight.Int64,
		IssuedAt:                issuedAt.Time.UTC(),
		LastError:               strings.TrimSpace(lastError.String),
	}
	if firstObservedAt.Valid {
		timeValue := firstObservedAt.Time.UTC()
		record.FirstObservedAt = &timeValue
	}
	if paidAt.Valid {
		timeValue := paidAt.Time.UTC()
		record.PaidAt = &timeValue
	}
	if confirmedAt.Valid {
		timeValue := confirmedAt.Time.UTC()
		record.ConfirmedAt = &timeValue
	}
	if expiresAt.Valid {
		timeValue := expiresAt.Time.UTC()
		record.ExpiresAt = &timeValue
	}

	return record, true, nil
}
