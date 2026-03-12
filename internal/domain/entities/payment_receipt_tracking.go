package entities

import (
	"errors"
	"strings"
	"time"

	"payrune/internal/domain/events"
	"payrune/internal/domain/valueobjects"
)

type PaymentReceiptTracking struct {
	TrackingID              int64
	PaymentAddressID        int64
	AddressPolicyID         string
	Chain                   valueobjects.ChainID
	Network                 valueobjects.NetworkID
	Address                 string
	IssuedAt                time.Time
	ExpectedAmountMinor     int64
	RequiredConfirmations   int32
	Status                  valueobjects.PaymentReceiptStatus
	ObservedTotalMinor      int64
	ConfirmedTotalMinor     int64
	UnconfirmedTotalMinor   int64
	LastObservedBlockHeight int64
	FirstObservedAt         *time.Time
	PaidAt                  *time.Time
	ConfirmedAt             *time.Time
	ExpiresAt               *time.Time
	LastError               string
}

func NewPaymentReceiptTracking(
	paymentAddressID int64,
	addressPolicyID string,
	chain valueobjects.ChainID,
	network valueobjects.NetworkID,
	address string,
	issuedAt time.Time,
	expectedAmountMinor int64,
	requiredConfirmations int32,
) (PaymentReceiptTracking, error) {
	normalizedPolicyID := strings.TrimSpace(addressPolicyID)
	normalizedAddress := strings.TrimSpace(address)
	normalizedChain, chainOK := valueobjects.ParseChainID(string(chain))
	normalizedNetwork, networkOK := valueobjects.ParseNetworkID(string(network))

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
		Status:                valueobjects.PaymentReceiptStatusWatching,
	}, nil
}

func (t PaymentReceiptTracking) ApplyObservation(
	observation valueobjects.PaymentReceiptObservation,
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

func (t PaymentReceiptTracking) IsExpired(now time.Time) bool {
	if t.ExpiresAt == nil || now.IsZero() {
		return false
	}
	return !t.ExpiresAt.After(now)
}

func (t PaymentReceiptTracking) CanExpireByPaymentWindow() bool {
	return t.PaidAt == nil
}

func (t PaymentReceiptTracking) MarkExpired(reason string) (PaymentReceiptTracking, error) {
	normalizedReason := strings.TrimSpace(reason)
	if normalizedReason == "" {
		return PaymentReceiptTracking{}, errors.New("expired reason is required")
	}

	updated := t
	updated.Status = valueobjects.PaymentReceiptStatusFailedExpired
	updated.LastError = normalizedReason
	return updated, nil
}

func (t PaymentReceiptTracking) StatusChangedEvent(
	previousStatus valueobjects.PaymentReceiptStatus,
	changedAt time.Time,
) (events.PaymentReceiptStatusChanged, bool, error) {
	if previousStatus == t.Status {
		return events.PaymentReceiptStatusChanged{}, false, nil
	}

	event, err := events.NewPaymentReceiptStatusChanged(
		t.PaymentAddressID,
		previousStatus,
		t.Status,
		t.ObservedTotalMinor,
		t.ConfirmedTotalMinor,
		t.UnconfirmedTotalMinor,
		changedAt,
	)
	if err != nil {
		return events.PaymentReceiptStatusChanged{}, false, err
	}
	return event, true, nil
}

func PollablePaymentReceiptStatuses() []valueobjects.PaymentReceiptStatus {
	return []valueobjects.PaymentReceiptStatus{
		valueobjects.PaymentReceiptStatusWatching,
		valueobjects.PaymentReceiptStatusPartiallyPaid,
		valueobjects.PaymentReceiptStatusPaidUnconfirmed,
		valueobjects.PaymentReceiptStatusPaidUnconfirmedReverted,
	}
}

func decidePaymentReceiptStatus(
	tracking PaymentReceiptTracking,
	observation valueobjects.PaymentReceiptObservation,
) valueobjects.PaymentReceiptStatus {
	if observation.ConfirmedTotalMinor >= tracking.ExpectedAmountMinor {
		return valueobjects.PaymentReceiptStatusPaidConfirmed
	}
	if observation.ObservedTotalMinor >= tracking.ExpectedAmountMinor {
		return valueobjects.PaymentReceiptStatusPaidUnconfirmed
	}
	if tracking.PaidAt != nil {
		return valueobjects.PaymentReceiptStatusPaidUnconfirmedReverted
	}
	if observation.ObservedTotalMinor == 0 {
		return valueobjects.PaymentReceiptStatusWatching
	}
	return valueobjects.PaymentReceiptStatusPartiallyPaid
}
