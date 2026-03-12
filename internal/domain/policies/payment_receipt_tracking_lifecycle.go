package policies

import (
	"time"

	"payrune/internal/domain/entities"
	"payrune/internal/domain/valueobjects"
)

const (
	defaultPaymentReceiptExpiredReason = "payment window expired"
)

type PaymentReceiptTrackingLifecyclePolicy struct{}

func NewPaymentReceiptTrackingLifecyclePolicy() PaymentReceiptTrackingLifecyclePolicy {
	return PaymentReceiptTrackingLifecyclePolicy{}
}

func (p PaymentReceiptTrackingLifecyclePolicy) ExpireIfDue(
	tracking entities.PaymentReceiptTracking,
	now time.Time,
) (entities.PaymentReceiptTracking, bool, error) {
	if !tracking.CanExpireByPaymentWindow() {
		return tracking, false, nil
	}
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
	observation valueobjects.PaymentReceiptObservation,
	observedAt time.Time,
) (entities.PaymentReceiptTracking, error) {
	return tracking.ApplyObservation(observation, observedAt)
}
