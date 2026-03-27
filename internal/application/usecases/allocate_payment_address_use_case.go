package usecases

import (
	"context"
	"errors"
	"strconv"
	"time"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/policies"
)

type allocatePaymentAddressUseCase struct {
	unitOfWork           outport.UnitOfWork
	issuedAddressDeriver outport.IssuedPaymentAddressDeriver
	policyReader         outport.AddressPolicyReader
	issuancePolicy       policies.PaymentAddressAllocationIssuancePolicy
	clock                outport.Clock
}

func NewAllocatePaymentAddressUseCase(
	unitOfWork outport.UnitOfWork,
	issuedAddressDeriver outport.IssuedPaymentAddressDeriver,
	policyReader outport.AddressPolicyReader,
	issuancePolicy policies.PaymentAddressAllocationIssuancePolicy,
	clock outport.Clock,
) inport.AllocatePaymentAddressUseCase {
	return &allocatePaymentAddressUseCase{
		unitOfWork:           unitOfWork,
		issuedAddressDeriver: issuedAddressDeriver,
		policyReader:         policyReader,
		issuancePolicy:       issuancePolicy,
		clock:                clock,
	}
}

func (uc *allocatePaymentAddressUseCase) Execute(
	ctx context.Context,
	input dto.AllocatePaymentAddressInput,
) (dto.AllocatePaymentAddressResponse, error) {
	if uc.unitOfWork == nil {
		return dto.AllocatePaymentAddressResponse{}, errors.New("unit of work is not configured")
	}
	if uc.issuedAddressDeriver == nil {
		return dto.AllocatePaymentAddressResponse{}, errors.New("issued payment address deriver is not configured")
	}
	if uc.policyReader == nil {
		return dto.AllocatePaymentAddressResponse{}, errors.New("address policy reader is not configured")
	}
	if uc.clock == nil {
		return dto.AllocatePaymentAddressResponse{}, errors.New("clock is not configured")
	}
	if !uc.issuedAddressDeriver.SupportsChain(input.Chain) {
		return dto.AllocatePaymentAddressResponse{}, inport.ErrChainNotSupported
	}

	response, found, err := uc.loadReplayedAllocationResponse(ctx, input)
	if err != nil {
		return dto.AllocatePaymentAddressResponse{}, err
	}
	if found {
		return response, nil
	}

	policy, ok, err := uc.policyReader.FindIssuanceByID(ctx, input.AddressPolicyID)
	if err != nil {
		return dto.AllocatePaymentAddressResponse{}, err
	}
	if !ok {
		return dto.AllocatePaymentAddressResponse{}, inport.ErrAddressPolicyNotFound
	}

	issuedAt := uc.clock.NowUTC()
	issuancePlan, err := uc.issuancePolicy.Plan(
		policy,
		input.Chain,
		input.ExpectedAmountMinor,
		input.CustomerReference,
		issuedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, entities.ErrAddressPolicyChainMismatch):
			return dto.AllocatePaymentAddressResponse{}, inport.ErrAddressPolicyNotFound
		case errors.Is(err, entities.ErrAddressPolicyNotEnabled):
			return dto.AllocatePaymentAddressResponse{}, inport.ErrAddressPolicyNotEnabled
		case errors.Is(err, entities.ErrExpectedAmountMinorInvalid):
			return dto.AllocatePaymentAddressResponse{}, inport.ErrInvalidExpectedAmount
		default:
			return dto.AllocatePaymentAddressResponse{}, err
		}
	}

	issuedAllocation, err := uc.issueAllocation(ctx, input, issuancePlan, issuedAt)
	if err != nil {
		if errors.Is(err, outport.ErrPaymentAddressIdempotencyKeyExists) {
			response, found, lookupErr := uc.loadReplayedAllocationResponse(ctx, input)
			if lookupErr != nil {
				return dto.AllocatePaymentAddressResponse{}, lookupErr
			}
			if found {
				return response, nil
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

	if issuedAllocation.PaymentAddressID <= 0 {
		return dto.AllocatePaymentAddressResponse{}, errors.New("payment address id must be greater than zero")
	}

	return dto.AllocatePaymentAddressResponse{
		PaymentAddressID:    strconv.FormatInt(issuedAllocation.PaymentAddressID, 10),
		AddressPolicyID:     policy.AddressPolicy.AddressPolicyID,
		ExpectedAmountMinor: issuedAllocation.ExpectedAmountMinor,
		Chain:               string(issuedAllocation.Chain),
		Network:             string(issuedAllocation.Network),
		Scheme:              issuedAllocation.Scheme,
		MinorUnit:           policy.AddressPolicy.MinorUnit,
		Decimals:            policy.AddressPolicy.Decimals,
		Address:             issuedAllocation.Address,
		CustomerReference:   issuedAllocation.CustomerReference,
	}, nil
}

func (uc *allocatePaymentAddressUseCase) loadReplayedAllocationResponse(
	ctx context.Context,
	input dto.AllocatePaymentAddressInput,
) (dto.AllocatePaymentAddressResponse, bool, error) {
	if input.IdempotencyKey == "" {
		return dto.AllocatePaymentAddressResponse{}, false, nil
	}

	var allocation entities.PaymentAddressAllocation
	found := false

	err := uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
		allocationStore := txScope.PaymentAddressAllocation
		idempotencyStore := txScope.PaymentAddressIdempotency
		if allocationStore == nil {
			return errors.New("payment address allocation store is not configured")
		}
		if idempotencyStore == nil {
			return errors.New("payment address idempotency store is not configured")
		}

		record, recordFound, err := idempotencyStore.FindByKey(
			ctx,
			outport.FindPaymentAddressIdempotencyInput{
				Chain:          input.Chain,
				IdempotencyKey: input.IdempotencyKey,
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
		if record.AddressPolicyID != input.AddressPolicyID ||
			record.ExpectedAmountMinor != input.ExpectedAmountMinor ||
			record.CustomerReference != input.CustomerReference {
			return inport.ErrIdempotencyKeyConflict
		}

		existingAllocation, allocationFound, err := allocationStore.FindIssuedByID(
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
		return dto.AllocatePaymentAddressResponse{}, false, err
	}
	if !found {
		return dto.AllocatePaymentAddressResponse{}, false, nil
	}

	policy, ok, err := uc.policyReader.FindIssuanceByID(ctx, allocation.AddressPolicyID)
	if err != nil {
		return dto.AllocatePaymentAddressResponse{}, false, err
	}
	if !ok {
		return dto.AllocatePaymentAddressResponse{}, false, inport.ErrAddressPolicyNotFound
	}
	if allocation.PaymentAddressID <= 0 {
		return dto.AllocatePaymentAddressResponse{}, false, errors.New("payment address id must be greater than zero")
	}

	return dto.AllocatePaymentAddressResponse{
		PaymentAddressID:    strconv.FormatInt(allocation.PaymentAddressID, 10),
		AddressPolicyID:     policy.AddressPolicy.AddressPolicyID,
		ExpectedAmountMinor: allocation.ExpectedAmountMinor,
		Chain:               string(allocation.Chain),
		Network:             string(allocation.Network),
		Scheme:              allocation.Scheme,
		MinorUnit:           policy.AddressPolicy.MinorUnit,
		Decimals:            policy.AddressPolicy.Decimals,
		Address:             allocation.Address,
		CustomerReference:   allocation.CustomerReference,
		IdempotencyReplayed: true,
	}, true, nil
}

// Keep reservation, issued-allocation state, and receipt tracking creation in
// one transaction so the address index state cannot diverge from the receipt
// tracking state.
func (uc *allocatePaymentAddressUseCase) issueAllocation(
	ctx context.Context,
	input dto.AllocatePaymentAddressInput,
	issuancePlan policies.PaymentAddressAllocationIssuancePlan,
	issuedAt time.Time,
) (entities.PaymentAddressAllocation, error) {
	idempotencyKey := input.IdempotencyKey
	reserveInput := outport.ReservePaymentAddressAllocationInput{
		IssuancePolicy:      issuancePlan.Reservation.IssuancePolicy,
		ExpectedAmountMinor: issuancePlan.Reservation.ExpectedAmountMinor,
		CustomerReference:   issuancePlan.Reservation.CustomerReference,
	}

	var issuedAllocation entities.PaymentAddressAllocation
	var derivationFailureErr error

	err := uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
		allocationStore := txScope.PaymentAddressAllocation
		idempotencyStore := txScope.PaymentAddressIdempotency
		receiptTrackingStore := txScope.PaymentReceiptTracking
		if allocationStore == nil {
			return errors.New("payment address allocation store is not configured")
		}
		if idempotencyStore == nil {
			return errors.New("payment address idempotency store is not configured")
		}
		if receiptTrackingStore == nil {
			return errors.New("payment receipt tracking store is not configured")
		}

		if idempotencyKey != "" {
			_, err := idempotencyStore.Claim(ctx, outport.ClaimPaymentAddressIdempotencyInput{
				Chain:               input.Chain,
				IdempotencyKey:      idempotencyKey,
				AddressPolicyID:     issuancePlan.Reservation.IssuancePolicy.AddressPolicy.AddressPolicyID,
				ExpectedAmountMinor: issuancePlan.Reservation.ExpectedAmountMinor,
				CustomerReference:   issuancePlan.Reservation.CustomerReference,
			})
			if err != nil {
				return err
			}
		}

		allocation, err := reserveAllocation(ctx, allocationStore, reserveInput, issuancePlan.ReservationAttempts)
		if err != nil {
			return err
		}

		derived, err := uc.issuedAddressDeriver.DeriveIssuedAddress(ctx, outport.DeriveIssuedPaymentAddressInput{
			Policy:     issuancePlan.Reservation.IssuancePolicy,
			Allocation: allocation,
		})
		if err != nil {
			derivationFailureErr = err
			return persistDerivationFailure(ctx, allocationStore, idempotencyStore, input, allocation, err)
		}

		issuedAllocation, err = allocation.MarkIssued(
			issuancePlan.Reservation.IssuancePolicy,
			derived.Address,
			derived.AddressReference,
		)
		if err != nil {
			derivationFailureErr = err
			return persistDerivationFailure(ctx, allocationStore, idempotencyStore, input, allocation, err)
		}

		if err := allocationStore.Complete(ctx, issuedAllocation, issuedAt); err != nil {
			return err
		}

		receiptTracking, err := issuedAllocation.IssueReceiptTracking(
			issuedAt,
			issuancePlan.ReceiptTerms.RequiredConfirmations,
			issuancePlan.ReceiptTerms.ExpiresAt,
		)
		if err != nil {
			return err
		}
		if err := receiptTrackingStore.Create(ctx, receiptTracking, issuedAt); err != nil {
			return err
		}
		if idempotencyKey != "" {
			if err := idempotencyStore.Complete(ctx, outport.CompletePaymentAddressIdempotencyInput{
				Chain:            input.Chain,
				IdempotencyKey:   idempotencyKey,
				PaymentAddressID: issuedAllocation.PaymentAddressID,
			}); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return entities.PaymentAddressAllocation{}, err
	}
	if derivationFailureErr != nil {
		return entities.PaymentAddressAllocation{}, derivationFailureErr
	}
	return issuedAllocation, nil
}

func reserveAllocation(
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

func persistDerivationFailure(
	ctx context.Context,
	allocationStore outport.PaymentAddressAllocationStore,
	idempotencyStore outport.PaymentAddressIdempotencyStore,
	input dto.AllocatePaymentAddressInput,
	allocation entities.PaymentAddressAllocation,
	deriveErr error,
) error {
	failedAllocation, err := allocation.MarkDerivationFailed(deriveErr.Error())
	if err != nil {
		return err
	}
	if err := allocationStore.MarkDerivationFailed(ctx, failedAllocation); err != nil {
		return err
	}
	if input.IdempotencyKey == "" {
		return nil
	}
	return idempotencyStore.Release(ctx, outport.ReleasePaymentAddressIdempotencyInput{
		Chain:          input.Chain,
		IdempotencyKey: input.IdempotencyKey,
	})
}
