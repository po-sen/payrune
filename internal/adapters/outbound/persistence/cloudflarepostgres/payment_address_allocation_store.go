package cloudflarepostgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	outport "payrune/internal/application/ports/outbound"
)

const maxSlotIndex int64 = 0x7fffffff

type PaymentAddressAllocationStore struct {
	executor executor
}

func NewPaymentAddressAllocationStore(executor executor) *PaymentAddressAllocationStore {
	return &PaymentAddressAllocationStore{executor: executor}
}

func (r *PaymentAddressAllocationStore) FindIssuedByID(
	ctx context.Context,
	input outport.FindIssuedPaymentAddressAllocationByIDInput,
) (outport.PaymentAddressAllocationRecord, bool, error) {
	if input.PaymentAddressID <= 0 {
		return outport.PaymentAddressAllocationRecord{}, false, nil
	}

	var (
		paymentAddressID    int64
		addressPolicyID     string
		slotIndex           int64
		expectedAmountMinor int64
		customerReference   string
		rawChain            string
		rawNetwork          string
		scheme              string
		rawAssetReference   string
		address             string
		failureReason       string
	)

	err := r.executor.QueryRowContext(
		ctx,
		`SELECT a.id,
		        a.address_policy_id,
		        a.slot_index,
		        a.expected_amount_minor,
		        COALESCE(a.customer_reference, ''),
		        COALESCE(a.chain, ''),
		        COALESCE(a.network, ''),
		        COALESCE(a.scheme, ''),
		        COALESCE(a.asset_reference, ''),
		        COALESCE(a.address, ''),
		        COALESCE(a.failure_reason, '')
		   FROM address_policy_allocations a
		  WHERE a.id = $1
		    AND a.allocation_status = 'issued'
		  LIMIT 1`,
		input.PaymentAddressID,
	).Scan(
		&paymentAddressID,
		&addressPolicyID,
		&slotIndex,
		&expectedAmountMinor,
		&customerReference,
		&rawChain,
		&rawNetwork,
		&scheme,
		&rawAssetReference,
		&address,
		&failureReason,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return outport.PaymentAddressAllocationRecord{}, false, nil
	}
	if err != nil {
		return outport.PaymentAddressAllocationRecord{}, false, outport.ErrPaymentAddressAllocationStoreFailed
	}
	if slotIndex < 0 || slotIndex > maxSlotIndex {
		return outport.PaymentAddressAllocationRecord{}, false, outport.ErrAddressIndexExhausted
	}

	chain, ok := outport.NormalizeSupportedChain(rawChain)
	if !ok {
		return outport.PaymentAddressAllocationRecord{}, false, fmt.Errorf(
			"%w: %s",
			outport.ErrPaymentAddressAllocationPersistedChainInvalid,
			rawChain,
		)
	}
	parsedAddressPolicyID, ok := outport.NormalizeAddressPolicyID(addressPolicyID)
	if !ok {
		return outport.PaymentAddressAllocationRecord{}, false, fmt.Errorf(
			"%w: %s",
			outport.ErrPaymentAddressAllocationPersistedAddressPolicyIDInvalid,
			strings.TrimSpace(addressPolicyID),
		)
	}
	network, ok := outport.NormalizeNetworkID(rawNetwork)
	if !ok {
		return outport.PaymentAddressAllocationRecord{}, false, fmt.Errorf(
			"%w: %s",
			outport.ErrPaymentAddressAllocationPersistedNetworkInvalid,
			rawNetwork,
		)
	}
	normalizedScheme, ok := outport.NormalizeAddressScheme(scheme)
	if !ok {
		return outport.PaymentAddressAllocationRecord{}, false, outport.ErrPaymentAddressAllocationStoreFailed
	}
	assetReference := strings.TrimSpace(rawAssetReference)

	derivationFailureReason := normalizePaymentAddressAllocationDerivationFailureReason(failureReason)

	return outport.PaymentAddressAllocationRecord{
		PaymentAddressID:        paymentAddressID,
		AddressPolicyID:         parsedAddressPolicyID,
		SlotIndex:               uint32(slotIndex),
		ExpectedAmountMinor:     expectedAmountMinor,
		CustomerReference:       customerReference,
		Status:                  outport.PaymentAddressAllocationStatusIssued,
		Chain:                   chain,
		Network:                 network,
		Scheme:                  normalizedScheme,
		AssetReference:          assetReference,
		Address:                 strings.TrimSpace(address),
		DerivationFailureReason: derivationFailureReason,
	}, true, nil
}

func (r *PaymentAddressAllocationStore) Complete(
	ctx context.Context,
	input outport.CompletePaymentAddressAllocationInput,
) error {
	if input.IssuedAt.IsZero() {
		return outport.ErrPaymentAddressAllocationIssuedAtRequired
	}
	sweepMaterialJSON := strings.TrimSpace(input.SweepMaterial)
	if sweepMaterialJSON == "" {
		return outport.ErrPaymentAddressAllocationStoreFailed
	}
	assetReference := strings.TrimSpace(input.Allocation.AssetReference)

	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE address_policy_allocations
		 SET chain = $2,
		     network = $3,
		     scheme = $4,
		     address = $5,
		     asset_reference = $6,
		     sweep_material_json = $7,
		     failure_reason = NULL,
		     allocation_status = 'issued',
		     issued_at = $8
		 WHERE id = $1 AND allocation_status = 'reserved'`,
		input.Allocation.PaymentAddressID,
		string(input.Allocation.Chain),
		string(input.Allocation.Network),
		string(input.Allocation.Scheme),
		strings.TrimSpace(input.Allocation.Address),
		nullIfEmpty(assetReference),
		sweepMaterialJSON,
		input.IssuedAt.UTC(),
	)
	if err != nil {
		return outport.ErrPaymentAddressAllocationStoreFailed
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return outport.ErrPaymentAddressAllocationStoreFailed
	}
	if rowsAffected == 0 {
		return outport.ErrPaymentAddressAllocationNotReserved
	}

	return nil
}

func (r *PaymentAddressAllocationStore) MarkDerivationFailed(
	ctx context.Context,
	allocation outport.PaymentAddressAllocationRecord,
) error {
	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE address_policy_allocations
		 SET failure_reason = $2,
		     allocation_status = 'derivation_failed'
		 WHERE id = $1 AND allocation_status = 'reserved'`,
		allocation.PaymentAddressID,
		nullIfEmpty(string(allocation.DerivationFailureReason)),
	)
	if err != nil {
		return outport.ErrPaymentAddressAllocationStoreFailed
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return outport.ErrPaymentAddressAllocationStoreFailed
	}
	if rowsAffected == 0 {
		return outport.ErrPaymentAddressAllocationNotReserved
	}

	return nil
}

func (r *PaymentAddressAllocationStore) ReopenFailedReservation(
	ctx context.Context,
	input outport.ReservePaymentAddressAllocationInput,
) (outport.PaymentAddressAllocationRecord, bool, error) {
	customerReference := strings.TrimSpace(input.CustomerReference)
	addressPolicyID := input.IssuancePolicy.AddressPolicyID
	addressSpaceRef := strings.TrimSpace(input.IssuancePolicy.AddressSpaceRef)

	var paymentAddressID int64
	var slotIndex int64
	err := r.executor.QueryRowContext(
		ctx,
		`SELECT id, slot_index
		   FROM address_policy_allocations a
		  WHERE a.address_policy_id = $1
		    AND a.address_space_ref = $2
		    AND a.allocation_status = 'derivation_failed'
		  ORDER BY a.reserved_at ASC, a.id ASC
		  LIMIT 1
		  FOR UPDATE SKIP LOCKED`,
		addressPolicyID,
		addressSpaceRef,
	).Scan(&paymentAddressID, &slotIndex)
	if errors.Is(err, sql.ErrNoRows) {
		return outport.PaymentAddressAllocationRecord{}, false, nil
	}
	if err != nil {
		return outport.PaymentAddressAllocationRecord{}, false, outport.ErrPaymentAddressAllocationStoreFailed
	}
	if slotIndex < 0 || slotIndex > maxSlotIndex {
		return outport.PaymentAddressAllocationRecord{}, false, outport.ErrAddressIndexExhausted
	}

	if _, err := r.executor.ExecContext(
		ctx,
		`UPDATE address_policy_allocations
		 SET allocation_status = 'reserved',
		     expected_amount_minor = $2,
		     customer_reference = $3,
		     chain = NULL,
		     network = NULL,
		     scheme = NULL,
		     asset_reference = NULL,
		     address = NULL,
		     sweep_material_json = NULL,
		     failure_reason = NULL,
		     reserved_at = NOW(),
		     issued_at = NULL
		 WHERE id = $1`,
		paymentAddressID,
		input.ExpectedAmountMinor,
		nullIfEmpty(customerReference),
	); err != nil {
		return outport.PaymentAddressAllocationRecord{}, false, outport.ErrPaymentAddressAllocationStoreFailed
	}

	return outport.PaymentAddressAllocationRecord{
		PaymentAddressID:    paymentAddressID,
		AddressPolicyID:     addressPolicyID,
		SlotIndex:           uint32(slotIndex),
		ExpectedAmountMinor: input.ExpectedAmountMinor,
		CustomerReference:   customerReference,
		Status:              outport.PaymentAddressAllocationStatusReserved,
	}, true, nil
}

func (r *PaymentAddressAllocationStore) ReserveFresh(
	ctx context.Context,
	input outport.ReservePaymentAddressAllocationInput,
) (outport.PaymentAddressAllocationRecord, error) {
	customerReference := strings.TrimSpace(input.CustomerReference)
	addressPolicyID := input.IssuancePolicy.AddressPolicyID
	addressSpaceRef := strings.TrimSpace(input.IssuancePolicy.AddressSpaceRef)

	if _, err := r.executor.ExecContext(
		ctx,
		`INSERT INTO address_policy_cursors (
			   address_policy_id,
			   address_space_ref,
			   next_index
			 )
		 SELECT $1,
		        $2,
		        COALESCE(
		          (
		            SELECT MAX(slot_index) + 1
		              FROM address_policy_allocations
		             WHERE address_policy_id = $1
		               AND address_space_ref = $2
		          ),
		          0
		        )
		 ON CONFLICT DO NOTHING`,
		addressPolicyID,
		addressSpaceRef,
	); err != nil {
		return outport.PaymentAddressAllocationRecord{}, outport.ErrPaymentAddressAllocationStoreFailed
	}

	var nextIndex int64
	err := r.executor.QueryRowContext(
		ctx,
		`SELECT next_index
		 FROM address_policy_cursors
		 WHERE address_policy_id = $1
		   AND address_space_ref = $2
		 FOR UPDATE`,
		addressPolicyID,
		addressSpaceRef,
	).Scan(&nextIndex)
	if err != nil {
		return outport.PaymentAddressAllocationRecord{}, outport.ErrPaymentAddressAllocationStoreFailed
	}
	if nextIndex > maxSlotIndex {
		return outport.PaymentAddressAllocationRecord{}, outport.ErrAddressIndexExhausted
	}

	var paymentAddressID int64
	err = r.executor.QueryRowContext(
		ctx,
		`INSERT INTO address_policy_allocations (
			   address_policy_id,
			   address_space_ref,
			   slot_index,
			   expected_amount_minor,
			   customer_reference,
			   allocation_status
			 )
		 VALUES ($1, $2, $3, $4, $5, 'reserved')
		 RETURNING id`,
		addressPolicyID,
		addressSpaceRef,
		nextIndex,
		input.ExpectedAmountMinor,
		nullIfEmpty(customerReference),
	).Scan(&paymentAddressID)
	if err != nil {
		return outport.PaymentAddressAllocationRecord{}, outport.ErrPaymentAddressAllocationStoreFailed
	}

	if _, err := r.executor.ExecContext(
		ctx,
		`UPDATE address_policy_cursors
		 SET next_index = next_index + 1,
		     updated_at = NOW()
		 WHERE address_policy_id = $1
		   AND address_space_ref = $2`,
		addressPolicyID,
		addressSpaceRef,
	); err != nil {
		return outport.PaymentAddressAllocationRecord{}, outport.ErrPaymentAddressAllocationStoreFailed
	}

	return outport.PaymentAddressAllocationRecord{
		PaymentAddressID:    paymentAddressID,
		AddressPolicyID:     addressPolicyID,
		SlotIndex:           uint32(nextIndex),
		ExpectedAmountMinor: input.ExpectedAmountMinor,
		CustomerReference:   customerReference,
		Status:              outport.PaymentAddressAllocationStatusReserved,
	}, nil
}
