package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"payrune/internal/application/dto"
	applicationoutbox "payrune/internal/application/outbox"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/events"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
)

type fakeReceiptWebhookDispatchClock struct {
	now   time.Time
	times []time.Time
	calls int
}

func (f *fakeReceiptWebhookDispatchClock) NowUTC() time.Time {
	if f.calls < len(f.times) {
		value := f.times[f.calls]
		f.calls++
		return value
	}
	f.calls++
	return f.now
}

type fakeWebhookDispatchNotificationOutbox struct {
	claimRows      []applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage
	claimErr       error
	saveResultErr  error
	lastClaimInput outport.ClaimPaymentReceiptStatusNotificationsInput
	savedResults   []policies.PaymentReceiptStatusNotificationDeliveryResult
	enqueueInputs  []events.PaymentReceiptStatusChanged
	enqueueErr     error
}

func (f *fakeWebhookDispatchNotificationOutbox) EnqueueStatusChanged(
	_ context.Context,
	input events.PaymentReceiptStatusChanged,
) error {
	f.enqueueInputs = append(f.enqueueInputs, input)
	return f.enqueueErr
}

func (f *fakeWebhookDispatchNotificationOutbox) ClaimPending(
	_ context.Context,
	input outport.ClaimPaymentReceiptStatusNotificationsInput,
) ([]applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage, error) {
	f.lastClaimInput = input
	if f.claimErr != nil {
		return nil, f.claimErr
	}
	rows := make([]applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage, len(f.claimRows))
	copy(rows, f.claimRows)
	return rows, nil
}

func (f *fakeWebhookDispatchNotificationOutbox) SaveDeliveryResult(
	_ context.Context,
	result policies.PaymentReceiptStatusNotificationDeliveryResult,
) error {
	f.savedResults = append(f.savedResults, result)
	return f.saveResultErr
}

type fakeReceiptStatusNotifier struct {
	errorsByNotificationID map[int64]error
	inputs                 []outport.NotifyPaymentReceiptStatusChangedInput
}

func (f *fakeReceiptStatusNotifier) NotifyStatusChanged(
	_ context.Context,
	input outport.NotifyPaymentReceiptStatusChangedInput,
) error {
	f.inputs = append(f.inputs, input)
	if err := f.errorsByNotificationID[input.NotificationID]; err != nil {
		return err
	}
	return nil
}

type fakeReceiptWebhookDispatchUnitOfWork struct {
	notificationOutbox outport.PaymentReceiptStatusNotificationOutbox
	err                error
	calls              int
}

func (f *fakeReceiptWebhookDispatchUnitOfWork) WithinTransaction(
	_ context.Context,
	fn func(txScope outport.TxScope) error,
) error {
	f.calls++
	if f.err != nil {
		return f.err
	}
	return fn(outport.TxScope{
		PaymentReceiptStatusNotificationOutbox: f.notificationOutbox,
	})
}

func TestRunReceiptWebhookDispatchCycleUseCaseExecuteSuccess(t *testing.T) {
	claimNow := time.Date(2026, 3, 6, 18, 0, 0, 0, time.UTC)
	deliveredAt := claimNow.Add(3 * time.Second)
	outbox := &fakeWebhookDispatchNotificationOutbox{
		claimRows: []applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage{
			{
				NotificationID:        1,
				PaymentAddressID:      101,
				CustomerReference:     "order-1",
				PreviousStatus:        valueobjects.PaymentReceiptStatusWatching,
				CurrentStatus:         valueobjects.PaymentReceiptStatusPaidConfirmed,
				ObservedTotalMinor:    1000,
				ConfirmedTotalMinor:   1000,
				UnconfirmedTotalMinor: 0,
				StatusChangedAt:       claimNow.Add(-1 * time.Minute),
				DeliveryAttempts:      0,
			},
		},
	}
	notifier := &fakeReceiptStatusNotifier{errorsByNotificationID: map[int64]error{}}
	unitOfWork := &fakeReceiptWebhookDispatchUnitOfWork{notificationOutbox: outbox}
	useCase := NewRunReceiptWebhookDispatchCycleUseCase(
		unitOfWork,
		notifier,
		&fakeReceiptWebhookDispatchClock{
			now:   deliveredAt,
			times: []time.Time{claimNow, deliveredAt},
		},
	)

	output, err := useCase.Execute(context.Background(), dto.RunReceiptWebhookDispatchCycleInput{
		BatchSize:   10,
		DispatchTTL: 15 * time.Second,
		RetryDelay:  time.Minute,
		MaxAttempts: 5,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if output.ClaimedCount != 1 || output.SentCount != 1 || output.RetriedCount != 0 || output.FailedCount != 0 {
		t.Fatalf("unexpected output: %+v", output)
	}
	if got := len(outbox.savedResults); got != 1 {
		t.Fatalf("unexpected saved result count: got %d", got)
	}
	if outbox.savedResults[0].Status != valueobjects.PaymentReceiptNotificationDeliveryStatusSent {
		t.Fatalf("unexpected delivery status: got %q", outbox.savedResults[0].Status)
	}
	if outbox.savedResults[0].DeliveredAt == nil || !outbox.savedResults[0].DeliveredAt.Equal(deliveredAt) {
		t.Fatalf("unexpected delivered at: got %v want %s", outbox.savedResults[0].DeliveredAt, deliveredAt)
	}
	if !outbox.lastClaimInput.Now.Equal(claimNow) {
		t.Fatalf("unexpected claim now: got %s want %s", outbox.lastClaimInput.Now, claimNow)
	}
	if !outbox.lastClaimInput.ClaimUntil.Equal(claimNow.Add(15 * time.Second)) {
		t.Fatalf("unexpected claim until: got %s want %s", outbox.lastClaimInput.ClaimUntil, claimNow.Add(15*time.Second))
	}
	if notifier.inputs[0].NotificationID != 1 {
		t.Fatalf("unexpected notification id: got %d", notifier.inputs[0].NotificationID)
	}
}

func TestRunReceiptWebhookDispatchCycleUseCaseExecuteRetry(t *testing.T) {
	claimNow := time.Date(2026, 3, 6, 18, 5, 0, 0, time.UTC)
	failedAt := claimNow.Add(7 * time.Second)
	outbox := &fakeWebhookDispatchNotificationOutbox{
		claimRows: []applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage{
			{
				NotificationID:   2,
				PaymentAddressID: 202,
				PreviousStatus:   valueobjects.PaymentReceiptStatusWatching,
				CurrentStatus:    valueobjects.PaymentReceiptStatusPartiallyPaid,
				StatusChangedAt:  claimNow.Add(-time.Minute),
				DeliveryAttempts: 1,
			},
		},
	}
	notifier := &fakeReceiptStatusNotifier{
		errorsByNotificationID: map[int64]error{2: errors.New("timeout")},
	}
	unitOfWork := &fakeReceiptWebhookDispatchUnitOfWork{notificationOutbox: outbox}
	useCase := NewRunReceiptWebhookDispatchCycleUseCase(
		unitOfWork,
		notifier,
		&fakeReceiptWebhookDispatchClock{
			now:   failedAt,
			times: []time.Time{claimNow, failedAt},
		},
	)

	output, err := useCase.Execute(context.Background(), dto.RunReceiptWebhookDispatchCycleInput{
		BatchSize:   10,
		DispatchTTL: 20 * time.Second,
		RetryDelay:  2 * time.Minute,
		MaxAttempts: 5,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if output.RetriedCount != 1 || output.FailedCount != 0 {
		t.Fatalf("unexpected output: %+v", output)
	}
	if got := len(outbox.savedResults); got != 1 {
		t.Fatalf("unexpected saved result count: got %d", got)
	}
	if outbox.savedResults[0].Status != valueobjects.PaymentReceiptNotificationDeliveryStatusPending {
		t.Fatalf("unexpected delivery status: got %q", outbox.savedResults[0].Status)
	}
	if outbox.savedResults[0].Attempts != 2 {
		t.Fatalf("unexpected attempts: got %d", outbox.savedResults[0].Attempts)
	}
	expectedNextAttemptAt := failedAt.Add(2 * time.Minute)
	if outbox.savedResults[0].NextAttemptAt == nil || !outbox.savedResults[0].NextAttemptAt.Equal(expectedNextAttemptAt) {
		t.Fatalf("unexpected next attempt at: got %v want %s", outbox.savedResults[0].NextAttemptAt, expectedNextAttemptAt)
	}
}

func TestRunReceiptWebhookDispatchCycleUseCaseExecuteTerminalFailure(t *testing.T) {
	claimNow := time.Date(2026, 3, 6, 18, 10, 0, 0, time.UTC)
	failedAt := claimNow.Add(5 * time.Second)
	outbox := &fakeWebhookDispatchNotificationOutbox{
		claimRows: []applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage{
			{
				NotificationID:   3,
				PaymentAddressID: 303,
				PreviousStatus:   valueobjects.PaymentReceiptStatusWatching,
				CurrentStatus:    valueobjects.PaymentReceiptStatusFailedExpired,
				StatusChangedAt:  claimNow.Add(-time.Minute),
				DeliveryAttempts: 2,
			},
		},
	}
	notifier := &fakeReceiptStatusNotifier{
		errorsByNotificationID: map[int64]error{3: errors.New("webhook returned status 500")},
	}
	unitOfWork := &fakeReceiptWebhookDispatchUnitOfWork{notificationOutbox: outbox}
	useCase := NewRunReceiptWebhookDispatchCycleUseCase(
		unitOfWork,
		notifier,
		&fakeReceiptWebhookDispatchClock{
			now:   failedAt,
			times: []time.Time{claimNow, failedAt},
		},
	)

	output, err := useCase.Execute(context.Background(), dto.RunReceiptWebhookDispatchCycleInput{
		BatchSize:   10,
		DispatchTTL: 20 * time.Second,
		RetryDelay:  2 * time.Minute,
		MaxAttempts: 3,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if output.FailedCount != 1 || output.RetriedCount != 0 {
		t.Fatalf("unexpected output: %+v", output)
	}
	if got := len(outbox.savedResults); got != 1 {
		t.Fatalf("unexpected saved result count: got %d", got)
	}
	if outbox.savedResults[0].Status != valueobjects.PaymentReceiptNotificationDeliveryStatusFailed {
		t.Fatalf("unexpected delivery status: got %q", outbox.savedResults[0].Status)
	}
	if outbox.savedResults[0].Attempts != 3 {
		t.Fatalf("unexpected attempts: got %d", outbox.savedResults[0].Attempts)
	}
}

func TestRunReceiptWebhookDispatchCycleUseCaseExecuteValidation(t *testing.T) {
	useCase := NewRunReceiptWebhookDispatchCycleUseCase(
		&fakeReceiptWebhookDispatchUnitOfWork{
			notificationOutbox: &fakeWebhookDispatchNotificationOutbox{},
		},
		&fakeReceiptStatusNotifier{errorsByNotificationID: map[int64]error{}},
		&fakeReceiptWebhookDispatchClock{now: time.Now().UTC()},
	)

	_, err := useCase.Execute(context.Background(), dto.RunReceiptWebhookDispatchCycleInput{
		BatchSize:   0,
		RetryDelay:  time.Minute,
		MaxAttempts: 3,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
