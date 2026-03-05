package use_cases

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/in"
	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/value_objects"
)

type allocatePaymentAddressUseCase struct {
	unitOfWork                     outport.UnitOfWork
	deriver                        outport.BitcoinAddressDeriver
	policyReader                   outport.AddressPolicyReader
	requiredConfirmationsByNetwork map[value_objects.BitcoinNetwork]int32
}

const defaultIssueReceiptRequiredConfirmations int32 = 1

func NewAllocatePaymentAddressUseCase(
	unitOfWork outport.UnitOfWork,
	deriver outport.BitcoinAddressDeriver,
	policyReader outport.AddressPolicyReader,
	requiredConfirmationsByNetwork ...map[value_objects.BitcoinNetwork]int32,
) inport.AllocatePaymentAddressUseCase {
	confirmationsByNetwork := make(map[value_objects.BitcoinNetwork]int32)
	if len(requiredConfirmationsByNetwork) > 0 {
		for network, confirmations := range requiredConfirmationsByNetwork[0] {
			if confirmations <= 0 {
				continue
			}
			confirmationsByNetwork[network] = confirmations
		}
	}
	return &allocatePaymentAddressUseCase{
		unitOfWork:                     unitOfWork,
		deriver:                        deriver,
		policyReader:                   policyReader,
		requiredConfirmationsByNetwork: confirmationsByNetwork,
	}
}

func (uc *allocatePaymentAddressUseCase) Execute(
	ctx context.Context,
	input dto.AllocatePaymentAddressInput,
) (dto.AllocatePaymentAddressResponse, error) {
	if input.Chain != value_objects.ChainBitcoin {
		return dto.AllocatePaymentAddressResponse{}, inport.ErrChainNotSupported
	}

	policy, ok, err := uc.policyReader.FindByID(ctx, input.AddressPolicyID)
	if err != nil {
		return dto.AllocatePaymentAddressResponse{}, err
	}
	if !ok || policy.Chain != input.Chain {
		return dto.AllocatePaymentAddressResponse{}, inport.ErrAddressPolicyNotFound
	}
	if !policy.IsEnabled() {
		return dto.AllocatePaymentAddressResponse{}, inport.ErrAddressPolicyNotEnabled
	}
	if input.ExpectedAmountMinor <= 0 {
		return dto.AllocatePaymentAddressResponse{}, inport.ErrInvalidExpectedAmount
	}

	customerReference := strings.TrimSpace(input.CustomerReference)
	xpubFingerprintAlgo := strings.TrimSpace(policy.XPubFingerprintAlgo)
	xpubFingerprint := strings.TrimSpace(policy.XPubFingerprint)
	if xpubFingerprintAlgo == "" || xpubFingerprint == "" {
		return dto.AllocatePaymentAddressResponse{}, errors.New("address policy fingerprint is not configured")
	}

	reserveInput := outport.ReservePaymentAddressAllocationInput{
		Policy:              policy,
		ExpectedAmountMinor: input.ExpectedAmountMinor,
		CustomerReference:   customerReference,
	}

	var allocation entities.PaymentAddressAllocation
	var issuedAllocation entities.PaymentAddressAllocation
	var businessErr error
	if err := uc.unitOfWork.WithinTransaction(ctx, func(
		txRepositories outport.TxRepositories,
	) error {
		allocationRepository := txRepositories.PaymentAddressAllocation
		if allocationRepository == nil {
			return errors.New("payment address allocation repository is not configured")
		}
		receiptTrackingRepository := txRepositories.PaymentReceiptTracking
		if receiptTrackingRepository == nil {
			return errors.New("payment receipt tracking repository is not configured")
		}

		reopenedAllocation, reopened, reopenErr := allocationRepository.ReopenFailedReservation(ctx, reserveInput)
		if reopenErr != nil {
			return reopenErr
		}
		if reopened {
			allocation = reopenedAllocation
		} else {
			freshAllocation, reserveErr := allocationRepository.ReserveFresh(ctx, reserveInput)
			if reserveErr != nil {
				return reserveErr
			}
			allocation = freshAllocation
		}

		address, deriveErr := uc.deriver.DeriveAddress(
			policy.Network,
			policy.Scheme,
			policy.XPub,
			allocation.DerivationIndex,
		)
		if deriveErr != nil {
			failedAllocation, markErr := allocation.MarkDerivationFailed(deriveErr.Error())
			if markErr != nil {
				return markErr
			}
			if err := allocationRepository.MarkDerivationFailed(ctx, failedAllocation); err != nil {
				return err
			}
			businessErr = deriveErr
			return nil
		}

		relativePath, pathErr := uc.deriver.DerivationPath(policy.XPub, allocation.DerivationIndex)
		if pathErr != nil {
			failedAllocation, markErr := allocation.MarkDerivationFailed(pathErr.Error())
			if markErr != nil {
				return markErr
			}
			if err := allocationRepository.MarkDerivationFailed(ctx, failedAllocation); err != nil {
				return err
			}
			businessErr = pathErr
			return nil
		}

		updatedAllocation, markIssuedErr := allocation.MarkIssued(policy, address, relativePath)
		if markIssuedErr != nil {
			failedAllocation, markErr := allocation.MarkDerivationFailed(markIssuedErr.Error())
			if markErr != nil {
				return markErr
			}
			if err := allocationRepository.MarkDerivationFailed(ctx, failedAllocation); err != nil {
				return err
			}
			businessErr = markIssuedErr
			return nil
		}
		issuedAllocation = updatedAllocation

		if err := allocationRepository.Complete(ctx, issuedAllocation); err != nil {
			return err
		}
		if _, err := receiptTrackingRepository.RegisterIssuedAllocation(
			ctx,
			issuedAllocation.PaymentAddressID,
			uc.requiredConfirmationsForNetwork(policy.Network),
		); err != nil {
			return err
		}
		return nil
	}); err != nil {
		if errors.Is(err, outport.ErrAddressIndexExhausted) {
			return dto.AllocatePaymentAddressResponse{}, inport.ErrAddressPoolExhausted
		}
		return dto.AllocatePaymentAddressResponse{}, err
	}
	if businessErr != nil {
		return dto.AllocatePaymentAddressResponse{}, businessErr
	}
	if allocation.PaymentAddressID <= 0 {
		return dto.AllocatePaymentAddressResponse{}, errors.New("payment address id must be greater than zero")
	}

	return dto.AllocatePaymentAddressResponse{
		PaymentAddressID:    strconv.FormatInt(allocation.PaymentAddressID, 10),
		AddressPolicyID:     policy.AddressPolicyID,
		ExpectedAmountMinor: input.ExpectedAmountMinor,
		Chain:               string(policy.Chain),
		Network:             string(policy.Network),
		Scheme:              string(policy.Scheme),
		MinorUnit:           policy.MinorUnit,
		Decimals:            policy.Decimals,
		Address:             issuedAllocation.Address,
		CustomerReference:   customerReference,
	}, nil
}

func (uc *allocatePaymentAddressUseCase) requiredConfirmationsForNetwork(network value_objects.BitcoinNetwork) int32 {
	if uc.requiredConfirmationsByNetwork == nil {
		return defaultIssueReceiptRequiredConfirmations
	}
	if configured, ok := uc.requiredConfirmationsByNetwork[network]; ok && configured > 0 {
		return configured
	}
	return defaultIssueReceiptRequiredConfirmations
}
