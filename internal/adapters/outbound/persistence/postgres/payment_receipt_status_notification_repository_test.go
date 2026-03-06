package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/value_objects"
)

type stubSQLResult struct {
	lastInsertID int64
	rowsAffected int64
	err          error
}

func (r stubSQLResult) LastInsertId() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return r.lastInsertID, nil
}

func (r stubSQLResult) RowsAffected() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return r.rowsAffected, nil
}

type stubNotificationExecutor struct {
	execResult sql.Result
	execErr    error
	queryErr   error
	lastQuery  string
	lastArgs   []any
}

func (s *stubNotificationExecutor) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	s.lastQuery = query
	s.lastArgs = append([]any(nil), args...)
	if s.execErr != nil {
		return nil, s.execErr
	}
	if s.execResult == nil {
		return stubSQLResult{rowsAffected: 1}, nil
	}
	return s.execResult, nil
}

func (s *stubNotificationExecutor) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	if s.queryErr != nil {
		return nil, s.queryErr
	}
	panic("unexpected QueryContext call")
}

func (s *stubNotificationExecutor) QueryRowContext(context.Context, string, ...any) *sql.Row {
	panic("unexpected QueryRowContext call")
}

func TestPaymentReceiptStatusNotificationRepositoryEnqueueStatusChangedValidation(t *testing.T) {
	testCases := []struct {
		name  string
		input outport.EnqueuePaymentReceiptStatusChangedInput
	}{
		{
			name: "invalid payment address id",
			input: outport.EnqueuePaymentReceiptStatusChangedInput{
				PaymentAddressID: 0,
			},
		},
		{
			name: "same status",
			input: outport.EnqueuePaymentReceiptStatusChangedInput{
				PaymentAddressID: 1,
				PreviousStatus:   value_objects.PaymentReceiptStatusWatching,
				CurrentStatus:    value_objects.PaymentReceiptStatusWatching,
				StatusChangedAt:  time.Now().UTC(),
			},
		},
		{
			name: "negative amount",
			input: outport.EnqueuePaymentReceiptStatusChangedInput{
				PaymentAddressID:   1,
				PreviousStatus:     value_objects.PaymentReceiptStatusWatching,
				CurrentStatus:      value_objects.PaymentReceiptStatusPaidConfirmed,
				ObservedTotalMinor: -1,
				StatusChangedAt:    time.Now().UTC(),
			},
		},
		{
			name: "missing changed at",
			input: outport.EnqueuePaymentReceiptStatusChangedInput{
				PaymentAddressID: 1,
				PreviousStatus:   value_objects.PaymentReceiptStatusWatching,
				CurrentStatus:    value_objects.PaymentReceiptStatusPaidConfirmed,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := NewPaymentReceiptStatusNotificationRepository(&stubNotificationExecutor{})

			err := repository.EnqueueStatusChanged(context.Background(), tc.input)
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestPaymentReceiptStatusNotificationRepositoryEnqueueStatusChangedSuccess(t *testing.T) {
	now := time.Date(2026, 3, 6, 9, 30, 0, 0, time.UTC)
	executor := &stubNotificationExecutor{
		execResult: stubSQLResult{rowsAffected: 1},
	}
	repository := NewPaymentReceiptStatusNotificationRepository(executor)

	err := repository.EnqueueStatusChanged(context.Background(), outport.EnqueuePaymentReceiptStatusChangedInput{
		PaymentAddressID:      101,
		PreviousStatus:        value_objects.PaymentReceiptStatusWatching,
		CurrentStatus:         value_objects.PaymentReceiptStatusPaidUnconfirmed,
		ObservedTotalMinor:    1000,
		ConfirmedTotalMinor:   0,
		UnconfirmedTotalMinor: 1000,
		ConflictTotalMinor:    0,
		StatusChangedAt:       now,
	})
	if err != nil {
		t.Fatalf("EnqueueStatusChanged returned error: %v", err)
	}
	if executor.lastQuery == "" {
		t.Fatal("expected SQL query to be executed")
	}
	if !strings.Contains(executor.lastQuery, "INSERT INTO payment_receipt_status_notifications") {
		t.Fatalf("unexpected query: %s", executor.lastQuery)
	}
	if got := len(executor.lastArgs); got != 8 {
		t.Fatalf("unexpected arg count: got %d", got)
	}
	if got := executor.lastArgs[0]; got != int64(101) {
		t.Fatalf("unexpected payment address id arg: got %#v", got)
	}
	if got := executor.lastArgs[1]; got != "watching" {
		t.Fatalf("unexpected previous status arg: got %#v", got)
	}
	if got := executor.lastArgs[2]; got != "paid_unconfirmed" {
		t.Fatalf("unexpected current status arg: got %#v", got)
	}
	statusChangedAt, ok := executor.lastArgs[7].(time.Time)
	if !ok {
		t.Fatalf("unexpected status changed at type: %T", executor.lastArgs[7])
	}
	if !statusChangedAt.Equal(now) {
		t.Fatalf("unexpected status changed at arg: got %s want %s", statusChangedAt, now)
	}
}

func TestPaymentReceiptStatusNotificationRepositoryEnqueueStatusChangedAddressNotFound(t *testing.T) {
	repository := NewPaymentReceiptStatusNotificationRepository(&stubNotificationExecutor{
		execResult: stubSQLResult{rowsAffected: 0},
	})

	err := repository.EnqueueStatusChanged(context.Background(), outport.EnqueuePaymentReceiptStatusChangedInput{
		PaymentAddressID:      88,
		PreviousStatus:        value_objects.PaymentReceiptStatusWatching,
		CurrentStatus:         value_objects.PaymentReceiptStatusPaidConfirmed,
		ObservedTotalMinor:    100,
		ConfirmedTotalMinor:   100,
		UnconfirmedTotalMinor: 0,
		ConflictTotalMinor:    0,
		StatusChangedAt:       time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestPaymentReceiptStatusNotificationRepositoryEnqueueStatusChangedExecError(t *testing.T) {
	repository := NewPaymentReceiptStatusNotificationRepository(&stubNotificationExecutor{
		execErr: errors.New("db down"),
	})

	err := repository.EnqueueStatusChanged(context.Background(), outport.EnqueuePaymentReceiptStatusChangedInput{
		PaymentAddressID:      88,
		PreviousStatus:        value_objects.PaymentReceiptStatusWatching,
		CurrentStatus:         value_objects.PaymentReceiptStatusPaidConfirmed,
		ObservedTotalMinor:    100,
		ConfirmedTotalMinor:   100,
		UnconfirmedTotalMinor: 0,
		ConflictTotalMinor:    0,
		StatusChangedAt:       time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("expected exec error")
	}
}

func TestPaymentReceiptStatusNotificationRepositoryClaimPendingValidation(t *testing.T) {
	repository := NewPaymentReceiptStatusNotificationRepository(&stubNotificationExecutor{})

	_, err := repository.ClaimPending(context.Background(), outport.ClaimPaymentReceiptStatusNotificationsInput{
		Now:        time.Time{},
		Limit:      1,
		ClaimUntil: time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("expected missing now error")
	}

	_, err = repository.ClaimPending(context.Background(), outport.ClaimPaymentReceiptStatusNotificationsInput{
		Now:        time.Now().UTC(),
		Limit:      0,
		ClaimUntil: time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("expected invalid limit error")
	}
}

func TestScanPaymentReceiptStatusNotificationSupportsDeliveryFields(t *testing.T) {
	now := time.Date(2026, 3, 6, 16, 0, 0, 0, time.UTC)
	deliveredAt := now.Add(2 * time.Minute)

	notification, err := scanPaymentReceiptStatusNotification(stubScanner{
		values: []any{
			int64(1),
			int64(11),
			"order-1",
			"watching",
			"paid_confirmed",
			int64(1000),
			int64(1000),
			int64(0),
			int64(0),
			now,
			"sent",
			int32(2),
			now.Add(5 * time.Minute),
			"",
			sql.NullTime{Valid: true, Time: deliveredAt},
		},
	})
	if err != nil {
		t.Fatalf("scanPaymentReceiptStatusNotification returned error: %v", err)
	}
	if notification.DeliveryStatus != value_objects.PaymentReceiptNotificationDeliveryStatusSent {
		t.Fatalf("unexpected delivery status: got %q", notification.DeliveryStatus)
	}
	if notification.DeliveredAt == nil || !notification.DeliveredAt.Equal(deliveredAt) {
		t.Fatalf("unexpected delivered at: got %+v", notification.DeliveredAt)
	}
}

func TestPaymentReceiptStatusNotificationRepositoryMarkSentValidation(t *testing.T) {
	repository := NewPaymentReceiptStatusNotificationRepository(&stubNotificationExecutor{})

	err := repository.MarkSent(context.Background(), 0, time.Now().UTC())
	if err == nil {
		t.Fatal("expected invalid id error")
	}
}

func TestPaymentReceiptStatusNotificationRepositoryMarkRetryScheduledSuccess(t *testing.T) {
	executor := &stubNotificationExecutor{
		execResult: stubSQLResult{rowsAffected: 1},
	}
	repository := NewPaymentReceiptStatusNotificationRepository(executor)
	nextAttemptAt := time.Date(2026, 3, 6, 17, 0, 0, 0, time.UTC)

	err := repository.MarkRetryScheduled(context.Background(), outport.MarkPaymentReceiptStatusNotificationRetryInput{
		NotificationID: 99,
		Attempts:       2,
		LastError:      "timeout",
		NextAttemptAt:  nextAttemptAt,
	})
	if err != nil {
		t.Fatalf("MarkRetryScheduled returned error: %v", err)
	}
	if !strings.Contains(executor.lastQuery, "delivery_status = 'pending'") {
		t.Fatalf("unexpected query: %s", executor.lastQuery)
	}
}

func TestPaymentReceiptStatusNotificationRepositoryMarkFailedSuccess(t *testing.T) {
	executor := &stubNotificationExecutor{
		execResult: stubSQLResult{rowsAffected: 1},
	}
	repository := NewPaymentReceiptStatusNotificationRepository(executor)

	err := repository.MarkFailed(context.Background(), outport.MarkPaymentReceiptStatusNotificationFailureInput{
		NotificationID: 99,
		Attempts:       3,
		LastError:      "webhook returned status 500",
	})
	if err != nil {
		t.Fatalf("MarkFailed returned error: %v", err)
	}
	if !strings.Contains(executor.lastQuery, "delivery_status = 'failed'") {
		t.Fatalf("unexpected query: %s", executor.lastQuery)
	}
}
