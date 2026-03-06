package out

import (
	"context"
	"time"

	"payrune/internal/domain/entities"
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
	ClaimPending(
		ctx context.Context,
		input ClaimPaymentReceiptStatusNotificationsInput,
	) ([]entities.PaymentReceiptStatusNotification, error)
	MarkSent(
		ctx context.Context,
		notificationID int64,
		deliveredAt time.Time,
	) error
	MarkRetryScheduled(
		ctx context.Context,
		input MarkPaymentReceiptStatusNotificationRetryInput,
	) error
	MarkFailed(
		ctx context.Context,
		input MarkPaymentReceiptStatusNotificationFailureInput,
	) error
}

type ClaimPaymentReceiptStatusNotificationsInput struct {
	Now        time.Time
	Limit      int
	ClaimUntil time.Time
}

type MarkPaymentReceiptStatusNotificationRetryInput struct {
	NotificationID int64
	Attempts       int32
	LastError      string
	NextAttemptAt  time.Time
}

type MarkPaymentReceiptStatusNotificationFailureInput struct {
	NotificationID int64
	Attempts       int32
	LastError      string
}
