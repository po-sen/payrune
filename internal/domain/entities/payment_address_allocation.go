package entities

import (
	"errors"
	"strings"

	"payrune/internal/domain/value_objects"
)

type PaymentAddressAllocation struct {
	PaymentAddressID    int64
	AddressPolicyID     string
	DerivationIndex     uint32
	ExpectedAmountMinor int64
	CustomerReference   string
	Status              value_objects.PaymentAddressAllocationStatus
	Chain               value_objects.Chain
	Network             value_objects.BitcoinNetwork
	Scheme              value_objects.BitcoinAddressScheme
	Address             string
	DerivationPath      string
	FailureReason       string
}

func NewPaymentAddressAllocation(
	paymentAddressID int64,
	addressPolicyID string,
	derivationIndex uint32,
	expectedAmountMinor int64,
	customerReference string,
) (PaymentAddressAllocation, error) {
	normalizedPolicyID := strings.TrimSpace(addressPolicyID)
	if paymentAddressID <= 0 {
		return PaymentAddressAllocation{}, errors.New("payment address id must be greater than zero")
	}
	if normalizedPolicyID == "" {
		return PaymentAddressAllocation{}, errors.New("address policy id is required")
	}
	if expectedAmountMinor <= 0 {
		return PaymentAddressAllocation{}, errors.New("expected amount minor must be greater than zero")
	}

	return PaymentAddressAllocation{
		PaymentAddressID:    paymentAddressID,
		AddressPolicyID:     normalizedPolicyID,
		DerivationIndex:     derivationIndex,
		ExpectedAmountMinor: expectedAmountMinor,
		CustomerReference:   strings.TrimSpace(customerReference),
		Status:              value_objects.PaymentAddressAllocationStatusReserved,
	}, nil
}

func (a PaymentAddressAllocation) MarkIssued(
	policy AddressPolicy,
	address string,
	relativeDerivationPath string,
) (PaymentAddressAllocation, error) {
	policy = policy.Normalize()
	if policy.AddressPolicyID == "" {
		return PaymentAddressAllocation{}, errors.New("address policy id is required")
	}
	if policy.AddressPolicyID != a.AddressPolicyID {
		return PaymentAddressAllocation{}, errors.New("address policy mismatch")
	}

	normalizedAddress := strings.TrimSpace(address)
	if normalizedAddress == "" {
		return PaymentAddressAllocation{}, errors.New("address is required")
	}

	absolutePath, err := policy.AbsoluteDerivationPath(relativeDerivationPath)
	if err != nil {
		return PaymentAddressAllocation{}, err
	}

	issued := a
	issued.Status = value_objects.PaymentAddressAllocationStatusIssued
	issued.Chain = policy.Chain
	issued.Network = policy.Network
	issued.Scheme = policy.Scheme
	issued.Address = normalizedAddress
	issued.DerivationPath = absolutePath
	issued.FailureReason = ""

	return issued, nil
}

func (a PaymentAddressAllocation) MarkDerivationFailed(reason string) (PaymentAddressAllocation, error) {
	normalizedReason := strings.TrimSpace(reason)
	if normalizedReason == "" {
		return PaymentAddressAllocation{}, errors.New("derivation failure reason is required")
	}

	failed := a
	failed.Status = value_objects.PaymentAddressAllocationStatusDerivationFailed
	failed.FailureReason = normalizedReason
	failed.Chain = ""
	failed.Network = ""
	failed.Scheme = ""
	failed.Address = ""
	failed.DerivationPath = ""
	return failed, nil
}
