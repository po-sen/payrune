package policies

import (
	"time"

	"payrune/internal/domain/entities"
	"payrune/internal/domain/value_objects"
)

const (
	defaultPaymentReceiptExpiredReason                  = "payment window expired"
	defaultPaymentReceiptPaidUnconfirmedExpiryExtension = 7 * 24 * time.Hour
)

type PaymentReceiptTrackingLifecyclePolicy struct {
	paidUnconfirmedExpiryExtension time.Duration
}

func NewPaymentReceiptTrackingLifecyclePolicy(
	paidUnconfirmedExpiryExtension time.Duration,
) PaymentReceiptTrackingLifecyclePolicy {
	return PaymentReceiptTrackingLifecyclePolicy{
		paidUnconfirmedExpiryExtension: paidUnconfirmedExpiryExtension,
	}
}

func (p PaymentReceiptTrackingLifecyclePolicy) ExpireIfDue(
	tracking entities.PaymentReceiptTracking,
	now time.Time,
) (entities.PaymentReceiptTracking, bool, error) {
	if !tracking.IsExpired(now) {
		return tracking, false, nil
	}

	expiredTracking, err := tracking.MarkExpired(defaultPaymentReceiptExpiredReason)
	if err != nil {
		return entities.PaymentReceiptTracking{}, false, err
	}
	return expiredTracking, true, nil
}

func (p PaymentReceiptTrackingLifecyclePolicy) ApplyObservation(
	tracking entities.PaymentReceiptTracking,
	observation value_objects.PaymentReceiptObservation,
	observedAt time.Time,
) (entities.PaymentReceiptTracking, error) {
	updatedTracking, err := tracking.ApplyObservation(observation, observedAt)
	if err != nil {
		return entities.PaymentReceiptTracking{}, err
	}

	return updatedTracking.ExtendExpiryOnTransitionToPaidUnconfirmed(
		tracking.Status,
		observedAt,
		p.paidUnconfirmedExpiryExtensionOrDefault(),
	), nil
}

func (p PaymentReceiptTrackingLifecyclePolicy) paidUnconfirmedExpiryExtensionOrDefault() time.Duration {
	if p.paidUnconfirmedExpiryExtension > 0 {
		return p.paidUnconfirmedExpiryExtension
	}
	return defaultPaymentReceiptPaidUnconfirmedExpiryExtension
}
