package entities

import (
	"errors"
	"strings"
	"time"

	"payrune/internal/domain/value_objects"
)

type PaymentReceiptTracking struct {
	TrackingID              int64
	PaymentAddressID        int64
	AddressPolicyID         string
	Chain                   value_objects.ChainID
	Network                 value_objects.NetworkID
	Address                 string
	IssuedAt                time.Time
	ExpectedAmountMinor     int64
	RequiredConfirmations   int32
	Status                  value_objects.PaymentReceiptStatus
	ObservedTotalMinor      int64
	ConfirmedTotalMinor     int64
	UnconfirmedTotalMinor   int64
	ConflictTotalMinor      int64
	LastObservedBlockHeight int64
	FirstObservedAt         *time.Time
	PaidAt                  *time.Time
	ConfirmedAt             *time.Time
	LastError               string
}

func NewPaymentReceiptTracking(
	paymentAddressID int64,
	addressPolicyID string,
	chain value_objects.ChainID,
	network value_objects.NetworkID,
	address string,
	issuedAt time.Time,
	expectedAmountMinor int64,
	requiredConfirmations int32,
) (PaymentReceiptTracking, error) {
	normalizedPolicyID := strings.TrimSpace(addressPolicyID)
	normalizedAddress := strings.TrimSpace(address)
	normalizedChain, chainOK := value_objects.ParseChainID(string(chain))
	normalizedNetwork, networkOK := value_objects.ParseNetworkID(string(network))

	if paymentAddressID <= 0 {
		return PaymentReceiptTracking{}, errors.New("payment address id must be greater than zero")
	}
	if normalizedPolicyID == "" {
		return PaymentReceiptTracking{}, errors.New("address policy id is required")
	}
	if !chainOK {
		return PaymentReceiptTracking{}, errors.New("chain is invalid")
	}
	if !networkOK {
		return PaymentReceiptTracking{}, errors.New("network is invalid")
	}
	if normalizedAddress == "" {
		return PaymentReceiptTracking{}, errors.New("address is required")
	}
	if issuedAt.IsZero() {
		return PaymentReceiptTracking{}, errors.New("issued at is required")
	}
	if expectedAmountMinor <= 0 {
		return PaymentReceiptTracking{}, errors.New("expected amount minor must be greater than zero")
	}
	if requiredConfirmations <= 0 {
		return PaymentReceiptTracking{}, errors.New("required confirmations must be greater than zero")
	}

	return PaymentReceiptTracking{
		PaymentAddressID:      paymentAddressID,
		AddressPolicyID:       normalizedPolicyID,
		Chain:                 normalizedChain,
		Network:               normalizedNetwork,
		Address:               normalizedAddress,
		IssuedAt:              issuedAt.UTC(),
		ExpectedAmountMinor:   expectedAmountMinor,
		RequiredConfirmations: requiredConfirmations,
		Status:                value_objects.PaymentReceiptStatusWatching,
	}, nil
}

func (t PaymentReceiptTracking) ApplyObservation(
	observation value_objects.PaymentReceiptObservation,
	observedAt time.Time,
) (PaymentReceiptTracking, error) {
	if err := observation.Validate(); err != nil {
		return PaymentReceiptTracking{}, err
	}
	if observedAt.IsZero() {
		return PaymentReceiptTracking{}, errors.New("observed time is required")
	}

	updated := t
	updated.ObservedTotalMinor = observation.ObservedTotalMinor
	updated.ConfirmedTotalMinor = observation.ConfirmedTotalMinor
	updated.UnconfirmedTotalMinor = observation.UnconfirmedTotalMinor
	updated.ConflictTotalMinor = observation.ConflictTotalMinor
	updated.LastObservedBlockHeight = observation.LatestBlockHeight
	updated.LastError = ""

	if observation.ObservedTotalMinor > 0 && updated.FirstObservedAt == nil {
		firstObservedAt := observedAt
		updated.FirstObservedAt = &firstObservedAt
	}
	if observation.ObservedTotalMinor >= updated.ExpectedAmountMinor && updated.PaidAt == nil {
		paidAt := observedAt
		updated.PaidAt = &paidAt
	}
	if observation.ConfirmedTotalMinor >= updated.ExpectedAmountMinor && updated.ConfirmedAt == nil {
		confirmedAt := observedAt
		updated.ConfirmedAt = &confirmedAt
	}

	updated.Status = decidePaymentReceiptStatus(updated, observation)
	return updated, nil
}

func (t PaymentReceiptTracking) MarkPollingError(reason string) (PaymentReceiptTracking, error) {
	normalizedReason := strings.TrimSpace(reason)
	if normalizedReason == "" {
		return PaymentReceiptTracking{}, errors.New("polling error reason is required")
	}

	updated := t
	updated.LastError = normalizedReason
	return updated, nil
}

func decidePaymentReceiptStatus(
	tracking PaymentReceiptTracking,
	observation value_objects.PaymentReceiptObservation,
) value_objects.PaymentReceiptStatus {
	if observation.ConflictTotalMinor > 0 {
		return value_objects.PaymentReceiptStatusDoubleSpendSuspected
	}
	if observation.ObservedTotalMinor == 0 {
		return value_objects.PaymentReceiptStatusWatching
	}
	if observation.ConfirmedTotalMinor >= tracking.ExpectedAmountMinor {
		return value_objects.PaymentReceiptStatusPaidConfirmed
	}
	if observation.ObservedTotalMinor >= tracking.ExpectedAmountMinor {
		return value_objects.PaymentReceiptStatusPaidUnconfirmed
	}
	return value_objects.PaymentReceiptStatusPartiallyPaid
}
