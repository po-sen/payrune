package use_cases

import (
	"context"
	"errors"
	"testing"
	"time"

	"payrune/internal/application/dto"
	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/value_objects"
)

type fakeReceiptWebhookDispatchClock struct {
	now time.Time
}

func (f *fakeReceiptWebhookDispatchClock) NowUTC() time.Time {
	return f.now
}

type fakeWebhookDispatchNotificationRepository struct {
	claimRows        []entities.PaymentReceiptStatusNotification
	claimErr         error
	markSentErr      error
	markRetryErr     error
	markFailedErr    error
	lastClaimInput   outport.ClaimPaymentReceiptStatusNotificationsInput
	markSentIDs      []int64
	markSentTimes    []time.Time
	markRetryInputs  []outport.MarkPaymentReceiptStatusNotificationRetryInput
	markFailedInputs []outport.MarkPaymentReceiptStatusNotificationFailureInput
	enqueueInputs    []outport.EnqueuePaymentReceiptStatusChangedInput
	enqueueErr       error
}

func (f *fakeWebhookDispatchNotificationRepository) EnqueueStatusChanged(
	_ context.Context,
	input outport.EnqueuePaymentReceiptStatusChangedInput,
) error {
	f.enqueueInputs = append(f.enqueueInputs, input)
	return f.enqueueErr
}

func (f *fakeWebhookDispatchNotificationRepository) ClaimPending(
	_ context.Context,
	input outport.ClaimPaymentReceiptStatusNotificationsInput,
) ([]entities.PaymentReceiptStatusNotification, error) {
	f.lastClaimInput = input
	if f.claimErr != nil {
		return nil, f.claimErr
	}
	rows := make([]entities.PaymentReceiptStatusNotification, len(f.claimRows))
	copy(rows, f.claimRows)
	return rows, nil
}

func (f *fakeWebhookDispatchNotificationRepository) MarkSent(
	_ context.Context,
	notificationID int64,
	deliveredAt time.Time,
) error {
	f.markSentIDs = append(f.markSentIDs, notificationID)
	f.markSentTimes = append(f.markSentTimes, deliveredAt)
	return f.markSentErr
}

func (f *fakeWebhookDispatchNotificationRepository) MarkRetryScheduled(
	_ context.Context,
	input outport.MarkPaymentReceiptStatusNotificationRetryInput,
) error {
	f.markRetryInputs = append(f.markRetryInputs, input)
	return f.markRetryErr
}

func (f *fakeWebhookDispatchNotificationRepository) MarkFailed(
	_ context.Context,
	input outport.MarkPaymentReceiptStatusNotificationFailureInput,
) error {
	f.markFailedInputs = append(f.markFailedInputs, input)
	return f.markFailedErr
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
	notificationRepository outport.PaymentReceiptStatusNotificationRepository
	err                    error
	calls                  int
}

func (f *fakeReceiptWebhookDispatchUnitOfWork) WithinTransaction(
	_ context.Context,
	fn func(txRepositories outport.TxRepositories) error,
) error {
	f.calls++
	if f.err != nil {
		return f.err
	}
	return fn(outport.TxRepositories{
		PaymentReceiptStatusNotification: f.notificationRepository,
	})
}

func TestRunReceiptWebhookDispatchCycleUseCaseExecuteSuccess(t *testing.T) {
	now := time.Date(2026, 3, 6, 18, 0, 0, 0, time.UTC)
	repository := &fakeWebhookDispatchNotificationRepository{
		claimRows: []entities.PaymentReceiptStatusNotification{
			{
				NotificationID:        1,
				PaymentAddressID:      101,
				CustomerReference:     "order-1",
				PreviousStatus:        value_objects.PaymentReceiptStatusWatching,
				CurrentStatus:         value_objects.PaymentReceiptStatusPaidConfirmed,
				ObservedTotalMinor:    1000,
				ConfirmedTotalMinor:   1000,
				UnconfirmedTotalMinor: 0,
				ConflictTotalMinor:    0,
				StatusChangedAt:       now.Add(-1 * time.Minute),
				DeliveryAttempts:      0,
			},
		},
	}
	notifier := &fakeReceiptStatusNotifier{errorsByNotificationID: map[int64]error{}}
	unitOfWork := &fakeReceiptWebhookDispatchUnitOfWork{notificationRepository: repository}
	useCase := NewRunReceiptWebhookDispatchCycleUseCase(
		unitOfWork,
		notifier,
		&fakeReceiptWebhookDispatchClock{now: now},
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
	if got := len(repository.markSentIDs); got != 1 {
		t.Fatalf("unexpected mark sent count: got %d", got)
	}
	if notifier.inputs[0].NotificationID != 1 {
		t.Fatalf("unexpected notification id: got %d", notifier.inputs[0].NotificationID)
	}
}

func TestRunReceiptWebhookDispatchCycleUseCaseExecuteRetry(t *testing.T) {
	now := time.Date(2026, 3, 6, 18, 5, 0, 0, time.UTC)
	repository := &fakeWebhookDispatchNotificationRepository{
		claimRows: []entities.PaymentReceiptStatusNotification{
			{
				NotificationID:   2,
				PaymentAddressID: 202,
				PreviousStatus:   value_objects.PaymentReceiptStatusWatching,
				CurrentStatus:    value_objects.PaymentReceiptStatusPartiallyPaid,
				StatusChangedAt:  now.Add(-time.Minute),
				DeliveryAttempts: 1,
			},
		},
	}
	notifier := &fakeReceiptStatusNotifier{
		errorsByNotificationID: map[int64]error{2: errors.New("timeout")},
	}
	unitOfWork := &fakeReceiptWebhookDispatchUnitOfWork{notificationRepository: repository}
	useCase := NewRunReceiptWebhookDispatchCycleUseCase(
		unitOfWork,
		notifier,
		&fakeReceiptWebhookDispatchClock{now: now},
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
	if got := len(repository.markRetryInputs); got != 1 {
		t.Fatalf("unexpected retry count: got %d", got)
	}
	if repository.markRetryInputs[0].Attempts != 2 {
		t.Fatalf("unexpected attempts: got %d", repository.markRetryInputs[0].Attempts)
	}
}

func TestRunReceiptWebhookDispatchCycleUseCaseExecuteTerminalFailure(t *testing.T) {
	now := time.Date(2026, 3, 6, 18, 10, 0, 0, time.UTC)
	repository := &fakeWebhookDispatchNotificationRepository{
		claimRows: []entities.PaymentReceiptStatusNotification{
			{
				NotificationID:   3,
				PaymentAddressID: 303,
				PreviousStatus:   value_objects.PaymentReceiptStatusWatching,
				CurrentStatus:    value_objects.PaymentReceiptStatusFailedExpired,
				StatusChangedAt:  now.Add(-time.Minute),
				DeliveryAttempts: 2,
			},
		},
	}
	notifier := &fakeReceiptStatusNotifier{
		errorsByNotificationID: map[int64]error{3: errors.New("webhook returned status 500")},
	}
	unitOfWork := &fakeReceiptWebhookDispatchUnitOfWork{notificationRepository: repository}
	useCase := NewRunReceiptWebhookDispatchCycleUseCase(
		unitOfWork,
		notifier,
		&fakeReceiptWebhookDispatchClock{now: now},
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
	if got := len(repository.markFailedInputs); got != 1 {
		t.Fatalf("unexpected failed count: got %d", got)
	}
	if repository.markFailedInputs[0].Attempts != 3 {
		t.Fatalf("unexpected attempts: got %d", repository.markFailedInputs[0].Attempts)
	}
}

func TestRunReceiptWebhookDispatchCycleUseCaseExecuteValidation(t *testing.T) {
	useCase := NewRunReceiptWebhookDispatchCycleUseCase(
		&fakeReceiptWebhookDispatchUnitOfWork{
			notificationRepository: &fakeWebhookDispatchNotificationRepository{},
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
