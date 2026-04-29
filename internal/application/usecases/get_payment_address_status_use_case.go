package usecases

import (
	"context"
	"errors"
	"strconv"
	"strings"

	inport "payrune/internal/application/ports/inbound"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

type getPaymentAddressStatusUseCase struct {
	finder       outport.PaymentAddressStatusFinder
	policyReader outport.AddressPolicyReader
}

func NewGetPaymentAddressStatusUseCase(
	finder outport.PaymentAddressStatusFinder,
	policyReader outport.AddressPolicyReader,
) inport.GetPaymentAddressStatusUseCase {
	return &getPaymentAddressStatusUseCase{
		finder:       finder,
		policyReader: policyReader,
	}
}

func (uc *getPaymentAddressStatusUseCase) Execute(
	ctx context.Context,
	input inport.GetPaymentAddressStatusInput,
) (inport.GetPaymentAddressStatusResponse, error) {
	if uc.finder == nil {
		return inport.GetPaymentAddressStatusResponse{}, inport.ErrPaymentAddressStatusFinderNotConfigured
	}
	if uc.policyReader == nil {
		return inport.GetPaymentAddressStatusResponse{}, inport.ErrAddressPolicyReaderNotConfigured
	}
	normalizedChain, ok := valueobjects.ParseSupportedChain(input.Chain)
	if !ok {
		return inport.GetPaymentAddressStatusResponse{}, inport.ErrChainNotSupported
	}
	input.Chain = string(normalizedChain)

	record, found, err := uc.finder.FindByID(ctx, outport.FindPaymentAddressStatusInput{
		Chain:            input.Chain,
		PaymentAddressID: input.PaymentAddressID,
	})
	if err != nil {
		switch {
		case errors.Is(err, outport.ErrPaymentAddressStatusIncomplete):
			return inport.GetPaymentAddressStatusResponse{}, inport.ErrInternalFailure
		case errors.Is(err, outport.ErrPaymentAddressStatusPersistedChainInvalid):
			return inport.GetPaymentAddressStatusResponse{}, inport.ErrInternalFailure
		case errors.Is(err, outport.ErrPaymentAddressStatusPersistedNetworkInvalid):
			return inport.GetPaymentAddressStatusResponse{}, inport.ErrInternalFailure
		case errors.Is(err, outport.ErrPaymentAddressStatusPersistedReceiptStatusInvalid):
			return inport.GetPaymentAddressStatusResponse{}, inport.ErrInternalFailure
		default:
			return inport.GetPaymentAddressStatusResponse{}, inport.ErrDependencyFailure
		}
	}
	if !found {
		return inport.GetPaymentAddressStatusResponse{}, inport.ErrPaymentAddressNotFound
	}

	policy, ok, err := uc.policyReader.FindIssuanceByID(ctx, record.AddressPolicyID)
	if err != nil {
		return inport.GetPaymentAddressStatusResponse{}, inport.ErrDependencyFailure
	}
	if !ok || policy.Chain != input.Chain {
		return inport.GetPaymentAddressStatusResponse{}, inport.ErrPaymentAddressPolicyNotConfigured
	}

	return inport.GetPaymentAddressStatusResponse{
		PaymentAddressID:        strconv.FormatInt(record.PaymentAddressID, 10),
		AddressPolicyID:         record.AddressPolicyID,
		ExpectedAmountMinor:     record.ExpectedAmountMinor,
		Chain:                   record.Chain,
		Network:                 record.Network,
		Scheme:                  record.Scheme,
		AssetReference:          strings.TrimSpace(policy.AssetReference),
		Decimals:                policy.Decimals,
		Address:                 record.Address,
		CustomerReference:       record.CustomerReference,
		PaymentStatus:           record.PaymentStatus,
		ObservedTotalMinor:      record.ObservedTotalMinor,
		ConfirmedTotalMinor:     record.ConfirmedTotalMinor,
		UnconfirmedTotalMinor:   record.UnconfirmedTotalMinor,
		RequiredConfirmations:   record.RequiredConfirmations,
		LastObservedBlockHeight: record.LastObservedBlockHeight,
		IssuedAt:                record.IssuedAt,
		FirstObservedAt:         record.FirstObservedAt,
		PaidAt:                  record.PaidAt,
		ConfirmedAt:             record.ConfirmedAt,
		ExpiresAt:               record.ExpiresAt,
		LastError:               outport.PaymentReceiptTrackingFailureReasonMessage(record.LastFailureReason),
	}, nil
}
