package entities

import (
	"strings"
	"time"

	"payrune/internal/domain/valueobjects"
)

type PaymentAddressAllocation struct {
	PaymentAddressID        int64
	AddressPolicyID         string
	SlotIndex               uint32
	ExpectedAmountMinor     int64
	CustomerReference       string
	Status                  valueobjects.PaymentAddressAllocationStatus
	Chain                   valueobjects.SupportedChain
	Network                 valueobjects.NetworkID
	Scheme                  string
	Address                 string
	SweepMaterialJSON       string
	DerivationFailureReason valueobjects.PaymentAddressAllocationDerivationFailureReason
}

func NewPaymentAddressAllocation(
	paymentAddressID int64,
	addressPolicyID string,
	slotIndex uint32,
	expectedAmountMinor int64,
	customerReference string,
) (PaymentAddressAllocation, error) {
	normalizedPolicyID := strings.TrimSpace(addressPolicyID)
	if paymentAddressID <= 0 {
		return PaymentAddressAllocation{}, ErrPaymentAddressIDInvalid
	}
	if normalizedPolicyID == "" {
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
	policy AddressIssuancePolicy,
	address string,
	sweepMaterialJSON string,
) (PaymentAddressAllocation, error) {
	policy = policy.Normalize()
	if policy.AddressPolicy.AddressPolicyID == "" {
		return PaymentAddressAllocation{}, ErrAddressPolicyIDRequired
	}
	if policy.AddressPolicy.AddressPolicyID != a.AddressPolicyID {
		return PaymentAddressAllocation{}, ErrAddressPolicyMismatch
	}

	normalizedAddress := strings.TrimSpace(address)
	if normalizedAddress == "" {
		return PaymentAddressAllocation{}, ErrAddressRequired
	}
	normalizedSweepMaterialJSON := strings.TrimSpace(sweepMaterialJSON)
	if normalizedSweepMaterialJSON == "" {
		return PaymentAddressAllocation{}, ErrSweepMaterialRequired
	}

	issued := a
	issued.Status = valueobjects.PaymentAddressAllocationStatusIssued
	issued.Chain = policy.AddressPolicy.Chain
	issued.Network = policy.AddressPolicy.Network
	issued.Scheme = policy.AddressPolicy.Scheme
	issued.Address = normalizedAddress
	issued.SweepMaterialJSON = normalizedSweepMaterialJSON
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
	failed.SweepMaterialJSON = ""
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
