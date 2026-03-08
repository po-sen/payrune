package use_cases

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/in"
	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/policies"
)

type allocatePaymentAddressUseCase struct {
	unitOfWork     outport.UnitOfWork
	deriver        outport.ChainAddressDeriver
	policyReader   outport.AddressPolicyReader
	issuancePolicy policies.PaymentAddressAllocationIssuancePolicy
	clock          outport.Clock
}

type allocatePaymentAddressTxOutcome struct {
	allocation       entities.PaymentAddressAllocation
	issuedAllocation entities.PaymentAddressAllocation
	// persistedDerivationFailureErr is returned after commit when derivation
	// failed but the failure state was persisted successfully.
	persistedDerivationFailureErr error
}

type allocatePaymentAddressTxScope struct {
	allocationStore      outport.PaymentAddressAllocationStore
	idempotencyStore     outport.PaymentAddressIdempotencyStore
	receiptTrackingStore outport.PaymentReceiptTrackingStore
}

type allocatePaymentAddressDerivationOutcome struct {
	issuedAllocation entities.PaymentAddressAllocation
	// persistedDerivationFailureErr is returned after commit when derivation
	// failed but the failure state was persisted successfully.
	persistedDerivationFailureErr error
}

type allocatePaymentAddressReplayTxScope struct {
	allocationStore  outport.PaymentAddressAllocationStore
	idempotencyStore outport.PaymentAddressIdempotencyStore
}

func NewAllocatePaymentAddressUseCase(
	unitOfWork outport.UnitOfWork,
	deriver outport.ChainAddressDeriver,
	policyReader outport.AddressPolicyReader,
	issuancePolicy policies.PaymentAddressAllocationIssuancePolicy,
	clock outport.Clock,
) inport.AllocatePaymentAddressUseCase {
	return &allocatePaymentAddressUseCase{
		unitOfWork:     unitOfWork,
		deriver:        deriver,
		policyReader:   policyReader,
		issuancePolicy: issuancePolicy,
		clock:          clock,
	}
}

func (uc *allocatePaymentAddressUseCase) Execute(
	ctx context.Context,
	input dto.AllocatePaymentAddressInput,
) (dto.AllocatePaymentAddressResponse, error) {
	if uc.unitOfWork == nil {
		return dto.AllocatePaymentAddressResponse{}, errors.New("unit of work is not configured")
	}
	if uc.deriver == nil {
		return dto.AllocatePaymentAddressResponse{}, errors.New("chain address deriver is not configured")
	}
	if uc.policyReader == nil {
		return dto.AllocatePaymentAddressResponse{}, errors.New("address policy reader is not configured")
	}
	if uc.clock == nil {
		return dto.AllocatePaymentAddressResponse{}, errors.New("clock is not configured")
	}
	if !uc.deriver.SupportsChain(input.Chain) {
		return dto.AllocatePaymentAddressResponse{}, inport.ErrChainNotSupported
	}

	existingAllocation, found, err := uc.findExistingIssuedAllocation(ctx, input)
	if err != nil {
		return dto.AllocatePaymentAddressResponse{}, err
	}
	if found {
		return uc.buildExistingIssuedAllocationResponse(ctx, existingAllocation)
	}

	policy, err := uc.loadIssuancePolicy(ctx, input.AddressPolicyID)
	if err != nil {
		return dto.AllocatePaymentAddressResponse{}, err
	}
	issuedAt := uc.clock.NowUTC()
	issuancePlan, err := uc.buildIssuancePlan(policy, input, issuedAt)
	if err != nil {
		return dto.AllocatePaymentAddressResponse{}, err
	}
	txOutcome, err := uc.executeIssuanceTransaction(ctx, input, issuancePlan, issuedAt)
	if err != nil {
		if errors.Is(err, outport.ErrPaymentAddressIdempotencyKeyExists) {
			existingAllocation, found, lookupErr := uc.findExistingIssuedAllocation(ctx, input)
			if lookupErr != nil {
				return dto.AllocatePaymentAddressResponse{}, lookupErr
			}
			if found {
				return uc.buildExistingIssuedAllocationResponse(ctx, existingAllocation)
			}
			return dto.AllocatePaymentAddressResponse{}, errors.New(
				"idempotency key claim conflict occurred but no completed idempotency record was found",
			)
		}
		if errors.Is(err, outport.ErrAddressIndexExhausted) {
			return dto.AllocatePaymentAddressResponse{}, inport.ErrAddressPoolExhausted
		}
		return dto.AllocatePaymentAddressResponse{}, err
	}
	if txOutcome.persistedDerivationFailureErr != nil {
		return dto.AllocatePaymentAddressResponse{}, txOutcome.persistedDerivationFailureErr
	}

	return uc.buildAllocatePaymentAddressResponse(policy.AddressPolicy, txOutcome.issuedAllocation)
}

func (uc *allocatePaymentAddressUseCase) loadIssuancePolicy(
	ctx context.Context,
	addressPolicyID string,
) (entities.AddressIssuancePolicy, error) {
	policy, ok, err := uc.policyReader.FindIssuanceByID(ctx, addressPolicyID)
	if err != nil {
		return entities.AddressIssuancePolicy{}, err
	}
	if !ok {
		return entities.AddressIssuancePolicy{}, inport.ErrAddressPolicyNotFound
	}
	return policy, nil
}

func (uc *allocatePaymentAddressUseCase) buildIssuancePlan(
	policy entities.AddressIssuancePolicy,
	input dto.AllocatePaymentAddressInput,
	issuedAt time.Time,
) (policies.PaymentAddressAllocationIssuancePlan, error) {
	issuancePlan, err := uc.issuancePolicy.Plan(
		policy,
		input.Chain,
		input.ExpectedAmountMinor,
		input.CustomerReference,
		issuedAt,
	)
	if err != nil {
		return policies.PaymentAddressAllocationIssuancePlan{}, mapAllocatePaymentAddressIssuancePlanError(err)
	}
	return issuancePlan, nil
}

// Keep reservation, issued-allocation state, and receipt tracking creation in
// one transaction so the address index state cannot diverge from the receipt
// tracking state.
func (uc *allocatePaymentAddressUseCase) executeIssuanceTransaction(
	ctx context.Context,
	input dto.AllocatePaymentAddressInput,
	issuancePlan policies.PaymentAddressAllocationIssuancePlan,
	issuedAt time.Time,
) (allocatePaymentAddressTxOutcome, error) {
	var outcome allocatePaymentAddressTxOutcome

	reserveInput := outport.ReservePaymentAddressAllocationInput{
		IssuancePolicy:      issuancePlan.Reservation.IssuancePolicy,
		ExpectedAmountMinor: issuancePlan.Reservation.ExpectedAmountMinor,
		CustomerReference:   issuancePlan.Reservation.CustomerReference,
	}
	idempotencyClaimInput := outport.ClaimPaymentAddressIdempotencyInput{
		Chain:               input.Chain,
		IdempotencyKey:      strings.TrimSpace(input.IdempotencyKey),
		AddressPolicyID:     issuancePlan.Reservation.IssuancePolicy.AddressPolicy.AddressPolicyID,
		ExpectedAmountMinor: issuancePlan.Reservation.ExpectedAmountMinor,
		CustomerReference:   issuancePlan.Reservation.CustomerReference,
	}

	err := uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
		stores, err := requireAllocatePaymentAddressTxScope(txScope)
		if err != nil {
			return err
		}
		if err := uc.claimIdempotencyKeyIfPresent(ctx, stores.idempotencyStore, idempotencyClaimInput); err != nil {
			return err
		}

		allocation, err := uc.reserveAllocation(
			ctx,
			stores.allocationStore,
			reserveInput,
			issuancePlan.ReservationAttempts,
		)
		if err != nil {
			return err
		}
		outcome.allocation = allocation

		derivationOutcome, err := uc.deriveIssuedAllocation(
			ctx,
			stores.allocationStore,
			issuancePlan.Reservation.IssuancePolicy,
			allocation,
		)
		if err != nil {
			return err
		}
		if derivationOutcome.persistedDerivationFailureErr != nil {
			outcome.persistedDerivationFailureErr = derivationOutcome.persistedDerivationFailureErr
			if err := uc.releaseIdempotencyKeyIfPresent(
				ctx,
				stores.idempotencyStore,
				outport.ReleasePaymentAddressIdempotencyInput{
					Chain:          input.Chain,
					IdempotencyKey: idempotencyClaimInput.IdempotencyKey,
				},
			); err != nil {
				return err
			}
			return nil
		}
		outcome.issuedAllocation = derivationOutcome.issuedAllocation

		return uc.persistIssuedAllocation(
			ctx,
			stores,
			outcome.issuedAllocation,
			outport.CompletePaymentAddressIdempotencyInput{
				Chain:            input.Chain,
				IdempotencyKey:   idempotencyClaimInput.IdempotencyKey,
				PaymentAddressID: outcome.issuedAllocation.PaymentAddressID,
			},
			issuedAt,
			issuancePlan.ReceiptTerms,
		)
	})
	if err != nil {
		return allocatePaymentAddressTxOutcome{}, err
	}
	return outcome, nil
}

func (uc *allocatePaymentAddressUseCase) reserveAllocation(
	ctx context.Context,
	allocationStore outport.PaymentAddressAllocationStore,
	reserveInput outport.ReservePaymentAddressAllocationInput,
	attempts []policies.PaymentAddressAllocationReservationAttempt,
) (entities.PaymentAddressAllocation, error) {
	for _, attempt := range attempts {
		switch attempt {
		case policies.PaymentAddressAllocationReservationAttemptReopenFailed:
			reopenedAllocation, reopened, err := allocationStore.ReopenFailedReservation(ctx, reserveInput)
			if err != nil {
				return entities.PaymentAddressAllocation{}, err
			}
			if reopened {
				return reopenedAllocation, nil
			}
		case policies.PaymentAddressAllocationReservationAttemptReserveFresh:
			return allocationStore.ReserveFresh(ctx, reserveInput)
		default:
			return entities.PaymentAddressAllocation{}, errors.New("payment address allocation reservation attempt is invalid")
		}
	}

	return entities.PaymentAddressAllocation{}, errors.New("payment address allocation reservation attempts are required")
}

func (uc *allocatePaymentAddressUseCase) deriveIssuedAllocation(
	ctx context.Context,
	allocationStore outport.PaymentAddressAllocationStore,
	policy entities.AddressIssuancePolicy,
	allocation entities.PaymentAddressAllocation,
) (allocatePaymentAddressDerivationOutcome, error) {
	output, err := uc.deriver.DeriveAddress(ctx, outport.DeriveChainAddressInput{
		Chain:            policy.AddressPolicy.Chain,
		Network:          policy.AddressPolicy.Network,
		Scheme:           policy.AddressPolicy.Scheme,
		AccountPublicKey: policy.DerivationConfig.AccountPublicKey,
		Index:            allocation.DerivationIndex,
	})
	if err != nil {
		if persistErr := uc.persistDerivationFailure(ctx, allocationStore, allocation, err); persistErr != nil {
			return allocatePaymentAddressDerivationOutcome{}, persistErr
		}
		return allocatePaymentAddressDerivationOutcome{persistedDerivationFailureErr: err}, nil
	}

	issuedAllocation, err := allocation.MarkIssued(policy, output.Address, output.RelativeDerivationPath)
	if err != nil {
		if persistErr := uc.persistDerivationFailure(ctx, allocationStore, allocation, err); persistErr != nil {
			return allocatePaymentAddressDerivationOutcome{}, persistErr
		}
		return allocatePaymentAddressDerivationOutcome{persistedDerivationFailureErr: err}, nil
	}

	return allocatePaymentAddressDerivationOutcome{issuedAllocation: issuedAllocation}, nil
}

func (uc *allocatePaymentAddressUseCase) persistDerivationFailure(
	ctx context.Context,
	allocationStore outport.PaymentAddressAllocationStore,
	allocation entities.PaymentAddressAllocation,
	cause error,
) error {
	failedAllocation, err := allocation.MarkDerivationFailed(cause.Error())
	if err != nil {
		return err
	}
	return allocationStore.MarkDerivationFailed(ctx, failedAllocation)
}

func (uc *allocatePaymentAddressUseCase) persistIssuedAllocation(
	ctx context.Context,
	stores allocatePaymentAddressTxScope,
	issuedAllocation entities.PaymentAddressAllocation,
	idempotencyCompleteInput outport.CompletePaymentAddressIdempotencyInput,
	issuedAt time.Time,
	receiptTerms policies.PaymentReceiptIssuanceTerms,
) error {
	if err := stores.allocationStore.Complete(ctx, issuedAllocation, issuedAt); err != nil {
		return err
	}

	receiptTracking, err := issuedAllocation.IssueReceiptTracking(
		issuedAt,
		receiptTerms.RequiredConfirmations,
		receiptTerms.ExpiresAt,
	)
	if err != nil {
		return err
	}

	if err := stores.receiptTrackingStore.Create(ctx, receiptTracking, issuedAt); err != nil {
		return err
	}

	return uc.completeIdempotencyKeyIfPresent(ctx, stores.idempotencyStore, idempotencyCompleteInput)
}

func requireAllocatePaymentAddressTxScope(
	txScope outport.TxScope,
) (allocatePaymentAddressTxScope, error) {
	if txScope.PaymentAddressAllocation == nil {
		return allocatePaymentAddressTxScope{}, errors.New("payment address allocation store is not configured")
	}
	if txScope.PaymentAddressIdempotency == nil {
		return allocatePaymentAddressTxScope{}, errors.New("payment address idempotency store is not configured")
	}
	if txScope.PaymentReceiptTracking == nil {
		return allocatePaymentAddressTxScope{}, errors.New("payment receipt tracking store is not configured")
	}

	return allocatePaymentAddressTxScope{
		allocationStore:      txScope.PaymentAddressAllocation,
		idempotencyStore:     txScope.PaymentAddressIdempotency,
		receiptTrackingStore: txScope.PaymentReceiptTracking,
	}, nil
}

func requireAllocatePaymentAddressReplayTxScope(
	txScope outport.TxScope,
) (allocatePaymentAddressReplayTxScope, error) {
	if txScope.PaymentAddressAllocation == nil {
		return allocatePaymentAddressReplayTxScope{}, errors.New("payment address allocation store is not configured")
	}
	if txScope.PaymentAddressIdempotency == nil {
		return allocatePaymentAddressReplayTxScope{}, errors.New("payment address idempotency store is not configured")
	}

	return allocatePaymentAddressReplayTxScope{
		allocationStore:  txScope.PaymentAddressAllocation,
		idempotencyStore: txScope.PaymentAddressIdempotency,
	}, nil
}

func (uc *allocatePaymentAddressUseCase) findExistingIssuedAllocation(
	ctx context.Context,
	input dto.AllocatePaymentAddressInput,
) (entities.PaymentAddressAllocation, bool, error) {
	idempotencyKey := strings.TrimSpace(input.IdempotencyKey)
	if idempotencyKey == "" {
		return entities.PaymentAddressAllocation{}, false, nil
	}

	var allocation entities.PaymentAddressAllocation
	found := false

	err := uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
		stores, err := requireAllocatePaymentAddressReplayTxScope(txScope)
		if err != nil {
			return err
		}

		record, recordFound, err := stores.idempotencyStore.FindByKey(
			ctx,
			outport.FindPaymentAddressIdempotencyInput{
				Chain:          input.Chain,
				IdempotencyKey: idempotencyKey,
			},
		)
		if err != nil {
			return err
		}
		if !recordFound {
			return nil
		}
		if record.PaymentAddressID <= 0 {
			return errors.New("payment address idempotency record is incomplete")
		}
		if record.AddressPolicyID != strings.TrimSpace(input.AddressPolicyID) ||
			record.ExpectedAmountMinor != input.ExpectedAmountMinor ||
			record.CustomerReference != strings.TrimSpace(input.CustomerReference) {
			return inport.ErrIdempotencyKeyConflict
		}

		existingAllocation, allocationFound, err := stores.allocationStore.FindIssuedByID(
			ctx,
			outport.FindIssuedPaymentAddressAllocationByIDInput{PaymentAddressID: record.PaymentAddressID},
		)
		if err != nil {
			return err
		}
		if !allocationFound {
			return errors.New("completed payment address idempotency record references missing issued allocation")
		}

		allocation = existingAllocation
		found = true
		return nil
	})
	if err != nil {
		return entities.PaymentAddressAllocation{}, false, err
	}
	return allocation, found, nil
}

func (uc *allocatePaymentAddressUseCase) buildExistingIssuedAllocationResponse(
	ctx context.Context,
	allocation entities.PaymentAddressAllocation,
) (dto.AllocatePaymentAddressResponse, error) {
	policy, err := uc.loadIssuancePolicy(ctx, allocation.AddressPolicyID)
	if err != nil {
		return dto.AllocatePaymentAddressResponse{}, err
	}

	response, err := uc.buildAllocatePaymentAddressResponse(policy.AddressPolicy, allocation)
	if err != nil {
		return dto.AllocatePaymentAddressResponse{}, err
	}
	response.IdempotencyReplayed = true
	return response, nil
}

func (uc *allocatePaymentAddressUseCase) claimIdempotencyKeyIfPresent(
	ctx context.Context,
	idempotencyStore outport.PaymentAddressIdempotencyStore,
	input outport.ClaimPaymentAddressIdempotencyInput,
) error {
	if strings.TrimSpace(input.IdempotencyKey) == "" {
		return nil
	}
	_, err := idempotencyStore.Claim(ctx, input)
	return err
}

func (uc *allocatePaymentAddressUseCase) releaseIdempotencyKeyIfPresent(
	ctx context.Context,
	idempotencyStore outport.PaymentAddressIdempotencyStore,
	input outport.ReleasePaymentAddressIdempotencyInput,
) error {
	if strings.TrimSpace(input.IdempotencyKey) == "" {
		return nil
	}
	return idempotencyStore.Release(ctx, input)
}

func (uc *allocatePaymentAddressUseCase) completeIdempotencyKeyIfPresent(
	ctx context.Context,
	idempotencyStore outport.PaymentAddressIdempotencyStore,
	input outport.CompletePaymentAddressIdempotencyInput,
) error {
	if strings.TrimSpace(input.IdempotencyKey) == "" {
		return nil
	}
	return idempotencyStore.Complete(ctx, input)
}

func (uc *allocatePaymentAddressUseCase) buildAllocatePaymentAddressResponse(
	policy entities.AddressPolicy,
	issuedAllocation entities.PaymentAddressAllocation,
) (dto.AllocatePaymentAddressResponse, error) {
	if issuedAllocation.PaymentAddressID <= 0 {
		return dto.AllocatePaymentAddressResponse{}, errors.New("payment address id must be greater than zero")
	}

	return dto.AllocatePaymentAddressResponse{
		PaymentAddressID:    strconv.FormatInt(issuedAllocation.PaymentAddressID, 10),
		AddressPolicyID:     policy.AddressPolicyID,
		ExpectedAmountMinor: issuedAllocation.ExpectedAmountMinor,
		Chain:               string(issuedAllocation.Chain),
		Network:             string(issuedAllocation.Network),
		Scheme:              issuedAllocation.Scheme,
		MinorUnit:           policy.MinorUnit,
		Decimals:            policy.Decimals,
		Address:             issuedAllocation.Address,
		CustomerReference:   issuedAllocation.CustomerReference,
	}, nil
}

func mapAllocatePaymentAddressIssuancePlanError(err error) error {
	switch {
	case errors.Is(err, entities.ErrAddressPolicyChainMismatch):
		return inport.ErrAddressPolicyNotFound
	case errors.Is(err, entities.ErrAddressPolicyNotEnabled):
		return inport.ErrAddressPolicyNotEnabled
	case errors.Is(err, entities.ErrExpectedAmountMinorInvalid):
		return inport.ErrInvalidExpectedAmount
	default:
		return err
	}
}
