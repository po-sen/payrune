package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/valueobjects"
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
) (entities.PaymentAddressAllocation, bool, error) {
	if input.PaymentAddressID <= 0 {
		return entities.PaymentAddressAllocation{}, false, nil
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
		address             string
		rawIssuanceRefKind  string
		issuanceRef         string
		failureReason       string
	)

	err := r.executor.QueryRowContext(
		ctx,
		`SELECT id,
		        address_policy_id,
		        slot_index,
		        expected_amount_minor,
		        COALESCE(customer_reference, ''),
		        COALESCE(chain, ''),
		        COALESCE(network, ''),
		        COALESCE(scheme, ''),
		        COALESCE(address, ''),
		        COALESCE(issuance_ref_kind, ''),
		        COALESCE(issuance_ref, ''),
		        COALESCE(failure_reason, '')
		   FROM address_policy_allocations
		  WHERE id = $1
		    AND allocation_status = 'issued'
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
		&address,
		&rawIssuanceRefKind,
		&issuanceRef,
		&failureReason,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return entities.PaymentAddressAllocation{}, false, nil
	}
	if err != nil {
		return entities.PaymentAddressAllocation{}, false, outport.ErrPaymentAddressAllocationStoreFailed
	}
	if slotIndex < 0 || slotIndex > maxSlotIndex {
		return entities.PaymentAddressAllocation{}, false, outport.ErrAddressIndexExhausted
	}

	chain, ok := valueobjects.ParseSupportedChain(rawChain)
	if !ok {
		return entities.PaymentAddressAllocation{}, false, fmt.Errorf(
			"%w: %s",
			outport.ErrPaymentAddressAllocationPersistedChainInvalid,
			rawChain,
		)
	}
	network, ok := valueobjects.ParseNetworkID(rawNetwork)
	if !ok {
		return entities.PaymentAddressAllocation{}, false, fmt.Errorf(
			"%w: %s",
			outport.ErrPaymentAddressAllocationPersistedNetworkInvalid,
			rawNetwork,
		)
	}

	derivationFailureReason, _ := valueobjects.ParsePaymentAddressAllocationDerivationFailureReason(
		strings.TrimSpace(failureReason),
	)
	issuanceRefKind, ok := valueobjects.ParseIssuanceRefKind(rawIssuanceRefKind)
	if !ok {
		return entities.PaymentAddressAllocation{}, false, fmt.Errorf(
			"%w: %s",
			outport.ErrPaymentAddressAllocationPersistedIssuanceRefKindInvalid,
			rawIssuanceRefKind,
		)
	}

	return entities.PaymentAddressAllocation{
		PaymentAddressID:        paymentAddressID,
		AddressPolicyID:         strings.TrimSpace(addressPolicyID),
		SlotIndex:               uint32(slotIndex),
		ExpectedAmountMinor:     expectedAmountMinor,
		CustomerReference:       customerReference,
		Status:                  valueobjects.PaymentAddressAllocationStatusIssued,
		Chain:                   chain,
		Network:                 network,
		Scheme:                  strings.TrimSpace(scheme),
		Address:                 strings.TrimSpace(address),
		IssuanceRefKind:         issuanceRefKind,
		IssuanceRef:             strings.TrimSpace(issuanceRef),
		DerivationFailureReason: derivationFailureReason,
	}, true, nil
}

func (r *PaymentAddressAllocationStore) Complete(
	ctx context.Context,
	allocation entities.PaymentAddressAllocation,
	issuedAt time.Time,
) error {
	if issuedAt.IsZero() {
		return outport.ErrPaymentAddressAllocationIssuedAtRequired
	}

	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE address_policy_allocations
		 SET chain = $2,
		     network = $3,
		     scheme = $4,
		     address = $5,
		     issuance_ref_kind = $6,
		     issuance_ref = $7,
		     allocation_status = 'issued',
		     issued_at = $8,
		     failure_reason = NULL
		 WHERE id = $1 AND allocation_status = 'reserved'`,
		allocation.PaymentAddressID,
		string(allocation.Chain),
		string(allocation.Network),
		string(allocation.Scheme),
		strings.TrimSpace(allocation.Address),
		nullIfEmpty(string(allocation.IssuanceRefKind)),
		nullIfEmpty(allocation.IssuanceRef),
		issuedAt.UTC(),
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
	allocation entities.PaymentAddressAllocation,
) error {
	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE address_policy_allocations
		 SET allocation_status = 'derivation_failed',
		     failure_reason = $2
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
) (entities.PaymentAddressAllocation, bool, error) {
	customerReference := strings.TrimSpace(input.CustomerReference)
	addressSpaceRef := strings.TrimSpace(input.IssuancePolicy.IssuanceConfig.AddressSpaceRef)

	var paymentAddressID int64
	var slotIndex int64
	err := r.executor.QueryRowContext(
		ctx,
		`SELECT id, slot_index
		 FROM address_policy_allocations
		 WHERE address_policy_id = $1
		   AND address_space_ref = $2
		   AND allocation_status = 'derivation_failed'
		 ORDER BY reserved_at ASC, id ASC
		 LIMIT 1
		 FOR UPDATE SKIP LOCKED`,
		input.IssuancePolicy.AddressPolicy.AddressPolicyID,
		addressSpaceRef,
	).Scan(&paymentAddressID, &slotIndex)
	if errors.Is(err, sql.ErrNoRows) {
		return entities.PaymentAddressAllocation{}, false, nil
	}
	if err != nil {
		return entities.PaymentAddressAllocation{}, false, outport.ErrPaymentAddressAllocationStoreFailed
	}
	if slotIndex < 0 || slotIndex > maxSlotIndex {
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
		     issuance_ref_kind = NULL,
		     issuance_ref = NULL,
		     reserved_at = NOW(),
		     issued_at = NULL
		 WHERE id = $1`,
		paymentAddressID,
		input.ExpectedAmountMinor,
		nullIfEmpty(customerReference),
	); err != nil {
		return entities.PaymentAddressAllocation{}, false, outport.ErrPaymentAddressAllocationStoreFailed
	}

	return entities.PaymentAddressAllocation{
		PaymentAddressID:    paymentAddressID,
		AddressPolicyID:     input.IssuancePolicy.AddressPolicy.AddressPolicyID,
		SlotIndex:           uint32(slotIndex),
		ExpectedAmountMinor: input.ExpectedAmountMinor,
		CustomerReference:   customerReference,
		Status:              valueobjects.PaymentAddressAllocationStatusReserved,
	}, true, nil
}

func (r *PaymentAddressAllocationStore) ReserveFresh(
	ctx context.Context,
	input outport.ReservePaymentAddressAllocationInput,
) (entities.PaymentAddressAllocation, error) {
	customerReference := strings.TrimSpace(input.CustomerReference)
	addressSpaceRef := strings.TrimSpace(input.IssuancePolicy.IssuanceConfig.AddressSpaceRef)

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
			 ON CONFLICT (address_policy_id, address_space_ref) DO NOTHING`,
		input.IssuancePolicy.AddressPolicy.AddressPolicyID,
		addressSpaceRef,
	); err != nil {
		return entities.PaymentAddressAllocation{}, outport.ErrPaymentAddressAllocationStoreFailed
	}

	var nextIndex int64
	err := r.executor.QueryRowContext(
		ctx,
		`SELECT next_index
		 FROM address_policy_cursors
		 WHERE address_policy_id = $1
		   AND address_space_ref = $2
		 FOR UPDATE`,
		input.IssuancePolicy.AddressPolicy.AddressPolicyID,
		addressSpaceRef,
	).Scan(&nextIndex)
	if err != nil {
		return entities.PaymentAddressAllocation{}, outport.ErrPaymentAddressAllocationStoreFailed
	}
	if nextIndex > maxSlotIndex {
		return entities.PaymentAddressAllocation{}, outport.ErrAddressIndexExhausted
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
		input.IssuancePolicy.AddressPolicy.AddressPolicyID,
		addressSpaceRef,
		nextIndex,
		input.ExpectedAmountMinor,
		nullIfEmpty(customerReference),
	).Scan(&paymentAddressID)
	if err != nil {
		return entities.PaymentAddressAllocation{}, outport.ErrPaymentAddressAllocationStoreFailed
	}

	if _, err := r.executor.ExecContext(
		ctx,
		`UPDATE address_policy_cursors
		 SET next_index = next_index + 1,
		     updated_at = NOW()
		 WHERE address_policy_id = $1
		   AND address_space_ref = $2`,
		input.IssuancePolicy.AddressPolicy.AddressPolicyID,
		addressSpaceRef,
	); err != nil {
		return entities.PaymentAddressAllocation{}, outport.ErrPaymentAddressAllocationStoreFailed
	}

	return entities.PaymentAddressAllocation{
		PaymentAddressID:    paymentAddressID,
		AddressPolicyID:     input.IssuancePolicy.AddressPolicy.AddressPolicyID,
		SlotIndex:           uint32(nextIndex),
		ExpectedAmountMinor: input.ExpectedAmountMinor,
		CustomerReference:   customerReference,
		Status:              valueobjects.PaymentAddressAllocationStatusReserved,
	}, nil
}

func nullIfEmpty(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}
