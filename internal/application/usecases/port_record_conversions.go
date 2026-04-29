package usecases

import (
	"errors"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/events"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
)

var errPortRecordInvalid = errors.New("port record is invalid")

func addressIssuancePolicyFromRecord(record outport.AddressIssuancePolicyRecord) (policies.AddressIssuancePolicy, error) {
	addressPolicyID, err := valueobjects.NewAddressPolicyID(record.AddressPolicyID)
	if err != nil {
		return policies.AddressIssuancePolicy{}, err
	}
	chain, ok := valueobjects.ParseSupportedChain(record.Chain)
	if !ok {
		return policies.AddressIssuancePolicy{}, errPortRecordInvalid
	}
	network, ok := valueobjects.ParseNetworkID(record.Network)
	if !ok {
		return policies.AddressIssuancePolicy{}, errPortRecordInvalid
	}
	scheme, ok := valueobjects.ParseAddressScheme(record.Scheme)
	if !ok {
		return policies.AddressIssuancePolicy{}, errPortRecordInvalid
	}

	return policies.AddressIssuancePolicy{
		AddressPolicyID: addressPolicyID,
		Chain:           chain,
		Network:         network,
		Scheme:          scheme,
		AssetReference:  record.AssetReference,
		Decimals:        record.Decimals,
		Enabled:         record.Enabled,
		IssuanceConfig: valueobjects.AddressIssuanceConfig{
			AddressSpaceRef:   record.AddressSpaceRef,
			IssuanceRefPrefix: record.IssuanceRefPrefix,
		},
	}.Normalize(), nil
}

func addressIssuancePolicyRecordFromDomain(policy policies.AddressIssuancePolicy) outport.AddressIssuancePolicyRecord {
	normalized := policy.Normalize()
	return outport.AddressIssuancePolicyRecord{
		AddressPolicyID:   string(normalized.AddressPolicyID),
		Chain:             string(normalized.Chain),
		Network:           string(normalized.Network),
		Scheme:            string(normalized.Scheme),
		AssetReference:    normalized.AssetReference,
		Decimals:          normalized.Decimals,
		Enabled:           normalized.Enabled,
		AddressSpaceRef:   normalized.IssuanceConfig.AddressSpaceRef,
		IssuanceRefPrefix: normalized.IssuanceConfig.IssuanceRefPrefix,
	}
}

func paymentAddressAllocationFromRecord(record outport.PaymentAddressAllocationRecord) (entities.PaymentAddressAllocation, error) {
	addressPolicyID, err := valueobjects.NewAddressPolicyID(record.AddressPolicyID)
	if err != nil {
		return entities.PaymentAddressAllocation{}, err
	}
	allocation := entities.PaymentAddressAllocation{
		PaymentAddressID:    record.PaymentAddressID,
		AddressPolicyID:     addressPolicyID,
		SlotIndex:           record.SlotIndex,
		ExpectedAmountMinor: record.ExpectedAmountMinor,
		CustomerReference:   record.CustomerReference,
		Status:              valueobjects.PaymentAddressAllocationStatus(record.Status),
		AssetReference:      record.AssetReference,
		Address:             record.Address,
	}
	if record.Chain != "" {
		chain, ok := valueobjects.ParseSupportedChain(record.Chain)
		if !ok {
			return entities.PaymentAddressAllocation{}, errPortRecordInvalid
		}
		allocation.Chain = chain
	}
	if record.Network != "" {
		network, ok := valueobjects.ParseNetworkID(record.Network)
		if !ok {
			return entities.PaymentAddressAllocation{}, errPortRecordInvalid
		}
		allocation.Network = network
	}
	if record.Scheme != "" {
		scheme, ok := valueobjects.ParseAddressScheme(record.Scheme)
		if !ok {
			return entities.PaymentAddressAllocation{}, errPortRecordInvalid
		}
		allocation.Scheme = scheme
	}
	if record.DerivationFailureReason != "" {
		reason, ok := valueobjects.ParsePaymentAddressAllocationDerivationFailureReason(record.DerivationFailureReason)
		if !ok {
			return entities.PaymentAddressAllocation{}, errPortRecordInvalid
		}
		allocation.DerivationFailureReason = reason
	}
	return allocation, nil
}

func paymentAddressAllocationRecordFromDomain(allocation entities.PaymentAddressAllocation) outport.PaymentAddressAllocationRecord {
	return outport.PaymentAddressAllocationRecord{
		PaymentAddressID:        allocation.PaymentAddressID,
		AddressPolicyID:         string(allocation.AddressPolicyID),
		SlotIndex:               allocation.SlotIndex,
		ExpectedAmountMinor:     allocation.ExpectedAmountMinor,
		CustomerReference:       allocation.CustomerReference,
		Status:                  string(allocation.Status),
		Chain:                   string(allocation.Chain),
		Network:                 string(allocation.Network),
		Scheme:                  string(allocation.Scheme),
		AssetReference:          allocation.AssetReference,
		Address:                 allocation.Address,
		DerivationFailureReason: string(allocation.DerivationFailureReason),
	}
}

func paymentReceiptTrackingFromRecord(record outport.PaymentReceiptTrackingRecord) (entities.PaymentReceiptTracking, error) {
	addressPolicyID, err := valueobjects.NewAddressPolicyID(record.AddressPolicyID)
	if err != nil {
		return entities.PaymentReceiptTracking{}, err
	}
	chain, ok := valueobjects.ParseChainID(record.Chain)
	if !ok {
		return entities.PaymentReceiptTracking{}, errPortRecordInvalid
	}
	network, ok := valueobjects.ParseNetworkID(record.Network)
	if !ok {
		return entities.PaymentReceiptTracking{}, errPortRecordInvalid
	}
	status, ok := valueobjects.ParsePaymentReceiptStatus(record.Status)
	if !ok {
		return entities.PaymentReceiptTracking{}, errPortRecordInvalid
	}

	tracking := entities.PaymentReceiptTracking{
		TrackingID:              record.TrackingID,
		PaymentAddressID:        record.PaymentAddressID,
		AddressPolicyID:         addressPolicyID,
		Chain:                   chain,
		Network:                 network,
		Address:                 record.Address,
		AssetReference:          record.AssetReference,
		IssuedAt:                record.IssuedAt,
		ExpectedAmountMinor:     record.ExpectedAmountMinor,
		RequiredConfirmations:   record.RequiredConfirmations,
		Status:                  status,
		ObservedTotalMinor:      record.ObservedTotalMinor,
		ConfirmedTotalMinor:     record.ConfirmedTotalMinor,
		UnconfirmedTotalMinor:   record.UnconfirmedTotalMinor,
		LastObservedBlockHeight: record.LastObservedBlockHeight,
		FirstObservedAt:         record.FirstObservedAt,
		PaidAt:                  record.PaidAt,
		ConfirmedAt:             record.ConfirmedAt,
		ExpiresAt:               record.ExpiresAt,
	}
	if record.LastFailureReason != "" {
		reason, ok := valueobjects.ParsePaymentReceiptTrackingFailureReason(record.LastFailureReason)
		if !ok {
			return entities.PaymentReceiptTracking{}, errPortRecordInvalid
		}
		tracking.LastFailureReason = reason
	}
	return tracking, nil
}

func paymentReceiptTrackingRecordFromDomain(tracking entities.PaymentReceiptTracking) outport.PaymentReceiptTrackingRecord {
	return outport.PaymentReceiptTrackingRecord{
		TrackingID:              tracking.TrackingID,
		PaymentAddressID:        tracking.PaymentAddressID,
		AddressPolicyID:         string(tracking.AddressPolicyID),
		Chain:                   string(tracking.Chain),
		Network:                 string(tracking.Network),
		Address:                 tracking.Address,
		AssetReference:          tracking.AssetReference,
		IssuedAt:                tracking.IssuedAt,
		ExpectedAmountMinor:     tracking.ExpectedAmountMinor,
		RequiredConfirmations:   tracking.RequiredConfirmations,
		Status:                  string(tracking.Status),
		ObservedTotalMinor:      tracking.ObservedTotalMinor,
		ConfirmedTotalMinor:     tracking.ConfirmedTotalMinor,
		UnconfirmedTotalMinor:   tracking.UnconfirmedTotalMinor,
		LastObservedBlockHeight: tracking.LastObservedBlockHeight,
		FirstObservedAt:         tracking.FirstObservedAt,
		PaidAt:                  tracking.PaidAt,
		ConfirmedAt:             tracking.ConfirmedAt,
		ExpiresAt:               tracking.ExpiresAt,
		LastFailureReason:       string(tracking.LastFailureReason),
	}
}

func paymentReceiptStatusChangedRecordFromDomain(event events.PaymentReceiptStatusChanged) outport.PaymentReceiptStatusChangedRecord {
	return outport.PaymentReceiptStatusChangedRecord{
		PaymentAddressID:      event.PaymentAddressID,
		PreviousStatus:        string(event.PreviousStatus),
		CurrentStatus:         string(event.CurrentStatus),
		ObservedTotalMinor:    event.ObservedTotalMinor,
		ConfirmedTotalMinor:   event.ConfirmedTotalMinor,
		UnconfirmedTotalMinor: event.UnconfirmedTotalMinor,
		StatusChangedAt:       event.StatusChangedAt,
	}
}
