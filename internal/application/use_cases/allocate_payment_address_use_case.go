package use_cases

import (
	"context"
	"errors"
	"strconv"
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
	// businessErr is returned after commit when derivation failed but the failure
	// state was persisted successfully.
	businessErr error
}

type allocatePaymentAddressTxScope struct {
	allocationStore      outport.PaymentAddressAllocationStore
	receiptTrackingStore outport.PaymentReceiptTrackingStore
}

type allocatePaymentAddressDerivationOutcome struct {
	issuedAllocation entities.PaymentAddressAllocation
	// businessErr is returned after commit when derivation failed but the failure
	// state was persisted successfully.
	businessErr error
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

	issuedAt := uc.clock.NowUTC()
	issuancePlan, err := uc.loadIssuancePlan(ctx, input, issuedAt)
	if err != nil {
		return dto.AllocatePaymentAddressResponse{}, err
	}
	txOutcome, err := uc.executeIssuanceTransaction(ctx, issuancePlan, issuedAt)
	if err != nil {
		if errors.Is(err, outport.ErrAddressIndexExhausted) {
			return dto.AllocatePaymentAddressResponse{}, inport.ErrAddressPoolExhausted
		}
		return dto.AllocatePaymentAddressResponse{}, err
	}
	if txOutcome.businessErr != nil {
		return dto.AllocatePaymentAddressResponse{}, txOutcome.businessErr
	}

	return uc.buildAllocatePaymentAddressResponse(
		issuancePlan,
		txOutcome.allocation,
		txOutcome.issuedAllocation,
	)
}

func (uc *allocatePaymentAddressUseCase) loadIssuancePlan(
	ctx context.Context,
	input dto.AllocatePaymentAddressInput,
	issuedAt time.Time,
) (policies.PaymentAddressAllocationIssuancePlan, error) {
	policy, ok, err := uc.policyReader.FindIssuanceByID(ctx, input.AddressPolicyID)
	if err != nil {
		return policies.PaymentAddressAllocationIssuancePlan{}, err
	}
	if !ok {
		return policies.PaymentAddressAllocationIssuancePlan{}, inport.ErrAddressPolicyNotFound
	}

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
	issuancePlan policies.PaymentAddressAllocationIssuancePlan,
	issuedAt time.Time,
) (allocatePaymentAddressTxOutcome, error) {
	var outcome allocatePaymentAddressTxOutcome

	reserveInput := outport.ReservePaymentAddressAllocationInput{
		IssuancePolicy:      issuancePlan.Reservation.IssuancePolicy,
		ExpectedAmountMinor: issuancePlan.Reservation.ExpectedAmountMinor,
		CustomerReference:   issuancePlan.Reservation.CustomerReference,
	}

	err := uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
		stores, err := requireAllocatePaymentAddressTxScope(txScope)
		if err != nil {
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
		if derivationOutcome.businessErr != nil {
			outcome.businessErr = derivationOutcome.businessErr
			return nil
		}
		outcome.issuedAllocation = derivationOutcome.issuedAllocation

		return uc.persistIssuedAllocation(
			ctx,
			stores,
			outcome.issuedAllocation,
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
		return allocatePaymentAddressDerivationOutcome{businessErr: err}, nil
	}

	issuedAllocation, err := allocation.MarkIssued(policy, output.Address, output.RelativeDerivationPath)
	if err != nil {
		if persistErr := uc.persistDerivationFailure(ctx, allocationStore, allocation, err); persistErr != nil {
			return allocatePaymentAddressDerivationOutcome{}, persistErr
		}
		return allocatePaymentAddressDerivationOutcome{businessErr: err}, nil
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

	return stores.receiptTrackingStore.Create(ctx, receiptTracking, issuedAt)
}

func requireAllocatePaymentAddressTxScope(
	txScope outport.TxScope,
) (allocatePaymentAddressTxScope, error) {
	if txScope.PaymentAddressAllocation == nil {
		return allocatePaymentAddressTxScope{}, errors.New("payment address allocation store is not configured")
	}
	if txScope.PaymentReceiptTracking == nil {
		return allocatePaymentAddressTxScope{}, errors.New("payment receipt tracking store is not configured")
	}

	return allocatePaymentAddressTxScope{
		allocationStore:      txScope.PaymentAddressAllocation,
		receiptTrackingStore: txScope.PaymentReceiptTracking,
	}, nil
}

func (uc *allocatePaymentAddressUseCase) buildAllocatePaymentAddressResponse(
	issuancePlan policies.PaymentAddressAllocationIssuancePlan,
	allocation entities.PaymentAddressAllocation,
	issuedAllocation entities.PaymentAddressAllocation,
) (dto.AllocatePaymentAddressResponse, error) {
	if allocation.PaymentAddressID <= 0 {
		return dto.AllocatePaymentAddressResponse{}, errors.New("payment address id must be greater than zero")
	}

	return dto.AllocatePaymentAddressResponse{
		PaymentAddressID:    strconv.FormatInt(allocation.PaymentAddressID, 10),
		AddressPolicyID:     issuancePlan.Reservation.IssuancePolicy.AddressPolicy.AddressPolicyID,
		ExpectedAmountMinor: issuancePlan.Reservation.ExpectedAmountMinor,
		Chain:               string(issuancePlan.Reservation.IssuancePolicy.AddressPolicy.Chain),
		Network:             string(issuancePlan.Reservation.IssuancePolicy.AddressPolicy.Network),
		Scheme:              string(issuancePlan.Reservation.IssuancePolicy.AddressPolicy.Scheme),
		MinorUnit:           issuancePlan.Reservation.IssuancePolicy.AddressPolicy.MinorUnit,
		Decimals:            issuancePlan.Reservation.IssuancePolicy.AddressPolicy.Decimals,
		Address:             issuedAllocation.Address,
		CustomerReference:   issuancePlan.Reservation.CustomerReference,
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
