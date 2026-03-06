package out

import (
	"context"
	"time"

	"payrune/internal/domain/value_objects"
)

type EnqueuePaymentReceiptStatusChangedInput struct {
	PaymentAddressID      int64
	PreviousStatus        value_objects.PaymentReceiptStatus
	CurrentStatus         value_objects.PaymentReceiptStatus
	ObservedTotalMinor    int64
	ConfirmedTotalMinor   int64
	UnconfirmedTotalMinor int64
	ConflictTotalMinor    int64
	StatusChangedAt       time.Time
}

type PaymentReceiptStatusNotificationRepository interface {
	EnqueueStatusChanged(
		ctx context.Context,
		input EnqueuePaymentReceiptStatusChangedInput,
	) error
}
