package usecases

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	inport "payrune/internal/application/ports/inbound"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
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
	input inport.AllocatePaymentAddressInput,
) (inport.AllocatePaymentAddressResponse, error) {
	if uc.unitOfWork == nil {
		return inport.AllocatePaymentAddressResponse{}, inport.ErrUnitOfWorkNotConfigured
	}
	if uc.issuedAddressDeriver == nil {
		return inport.AllocatePaymentAddressResponse{}, inport.ErrIssuedPaymentAddressDeriverNotConfigured
	}
	if uc.policyReader == nil {
		return inport.AllocatePaymentAddressResponse{}, inport.ErrAddressPolicyReaderNotConfigured
	}
	if uc.clock == nil {
		return inport.AllocatePaymentAddressResponse{}, inport.ErrClockNotConfigured
	}
	requestedChain, ok := valueobjects.ParseSupportedChain(input.Chain)
	if !ok {
		return inport.AllocatePaymentAddressResponse{}, inport.ErrChainNotSupported
	}
	input.Chain = string(requestedChain)
	if !uc.issuedAddressDeriver.SupportsChain(string(requestedChain)) {
		return inport.AllocatePaymentAddressResponse{}, inport.ErrChainNotSupported
	}

	addressPolicyID, err := valueobjects.NewAddressPolicyID(input.AddressPolicyID)
	if err != nil {
		return inport.AllocatePaymentAddressResponse{}, inport.ErrInvalidAddressPolicyID
	}

	response, found, err := uc.loadReplayedAllocationResponse(ctx, input, addressPolicyID)
	if err != nil {
		return inport.AllocatePaymentAddressResponse{}, err
	}
	if found {
		return response, nil
	}

	policyRecord, ok, err := uc.policyReader.FindIssuanceByID(ctx, string(addressPolicyID))
	if err != nil {
		return inport.AllocatePaymentAddressResponse{}, inport.ErrDependencyFailure
	}
	if !ok {
		return inport.AllocatePaymentAddressResponse{}, inport.ErrAddressPolicyNotFound
	}
	policy, err := addressIssuancePolicyFromRecord(policyRecord)
	if err != nil {
		return inport.AllocatePaymentAddressResponse{}, inport.ErrInternalFailure
	}

	issuedAt := uc.clock.NowUTC()
	issuancePlan, err := uc.issuancePolicy.Plan(
		policy,
		requestedChain,
		input.ExpectedAmountMinor,
		input.CustomerReference,
		issuedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, policies.ErrAddressPolicyChainMismatch):
			return inport.AllocatePaymentAddressResponse{}, inport.ErrAddressPolicyNotFound
		case errors.Is(err, policies.ErrAddressPolicyNotEnabled):
			return inport.AllocatePaymentAddressResponse{}, inport.ErrAddressPolicyNotEnabled
		case errors.Is(err, policies.ErrExpectedAmountMinorInvalid):
			return inport.AllocatePaymentAddressResponse{}, inport.ErrInvalidExpectedAmount
		default:
			return inport.AllocatePaymentAddressResponse{}, inport.ErrInternalFailure
		}
	}
	issuedAllocation, err := uc.issueAllocation(ctx, input, issuancePlan, issuedAt)
	if err != nil {
		if errors.Is(err, outport.ErrPaymentAddressIdempotencyKeyExists) {
			response, found, lookupErr := uc.loadReplayedAllocationResponse(ctx, input, addressPolicyID)
			if lookupErr != nil {
				return inport.AllocatePaymentAddressResponse{}, lookupErr
			}
			if found {
				return response, nil
			}
			return inport.AllocatePaymentAddressResponse{}, inport.ErrIdempotencyClaimConflictWithoutCompletedRecord
		}
		if errors.Is(err, outport.ErrAddressIndexExhausted) {
			return inport.AllocatePaymentAddressResponse{}, inport.ErrAddressPoolExhausted
		}
		return inport.AllocatePaymentAddressResponse{}, err
	}

	if issuedAllocation.PaymentAddressID <= 0 {
		return inport.AllocatePaymentAddressResponse{}, inport.ErrPaymentAddressIDMustBeGreaterThanZero
	}

	return inport.AllocatePaymentAddressResponse{
		PaymentAddressID:    strconv.FormatInt(issuedAllocation.PaymentAddressID, 10),
		AddressPolicyID:     string(policy.AddressPolicyID),
		ExpectedAmountMinor: issuedAllocation.ExpectedAmountMinor,
		Chain:               string(issuedAllocation.Chain),
		Network:             string(issuedAllocation.Network),
		Scheme:              string(issuedAllocation.Scheme),
		AssetReference:      strings.TrimSpace(policy.AssetReference),
		Decimals:            policy.Decimals,
		Address:             issuedAllocation.Address,
		CustomerReference:   issuedAllocation.CustomerReference,
	}, nil
}

func (uc *allocatePaymentAddressUseCase) loadReplayedAllocationResponse(
	ctx context.Context,
	input inport.AllocatePaymentAddressInput,
	addressPolicyID valueobjects.AddressPolicyID,
) (inport.AllocatePaymentAddressResponse, bool, error) {
	if input.IdempotencyKey == "" {
		return inport.AllocatePaymentAddressResponse{}, false, nil
	}

	var allocation entities.PaymentAddressAllocation
	found := false

	err := uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
		allocationStore := txScope.PaymentAddressAllocation
		idempotencyStore := txScope.PaymentAddressIdempotency
		if allocationStore == nil {
			return inport.ErrPaymentAddressAllocationStoreNotConfigured
		}
		if idempotencyStore == nil {
			return inport.ErrPaymentAddressIdempotencyStoreNotConfigured
		}

		record, recordFound, err := idempotencyStore.FindByKey(
			ctx,
			outport.FindPaymentAddressIdempotencyInput{
				Chain:          input.Chain,
				IdempotencyKey: input.IdempotencyKey,
			},
		)
		if err != nil {
			return inport.ErrDependencyFailure
		}
		if !recordFound {
			return nil
		}
		if record.PaymentAddressID <= 0 {
			return inport.ErrPaymentAddressIdempotencyRecordIncomplete
		}
		if record.AddressPolicyID != string(addressPolicyID) ||
			record.ExpectedAmountMinor != input.ExpectedAmountMinor ||
			record.CustomerReference != input.CustomerReference {
			return inport.ErrIdempotencyKeyConflict
		}

		allocationRecord, allocationFound, err := allocationStore.FindIssuedByID(
			ctx,
			outport.FindIssuedPaymentAddressAllocationByIDInput{PaymentAddressID: record.PaymentAddressID},
		)
		if err != nil {
			return inport.ErrDependencyFailure
		}
		if !allocationFound {
			return inport.ErrCompletedIdempotencyRecordMissingIssuedAllocation
		}
		existingAllocation, err := paymentAddressAllocationFromRecord(allocationRecord)
		if err != nil {
			return inport.ErrInternalFailure
		}

		allocation = existingAllocation
		found = true
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, inport.ErrPaymentAddressAllocationStoreNotConfigured):
			return inport.AllocatePaymentAddressResponse{}, false, err
		case errors.Is(err, inport.ErrPaymentAddressIdempotencyStoreNotConfigured):
			return inport.AllocatePaymentAddressResponse{}, false, err
		case errors.Is(err, inport.ErrPaymentAddressIdempotencyRecordIncomplete):
			return inport.AllocatePaymentAddressResponse{}, false, err
		case errors.Is(err, inport.ErrIdempotencyKeyConflict):
			return inport.AllocatePaymentAddressResponse{}, false, err
		case errors.Is(err, inport.ErrCompletedIdempotencyRecordMissingIssuedAllocation):
			return inport.AllocatePaymentAddressResponse{}, false, err
		case errors.Is(err, inport.ErrInternalFailure):
			return inport.AllocatePaymentAddressResponse{}, false, err
		default:
			return inport.AllocatePaymentAddressResponse{}, false, inport.ErrDependencyFailure
		}
	}
	if !found {
		return inport.AllocatePaymentAddressResponse{}, false, nil
	}

	policyRecord, ok, err := uc.policyReader.FindIssuanceByID(ctx, string(allocation.AddressPolicyID))
	if err != nil {
		return inport.AllocatePaymentAddressResponse{}, false, inport.ErrDependencyFailure
	}
	if !ok {
		return inport.AllocatePaymentAddressResponse{}, false, inport.ErrAddressPolicyNotFound
	}
	policy, err := addressIssuancePolicyFromRecord(policyRecord)
	if err != nil {
		return inport.AllocatePaymentAddressResponse{}, false, inport.ErrInternalFailure
	}
	if allocation.PaymentAddressID <= 0 {
		return inport.AllocatePaymentAddressResponse{}, false, inport.ErrPaymentAddressIDMustBeGreaterThanZero
	}

	return inport.AllocatePaymentAddressResponse{
		PaymentAddressID:    strconv.FormatInt(allocation.PaymentAddressID, 10),
		AddressPolicyID:     string(policy.AddressPolicyID),
		ExpectedAmountMinor: allocation.ExpectedAmountMinor,
		Chain:               string(allocation.Chain),
		Network:             string(allocation.Network),
		Scheme:              string(allocation.Scheme),
		AssetReference:      strings.TrimSpace(policy.AssetReference),
		Decimals:            policy.Decimals,
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
	input inport.AllocatePaymentAddressInput,
	issuancePlan policies.PaymentAddressAllocationIssuancePlan,
	issuedAt time.Time,
) (entities.PaymentAddressAllocation, error) {
	idempotencyKey := input.IdempotencyKey
	reserveInput := outport.ReservePaymentAddressAllocationInput{
		IssuancePolicy:      addressIssuancePolicyRecordFromDomain(issuancePlan.Reservation.IssuancePolicy),
		ExpectedAmountMinor: issuancePlan.Reservation.ExpectedAmountMinor,
		CustomerReference:   issuancePlan.Reservation.CustomerReference,
	}

	var issuedAllocation entities.PaymentAddressAllocation

	err := uc.unitOfWork.WithinTransaction(ctx, func(txScope outport.TxScope) error {
		allocationStore := txScope.PaymentAddressAllocation
		idempotencyStore := txScope.PaymentAddressIdempotency
		receiptTrackingStore := txScope.PaymentReceiptTracking
		if allocationStore == nil {
			return inport.ErrPaymentAddressAllocationStoreNotConfigured
		}
		if idempotencyStore == nil {
			return inport.ErrPaymentAddressIdempotencyStoreNotConfigured
		}
		if receiptTrackingStore == nil {
			return inport.ErrPaymentReceiptTrackingStoreNotConfigured
		}

		if idempotencyKey != "" {
			_, err := idempotencyStore.Claim(ctx, outport.ClaimPaymentAddressIdempotencyInput{
				Chain:               input.Chain,
				IdempotencyKey:      idempotencyKey,
				AddressPolicyID:     string(issuancePlan.Reservation.IssuancePolicy.AddressPolicyID),
				ExpectedAmountMinor: issuancePlan.Reservation.ExpectedAmountMinor,
				CustomerReference:   issuancePlan.Reservation.CustomerReference,
			})
			if err != nil {
				if errors.Is(err, outport.ErrPaymentAddressIdempotencyKeyExists) {
					return err
				}
				return inport.ErrDependencyFailure
			}
		}

		allocation, err := reserveAllocation(ctx, allocationStore, reserveInput, issuancePlan.ReservationAttempts)
		if err != nil {
			return err
		}

		derived, err := uc.issuedAddressDeriver.DeriveIssuedAddress(ctx, outport.DeriveIssuedPaymentAddressInput{
			Policy:     addressIssuancePolicyRecordFromDomain(issuancePlan.Reservation.IssuancePolicy),
			Allocation: paymentAddressAllocationRecordFromDomain(allocation),
		})
		if err != nil {
			return handleDerivationFailure(
				ctx,
				allocationStore,
				idempotencyStore,
				input,
				allocation,
				inport.ErrDependencyFailure,
			)
		}

		assetReference := strings.TrimSpace(issuancePlan.Reservation.IssuancePolicy.Normalize().AssetReference)

		issuedAllocation, err = allocation.MarkIssued(
			issuancePlan.Reservation.IssuancePolicy.AddressPolicyID,
			issuancePlan.Reservation.IssuancePolicy.Chain,
			issuancePlan.Reservation.IssuancePolicy.Network,
			issuancePlan.Reservation.IssuancePolicy.Scheme,
			assetReference,
			derived.Address,
		)
		if err != nil {
			return handleDerivationFailure(
				ctx,
				allocationStore,
				idempotencyStore,
				input,
				allocation,
				inport.ErrInternalFailure,
			)
		}

		if err := allocationStore.Complete(ctx, outport.CompletePaymentAddressAllocationInput{
			Allocation:    paymentAddressAllocationRecordFromDomain(issuedAllocation),
			SweepMaterial: derived.SweepMaterial,
			IssuedAt:      issuedAt,
		}); err != nil {
			return inport.ErrDependencyFailure
		}

		receiptTracking, err := issuedAllocation.IssueReceiptTracking(
			issuedAt,
			issuancePlan.ReceiptTerms.RequiredConfirmations,
			issuancePlan.ReceiptTerms.ExpiresAt,
		)
		if err != nil {
			return inport.ErrInternalFailure
		}
		if err := receiptTrackingStore.Create(ctx, paymentReceiptTrackingRecordFromDomain(receiptTracking), issuedAt); err != nil {
			return inport.ErrDependencyFailure
		}
		if idempotencyKey != "" {
			if err := idempotencyStore.Complete(ctx, outport.CompletePaymentAddressIdempotencyInput{
				Chain:            input.Chain,
				IdempotencyKey:   idempotencyKey,
				PaymentAddressID: issuedAllocation.PaymentAddressID,
			}); err != nil {
				return inport.ErrDependencyFailure
			}
		}

		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, outport.ErrPaymentAddressIdempotencyKeyExists):
			return entities.PaymentAddressAllocation{}, err
		case errors.Is(err, outport.ErrAddressIndexExhausted):
			return entities.PaymentAddressAllocation{}, err
		case errors.Is(err, inport.ErrPaymentAddressAllocationStoreNotConfigured):
			return entities.PaymentAddressAllocation{}, err
		case errors.Is(err, inport.ErrPaymentAddressIdempotencyStoreNotConfigured):
			return entities.PaymentAddressAllocation{}, err
		case errors.Is(err, inport.ErrPaymentReceiptTrackingStoreNotConfigured):
			return entities.PaymentAddressAllocation{}, err
		case errors.Is(err, inport.ErrPaymentAddressAllocationReservationAttemptInvalid):
			return entities.PaymentAddressAllocation{}, err
		case errors.Is(err, inport.ErrPaymentAddressAllocationReservationAttemptsRequired):
			return entities.PaymentAddressAllocation{}, err
		case errors.Is(err, inport.ErrDependencyFailure):
			return entities.PaymentAddressAllocation{}, err
		case errors.Is(err, inport.ErrInternalFailure):
			return entities.PaymentAddressAllocation{}, err
		default:
			return entities.PaymentAddressAllocation{}, inport.ErrDependencyFailure
		}
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
			reopenedRecord, reopened, err := allocationStore.ReopenFailedReservation(ctx, reserveInput)
			if err != nil {
				return entities.PaymentAddressAllocation{}, inport.ErrDependencyFailure
			}
			if reopened {
				reopenedAllocation, err := paymentAddressAllocationFromRecord(reopenedRecord)
				if err != nil {
					return entities.PaymentAddressAllocation{}, inport.ErrInternalFailure
				}
				return reopenedAllocation, nil
			}
		case policies.PaymentAddressAllocationReservationAttemptReserveFresh:
			allocationRecord, err := allocationStore.ReserveFresh(ctx, reserveInput)
			if err != nil {
				if errors.Is(err, outport.ErrAddressIndexExhausted) {
					return entities.PaymentAddressAllocation{}, err
				}
				return entities.PaymentAddressAllocation{}, inport.ErrDependencyFailure
			}
			allocation, err := paymentAddressAllocationFromRecord(allocationRecord)
			if err != nil {
				return entities.PaymentAddressAllocation{}, inport.ErrInternalFailure
			}
			return allocation, nil
		default:
			return entities.PaymentAddressAllocation{}, inport.ErrPaymentAddressAllocationReservationAttemptInvalid
		}
	}

	return entities.PaymentAddressAllocation{}, inport.ErrPaymentAddressAllocationReservationAttemptsRequired
}

func handleDerivationFailure(
	ctx context.Context,
	allocationStore outport.PaymentAddressAllocationStore,
	idempotencyStore outport.PaymentAddressIdempotencyStore,
	input inport.AllocatePaymentAddressInput,
	allocation entities.PaymentAddressAllocation,
	finalErr error,
) error {
	failedAllocation, err := allocation.MarkDerivationFailed(
		valueobjects.PaymentAddressAllocationDerivationFailureReasonDerivationFailed,
	)
	if err != nil {
		return inport.ErrInternalFailure
	}
	if err := allocationStore.MarkDerivationFailed(ctx, paymentAddressAllocationRecordFromDomain(failedAllocation)); err != nil {
		return inport.ErrDependencyFailure
	}
	if input.IdempotencyKey == "" {
		return finalErr
	}
	if err := idempotencyStore.Release(ctx, outport.ReleasePaymentAddressIdempotencyInput{
		Chain:          input.Chain,
		IdempotencyKey: input.IdempotencyKey,
	}); err != nil {
		return inport.ErrDependencyFailure
	}
	return finalErr
}
