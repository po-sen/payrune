package entities

import (
	"strings"
	"time"

	"payrune/internal/domain/valueobjects"
)

type PaymentAddressAllocation struct {
	PaymentAddressID        int64
	AddressPolicyID         valueobjects.AddressPolicyID
	SlotIndex               uint32
	ExpectedAmountMinor     int64
	CustomerReference       string
	Status                  valueobjects.PaymentAddressAllocationStatus
	Chain                   valueobjects.SupportedChain
	Network                 valueobjects.NetworkID
	Scheme                  valueobjects.AddressScheme
	Address                 string
	DerivationFailureReason valueobjects.PaymentAddressAllocationDerivationFailureReason
}

func NewPaymentAddressAllocation(
	paymentAddressID int64,
	addressPolicyID valueobjects.AddressPolicyID,
	slotIndex uint32,
	expectedAmountMinor int64,
	customerReference string,
) (PaymentAddressAllocation, error) {
	normalizedPolicyID := addressPolicyID.Normalize()
	if paymentAddressID <= 0 {
		return PaymentAddressAllocation{}, ErrPaymentAddressIDInvalid
	}
	if normalizedPolicyID.IsZero() {
		return PaymentAddressAllocation{}, ErrAddressPolicyIDRequired
	}
	if expectedAmountMinor <= 0 {
		return PaymentAddressAllocation{}, ErrExpectedAmountMinorInvalid
	}

	return PaymentAddressAllocation{
		PaymentAddressID:    paymentAddressID,
		AddressPolicyID:     normalizedPolicyID,
		SlotIndex:           slotIndex,
		ExpectedAmountMinor: expectedAmountMinor,
		CustomerReference:   strings.TrimSpace(customerReference),
		Status:              valueobjects.PaymentAddressAllocationStatusReserved,
	}, nil
}

func (a PaymentAddressAllocation) MarkIssued(
	addressPolicyID valueobjects.AddressPolicyID,
	chain valueobjects.SupportedChain,
	network valueobjects.NetworkID,
	scheme valueobjects.AddressScheme,
	address string,
) (PaymentAddressAllocation, error) {
	normalizedPolicyID := addressPolicyID.Normalize()
	if normalizedPolicyID.IsZero() {
		return PaymentAddressAllocation{}, ErrAddressPolicyIDRequired
	}
	if normalizedPolicyID != a.AddressPolicyID {
		return PaymentAddressAllocation{}, ErrAddressPolicyMismatch
	}
	normalizedScheme := scheme.Normalize()

	normalizedAddress := strings.TrimSpace(address)
	if normalizedAddress == "" {
		return PaymentAddressAllocation{}, ErrAddressRequired
	}

	issued := a
	issued.Status = valueobjects.PaymentAddressAllocationStatusIssued
	issued.Chain = chain
	issued.Network = network
	issued.Scheme = normalizedScheme
	issued.Address = normalizedAddress
	issued.DerivationFailureReason = ""

	return issued, nil
}

func (a PaymentAddressAllocation) MarkDerivationFailed(
	reason valueobjects.PaymentAddressAllocationDerivationFailureReason,
) (PaymentAddressAllocation, error) {
	if reason.IsZero() {
		return PaymentAddressAllocation{}, ErrDerivationFailureReasonRequired
	}

	failed := a
	failed.Status = valueobjects.PaymentAddressAllocationStatusDerivationFailed
	failed.DerivationFailureReason = reason
	failed.Chain = ""
	failed.Network = ""
	failed.Scheme = ""
	failed.Address = ""
	return failed, nil
}

func (a PaymentAddressAllocation) IssueReceiptTracking(
	issuedAt time.Time,
	requiredConfirmations int32,
	expiresAt time.Time,
) (PaymentReceiptTracking, error) {
	if a.Status != valueobjects.PaymentAddressAllocationStatusIssued {
		return PaymentReceiptTracking{}, ErrPaymentAddressAllocationNotIssued
	}
	if expiresAt.IsZero() {
		return PaymentReceiptTracking{}, ErrExpiresAtRequired
	}

	chainID, ok := valueobjects.ParseChainID(string(a.Chain))
	if !ok {
		return PaymentReceiptTracking{}, ErrChainInvalid
	}
	networkID, ok := valueobjects.ParseNetworkID(string(a.Network))
	if !ok {
		return PaymentReceiptTracking{}, ErrNetworkInvalid
	}

	tracking, err := NewPaymentReceiptTracking(
		a.PaymentAddressID,
		a.AddressPolicyID,
		chainID,
		networkID,
		a.Address,
		issuedAt,
		a.ExpectedAmountMinor,
		requiredConfirmations,
	)
	if err != nil {
		return PaymentReceiptTracking{}, err
	}

	expiresAtUTC := expiresAt.UTC()
	tracking.ExpiresAt = &expiresAtUTC
	return tracking, nil
}
