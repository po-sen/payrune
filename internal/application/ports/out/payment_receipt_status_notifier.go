package out

import (
	"context"
	"time"
)

const (
	PaymentReceiptStatusChangedEventType    = "payment_receipt.status_changed"
	PaymentReceiptStatusChangedEventVersion = 1
)

type NotifyPaymentReceiptStatusChangedInput struct {
	NotificationID        int64
	PaymentAddressID      int64
	CustomerReference     string
	PreviousStatus        string
	CurrentStatus         string
	ObservedTotalMinor    int64
	ConfirmedTotalMinor   int64
	UnconfirmedTotalMinor int64
	StatusChangedAt       time.Time
}

type PaymentReceiptStatusNotifier interface {
	NotifyStatusChanged(ctx context.Context, input NotifyPaymentReceiptStatusChangedInput) error
}
