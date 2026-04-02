package entities

import (
	"strings"
	"time"

	"payrune/internal/domain/events"
	"payrune/internal/domain/valueobjects"
)

type PaymentReceiptTracking struct {
	TrackingID              int64
	PaymentAddressID        int64
	AddressPolicyID         valueobjects.AddressPolicyID
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
	LastFailureReason       valueobjects.PaymentReceiptTrackingFailureReason
}

func NewPaymentReceiptTracking(
	paymentAddressID int64,
	addressPolicyID valueobjects.AddressPolicyID,
	chain valueobjects.ChainID,
	network valueobjects.NetworkID,
	address string,
	issuedAt time.Time,
	expectedAmountMinor int64,
	requiredConfirmations int32,
) (PaymentReceiptTracking, error) {
	normalizedPolicyID := addressPolicyID.Normalize()
	normalizedAddress := strings.TrimSpace(address)
	normalizedChain, chainOK := valueobjects.ParseChainID(string(chain))
	normalizedNetwork, networkOK := valueobjects.ParseNetworkID(string(network))

	if paymentAddressID <= 0 {
		return PaymentReceiptTracking{}, ErrPaymentAddressIDInvalid
	}
	if normalizedPolicyID.IsZero() {
		return PaymentReceiptTracking{}, ErrAddressPolicyIDRequired
	}
	if !chainOK {
		return PaymentReceiptTracking{}, ErrChainInvalid
	}
	if !networkOK {
		return PaymentReceiptTracking{}, ErrNetworkInvalid
	}
	if normalizedAddress == "" {
		return PaymentReceiptTracking{}, ErrAddressRequired
	}
	if issuedAt.IsZero() {
		return PaymentReceiptTracking{}, ErrIssuedAtRequired
	}
	if expectedAmountMinor <= 0 {
		return PaymentReceiptTracking{}, ErrExpectedAmountMinorInvalid
	}
	if requiredConfirmations <= 0 {
		return PaymentReceiptTracking{}, ErrRequiredConfirmationsInvalid
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
		return PaymentReceiptTracking{}, ErrObservedAtRequired
	}

	updated := t
	updated.ObservedTotalMinor = observation.ObservedTotalMinor
	updated.ConfirmedTotalMinor = observation.ConfirmedTotalMinor
	updated.UnconfirmedTotalMinor = observation.UnconfirmedTotalMinor
	updated.LastObservedBlockHeight = observation.LatestBlockHeight
	updated.LastFailureReason = ""

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

func (t PaymentReceiptTracking) MarkPollingFailure(
	reason valueobjects.PaymentReceiptTrackingFailureReason,
) (PaymentReceiptTracking, error) {
	if reason.IsZero() {
		return PaymentReceiptTracking{}, ErrPaymentReceiptFailureReasonRequired
	}

	updated := t
	updated.LastFailureReason = reason
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

func (t PaymentReceiptTracking) MarkExpired(
	reason valueobjects.PaymentReceiptTrackingFailureReason,
) (PaymentReceiptTracking, error) {
	if reason.IsZero() {
		return PaymentReceiptTracking{}, ErrPaymentReceiptFailureReasonRequired
	}

	updated := t
	updated.Status = valueobjects.PaymentReceiptStatusFailedExpired
	updated.LastFailureReason = reason
	return updated, nil
}

func (t PaymentReceiptTracking) ExpireIfDue(now time.Time) (PaymentReceiptTracking, bool, error) {
	if !t.CanExpireByPaymentWindow() {
		return t, false, nil
	}
	if !t.IsExpired(now) {
		return t, false, nil
	}

	expiredTracking, err := t.MarkExpired(valueobjects.PaymentReceiptTrackingFailureReasonPaymentWindowExpired)
	if err != nil {
		return PaymentReceiptTracking{}, false, err
	}
	return expiredTracking, true, nil
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
