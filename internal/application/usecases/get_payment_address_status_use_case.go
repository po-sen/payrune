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
	finder outport.PaymentAddressStatusFinder
}

func NewGetPaymentAddressStatusUseCase(
	finder outport.PaymentAddressStatusFinder,
) inport.GetPaymentAddressStatusUseCase {
	return &getPaymentAddressStatusUseCase{
		finder: finder,
	}
}

func (uc *getPaymentAddressStatusUseCase) Execute(
	ctx context.Context,
	input dto.GetPaymentAddressStatusInput,
) (dto.GetPaymentAddressStatusResponse, error) {
	if uc.finder == nil {
		return dto.GetPaymentAddressStatusResponse{}, errors.New("payment address status finder is not configured")
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

	return dto.GetPaymentAddressStatusResponse{
		PaymentAddressID:        strconv.FormatInt(record.PaymentAddressID, 10),
		AddressPolicyID:         record.AddressPolicyID,
		ExpectedAmountMinor:     record.ExpectedAmountMinor,
		Chain:                   string(record.Chain),
		Network:                 string(record.Network),
		Scheme:                  record.Scheme,
		AssetCode:               record.AssetCode,
		AssetType:               record.AssetType,
		TokenAddress:            record.TokenAddress,
		MinorUnit:               record.MinorUnit,
		Decimals:                record.Decimals,
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
