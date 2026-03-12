package usecases

import (
	"context"
	"errors"
	"strconv"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
	outport "payrune/internal/application/ports/outbound"
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
	input dto.GetPaymentAddressStatusInput,
) (dto.GetPaymentAddressStatusResponse, error) {
	if uc.finder == nil {
		return dto.GetPaymentAddressStatusResponse{}, errors.New("payment address status finder is not configured")
	}
	if uc.policyReader == nil {
		return dto.GetPaymentAddressStatusResponse{}, errors.New("address policy reader is not configured")
	}

	record, found, err := uc.finder.FindByID(ctx, outport.FindPaymentAddressStatusInput{
		Chain:            input.Chain,
		PaymentAddressID: input.PaymentAddressID,
	})
	if err != nil {
		return dto.GetPaymentAddressStatusResponse{}, err
	}
	if !found {
		return dto.GetPaymentAddressStatusResponse{}, inport.ErrPaymentAddressNotFound
	}

	policy, ok, err := uc.policyReader.FindIssuanceByID(ctx, record.AddressPolicyID)
	if err != nil {
		return dto.GetPaymentAddressStatusResponse{}, err
	}
	if !ok || policy.AddressPolicy.Chain != input.Chain {
		return dto.GetPaymentAddressStatusResponse{}, errors.New("payment address policy is not configured")
	}

	return dto.GetPaymentAddressStatusResponse{
		PaymentAddressID:        strconv.FormatInt(record.PaymentAddressID, 10),
		AddressPolicyID:         record.AddressPolicyID,
		ExpectedAmountMinor:     record.ExpectedAmountMinor,
		Chain:                   string(record.Chain),
		Network:                 string(record.Network),
		Scheme:                  record.Scheme,
		MinorUnit:               policy.AddressPolicy.MinorUnit,
		Decimals:                policy.AddressPolicy.Decimals,
		Address:                 record.Address,
		CustomerReference:       record.CustomerReference,
		PaymentStatus:           string(record.PaymentStatus),
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
		LastError:               record.LastError,
	}, nil
}
