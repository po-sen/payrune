package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/valueobjects"
)

const maxNonHardenedIndex int64 = 0x7fffffff

var errAllocationNotReserved = errors.New("address allocation is not reserved")

type PaymentAddressAllocationStore struct {
	executor Executor
}

func NewPaymentAddressAllocationStore(executor Executor) *PaymentAddressAllocationStore {
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
		derivationIndex     int64
		expectedAmountMinor int64
		customerReference   string
		rawChain            string
		rawNetwork          string
		scheme              string
		assetCode           string
		assetType           string
		tokenAddress        string
		minorUnit           string
		decimals            int32
		issuanceMethod      string
		address             string
		derivationPath      string
		failureReason       string
	)

	err := r.executor.QueryRowContext(
		ctx,
		`SELECT id,
		        address_policy_id,
		        derivation_index,
		        expected_amount_minor,
		        COALESCE(customer_reference, ''),
		        COALESCE(chain, ''),
		        COALESCE(network, ''),
		        COALESCE(scheme, ''),
		        COALESCE(asset_code, ''),
		        COALESCE(asset_type, ''),
		        COALESCE(token_address, ''),
		        COALESCE(minor_unit, ''),
		        COALESCE(decimals, 0),
		        COALESCE(issuance_method, ''),
		        COALESCE(address, ''),
		        COALESCE(derivation_path, ''),
		        COALESCE(failure_reason, '')
		   FROM address_policy_allocations
		  WHERE id = $1
		    AND allocation_status = 'issued'
		  LIMIT 1`,
		input.PaymentAddressID,
	).Scan(
		&paymentAddressID,
		&addressPolicyID,
		&derivationIndex,
		&expectedAmountMinor,
		&customerReference,
		&rawChain,
		&rawNetwork,
		&scheme,
		&assetCode,
		&assetType,
		&tokenAddress,
		&minorUnit,
		&decimals,
		&issuanceMethod,
		&address,
		&derivationPath,
		&failureReason,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return entities.PaymentAddressAllocation{}, false, nil
	}
	if err != nil {
		return entities.PaymentAddressAllocation{}, false, err
	}
	if derivationIndex < 0 || derivationIndex > maxNonHardenedIndex {
		return entities.PaymentAddressAllocation{}, false, outport.ErrAddressIndexExhausted
	}

	chain, ok := valueobjects.ParseSupportedChain(rawChain)
	if !ok {
		return entities.PaymentAddressAllocation{}, false, errors.New("persisted allocation chain is invalid")
	}
	network, ok := valueobjects.ParseNetworkID(rawNetwork)
	if !ok {
		return entities.PaymentAddressAllocation{}, false, errors.New("persisted allocation network is invalid")
	}

	return entities.PaymentAddressAllocation{
		PaymentAddressID:    paymentAddressID,
		AddressPolicyID:     strings.TrimSpace(addressPolicyID),
		DerivationIndex:     uint32(derivationIndex),
		ExpectedAmountMinor: expectedAmountMinor,
		CustomerReference:   customerReference,
		Status:              valueobjects.PaymentAddressAllocationStatusIssued,
		Chain:               chain,
		Network:             network,
		Scheme:              strings.TrimSpace(scheme),
		AssetCode:           strings.TrimSpace(assetCode),
		AssetType:           strings.TrimSpace(assetType),
		TokenAddress:        strings.TrimSpace(tokenAddress),
		MinorUnit:           strings.TrimSpace(minorUnit),
		Decimals:            uint8(decimals),
		IssuanceMethod:      strings.TrimSpace(issuanceMethod),
		Address:             strings.TrimSpace(address),
		DerivationPath:      strings.TrimSpace(derivationPath),
		FailureReason:       strings.TrimSpace(failureReason),
	}, true, nil
}

func (r *PaymentAddressAllocationStore) Complete(
	ctx context.Context,
	allocation entities.PaymentAddressAllocation,
	issuedAt time.Time,
) error {
	if issuedAt.IsZero() {
		return errors.New("issued at is required")
	}

	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE address_policy_allocations
		 SET chain = $2,
		     network = $3,
		     scheme = $4,
		     asset_code = $5,
		     asset_type = $6,
		     token_address = $7,
		     minor_unit = $8,
		     decimals = $9,
		     issuance_method = $10,
		     address = $11,
		     derivation_path = $12,
		     allocation_status = 'issued',
		     issued_at = $13,
		     failure_reason = NULL
		 WHERE id = $1 AND allocation_status = 'reserved'`,
		allocation.PaymentAddressID,
		string(allocation.Chain),
		string(allocation.Network),
		string(allocation.Scheme),
		strings.TrimSpace(allocation.AssetCode),
		strings.TrimSpace(allocation.AssetType),
		nullIfEmpty(allocation.TokenAddress),
		strings.TrimSpace(allocation.MinorUnit),
		allocation.Decimals,
		strings.TrimSpace(allocation.IssuanceMethod),
		strings.TrimSpace(allocation.Address),
		nullIfEmpty(allocation.DerivationPath),
		issuedAt.UTC(),
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

func (r *PaymentAddressAllocationStore) ReopenFailedReservation(
	ctx context.Context,
	input outport.ReservePaymentAddressAllocationInput,
) (entities.PaymentAddressAllocation, bool, error) {
	customerReference := strings.TrimSpace(input.CustomerReference)
	assetCode := strings.TrimSpace(input.IssuancePolicy.AddressPolicy.AssetCode)
	assetType := strings.TrimSpace(input.IssuancePolicy.AddressPolicy.AssetType)
	tokenAddress := nullIfEmpty(input.IssuancePolicy.AddressPolicy.TokenAddress)
	minorUnit := strings.TrimSpace(input.IssuancePolicy.AddressPolicy.MinorUnit)
	decimals := input.IssuancePolicy.AddressPolicy.Decimals
	issuanceMethod := issuanceMethodForPolicy(input.IssuancePolicy)

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
		input.IssuancePolicy.AddressPolicy.AddressPolicyID,
		input.IssuancePolicy.DerivationConfig.PublicKeyFingerprintAlgo,
		input.IssuancePolicy.DerivationConfig.PublicKeyFingerprint,
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
		     asset_code = $4,
		     asset_type = $5,
		     token_address = $6,
		     minor_unit = $7,
		     decimals = $8,
		     issuance_method = $9,
		     address = NULL,
		     derivation_path = NULL,
		     reserved_at = NOW(),
		     issued_at = NULL
		 WHERE id = $1`,
		paymentAddressID,
		input.ExpectedAmountMinor,
		nullIfEmpty(customerReference),
		assetCode,
		assetType,
		tokenAddress,
		minorUnit,
		decimals,
		issuanceMethod,
	); err != nil {
		return entities.PaymentAddressAllocation{}, false, err
	}

	return entities.PaymentAddressAllocation{
		PaymentAddressID:    paymentAddressID,
		AddressPolicyID:     input.IssuancePolicy.AddressPolicy.AddressPolicyID,
		DerivationIndex:     uint32(derivationIndex),
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
	assetCode := strings.TrimSpace(input.IssuancePolicy.AddressPolicy.AssetCode)
	assetType := strings.TrimSpace(input.IssuancePolicy.AddressPolicy.AssetType)
	tokenAddress := nullIfEmpty(input.IssuancePolicy.AddressPolicy.TokenAddress)
	minorUnit := strings.TrimSpace(input.IssuancePolicy.AddressPolicy.MinorUnit)
	decimals := input.IssuancePolicy.AddressPolicy.Decimals
	issuanceMethod := issuanceMethodForPolicy(input.IssuancePolicy)

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
		input.IssuancePolicy.AddressPolicy.AddressPolicyID,
		input.IssuancePolicy.DerivationConfig.PublicKeyFingerprintAlgo,
		input.IssuancePolicy.DerivationConfig.PublicKeyFingerprint,
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
		input.IssuancePolicy.AddressPolicy.AddressPolicyID,
		input.IssuancePolicy.DerivationConfig.PublicKeyFingerprintAlgo,
		input.IssuancePolicy.DerivationConfig.PublicKeyFingerprint,
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
			   asset_code,
			   asset_type,
			   token_address,
			   minor_unit,
			   decimals,
			   issuance_method,
			   allocation_status
			 )
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 'reserved')
		 RETURNING id`,
		input.IssuancePolicy.AddressPolicy.AddressPolicyID,
		input.IssuancePolicy.DerivationConfig.PublicKeyFingerprintAlgo,
		input.IssuancePolicy.DerivationConfig.PublicKeyFingerprint,
		nextIndex,
		input.ExpectedAmountMinor,
		nullIfEmpty(customerReference),
		assetCode,
		assetType,
		tokenAddress,
		minorUnit,
		decimals,
		issuanceMethod,
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
		input.IssuancePolicy.AddressPolicy.AddressPolicyID,
		input.IssuancePolicy.DerivationConfig.PublicKeyFingerprintAlgo,
		input.IssuancePolicy.DerivationConfig.PublicKeyFingerprint,
	); err != nil {
		return entities.PaymentAddressAllocation{}, err
	}

	return entities.PaymentAddressAllocation{
		PaymentAddressID:    paymentAddressID,
		AddressPolicyID:     input.IssuancePolicy.AddressPolicy.AddressPolicyID,
		DerivationIndex:     uint32(nextIndex),
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

func issuanceMethodForPolicy(policy entities.AddressIssuancePolicy) string {
	if policy.AddressPolicy.Chain == valueobjects.SupportedChainEthereum {
		return "create2_forwarder"
	}
	return "xpub_derivation"
}
