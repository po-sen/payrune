package entities

import (
	"errors"
	"strings"
	"time"

	"payrune/internal/domain/value_objects"
)

type PaymentAddressAllocation struct {
	PaymentAddressID    int64
	AddressPolicyID     string
	DerivationIndex     uint32
	ExpectedAmountMinor int64
	CustomerReference   string
	Status              value_objects.PaymentAddressAllocationStatus
	Chain               value_objects.SupportedChain
	Network             value_objects.NetworkID
	Scheme              string
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
	policy AddressIssuancePolicy,
	address string,
	relativeDerivationPath string,
) (PaymentAddressAllocation, error) {
	policy = policy.Normalize()
	if policy.AddressPolicy.AddressPolicyID == "" {
		return PaymentAddressAllocation{}, errors.New("address policy id is required")
	}
	if policy.AddressPolicy.AddressPolicyID != a.AddressPolicyID {
		return PaymentAddressAllocation{}, errors.New("address policy mismatch")
	}

	normalizedAddress := strings.TrimSpace(address)
	if normalizedAddress == "" {
		return PaymentAddressAllocation{}, errors.New("address is required")
	}

	absolutePath, err := policy.DerivationConfig.AbsoluteDerivationPath(relativeDerivationPath)
	if err != nil {
		return PaymentAddressAllocation{}, err
	}

	issued := a
	issued.Status = value_objects.PaymentAddressAllocationStatusIssued
	issued.Chain = policy.AddressPolicy.Chain
	issued.Network = policy.AddressPolicy.Network
	issued.Scheme = policy.AddressPolicy.Scheme
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

func (a PaymentAddressAllocation) IssueReceiptTracking(
	issuedAt time.Time,
	requiredConfirmations int32,
	expiresAt time.Time,
) (PaymentReceiptTracking, error) {
	if a.Status != value_objects.PaymentAddressAllocationStatusIssued {
		return PaymentReceiptTracking{}, errors.New("payment address allocation is not issued")
	}
	if expiresAt.IsZero() {
		return PaymentReceiptTracking{}, errors.New("expires at is required")
	}

	chainID, ok := value_objects.ParseChainID(string(a.Chain))
	if !ok {
		return PaymentReceiptTracking{}, errors.New("chain is invalid")
	}
	networkID, ok := value_objects.ParseNetworkID(string(a.Network))
	if !ok {
		return PaymentReceiptTracking{}, errors.New("network is invalid")
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
