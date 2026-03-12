package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/events"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
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

func TestPaymentReceiptStatusNotificationOutboxEnqueueStatusChangedSuccess(t *testing.T) {
	now := time.Date(2026, 3, 6, 9, 30, 0, 0, time.UTC)
	executor := &stubNotificationExecutor{
		execResult: stubSQLResult{rowsAffected: 1},
	}
	outboxStore := NewPaymentReceiptStatusNotificationOutboxStore(executor)

	err := outboxStore.EnqueueStatusChanged(context.Background(), events.PaymentReceiptStatusChanged{
		PaymentAddressID:      101,
		PreviousStatus:        valueobjects.PaymentReceiptStatusWatching,
		CurrentStatus:         valueobjects.PaymentReceiptStatusPaidUnconfirmed,
		ObservedTotalMinor:    1000,
		ConfirmedTotalMinor:   0,
		UnconfirmedTotalMinor: 1000,
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
	if got := len(executor.lastArgs); got != 7 {
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
	statusChangedAt, ok := executor.lastArgs[6].(time.Time)
	if !ok {
		t.Fatalf("unexpected status changed at type: %T", executor.lastArgs[6])
	}
	if !statusChangedAt.Equal(now) {
		t.Fatalf("unexpected status changed at arg: got %s want %s", statusChangedAt, now)
	}
}

func TestPaymentReceiptStatusNotificationOutboxEnqueueStatusChangedSupportsRevertedStatus(t *testing.T) {
	now := time.Date(2026, 3, 6, 9, 35, 0, 0, time.UTC)
	executor := &stubNotificationExecutor{
		execResult: stubSQLResult{rowsAffected: 1},
	}
	outboxStore := NewPaymentReceiptStatusNotificationOutboxStore(executor)

	err := outboxStore.EnqueueStatusChanged(context.Background(), events.PaymentReceiptStatusChanged{
		PaymentAddressID:      102,
		PreviousStatus:        valueobjects.PaymentReceiptStatusPaidUnconfirmed,
		CurrentStatus:         valueobjects.PaymentReceiptStatusPaidUnconfirmedReverted,
		ObservedTotalMinor:    400,
		ConfirmedTotalMinor:   0,
		UnconfirmedTotalMinor: 400,
		StatusChangedAt:       now,
	})
	if err != nil {
		t.Fatalf("EnqueueStatusChanged returned error: %v", err)
	}
	if got := executor.lastArgs[1]; got != "paid_unconfirmed" {
		t.Fatalf("unexpected previous status arg: got %#v", got)
	}
	if got := executor.lastArgs[2]; got != "paid_unconfirmed_reverted" {
		t.Fatalf("unexpected current status arg: got %#v", got)
	}
}

func TestPaymentReceiptStatusNotificationOutboxEnqueueStatusChangedAddressNotFound(t *testing.T) {
	outboxStore := NewPaymentReceiptStatusNotificationOutboxStore(&stubNotificationExecutor{
		execResult: stubSQLResult{rowsAffected: 0},
	})

	err := outboxStore.EnqueueStatusChanged(context.Background(), events.PaymentReceiptStatusChanged{
		PaymentAddressID:      88,
		PreviousStatus:        valueobjects.PaymentReceiptStatusWatching,
		CurrentStatus:         valueobjects.PaymentReceiptStatusPaidConfirmed,
		ObservedTotalMinor:    100,
		ConfirmedTotalMinor:   100,
		UnconfirmedTotalMinor: 0,
		StatusChangedAt:       time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestPaymentReceiptStatusNotificationOutboxEnqueueStatusChangedExecError(t *testing.T) {
	outboxStore := NewPaymentReceiptStatusNotificationOutboxStore(&stubNotificationExecutor{
		execErr: errors.New("db down"),
	})

	err := outboxStore.EnqueueStatusChanged(context.Background(), events.PaymentReceiptStatusChanged{
		PaymentAddressID:      88,
		PreviousStatus:        valueobjects.PaymentReceiptStatusWatching,
		CurrentStatus:         valueobjects.PaymentReceiptStatusPaidConfirmed,
		ObservedTotalMinor:    100,
		ConfirmedTotalMinor:   100,
		UnconfirmedTotalMinor: 0,
		StatusChangedAt:       time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("expected exec error")
	}
}

func TestPaymentReceiptStatusNotificationOutboxClaimPendingValidation(t *testing.T) {
	outboxStore := NewPaymentReceiptStatusNotificationOutboxStore(&stubNotificationExecutor{})

	_, err := outboxStore.ClaimPending(context.Background(), outport.ClaimPaymentReceiptStatusNotificationsInput{
		Now:        time.Time{},
		Limit:      1,
		ClaimUntil: time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("expected missing now error")
	}

	_, err = outboxStore.ClaimPending(context.Background(), outport.ClaimPaymentReceiptStatusNotificationsInput{
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

	notification, err := scanPaymentReceiptStatusNotificationOutboxMessage(stubScanner{
		values: []any{
			int64(1),
			int64(11),
			"order-1",
			"watching",
			"paid_confirmed",
			int64(1000),
			int64(1000),
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
		t.Fatalf("scanPaymentReceiptStatusNotificationOutboxMessage returned error: %v", err)
	}
	if notification.DeliveryStatus != valueobjects.PaymentReceiptNotificationDeliveryStatusSent {
		t.Fatalf("unexpected delivery status: got %q", notification.DeliveryStatus)
	}
	if notification.DeliveredAt == nil || !notification.DeliveredAt.Equal(deliveredAt) {
		t.Fatalf("unexpected delivered at: got %+v", notification.DeliveredAt)
	}
}

func TestScanPaymentReceiptStatusNotificationSupportsRevertedStatus(t *testing.T) {
	now := time.Date(2026, 3, 6, 16, 10, 0, 0, time.UTC)

	notification, err := scanPaymentReceiptStatusNotificationOutboxMessage(stubScanner{
		values: []any{
			int64(2),
			int64(12),
			"order-2",
			"paid_unconfirmed",
			"paid_unconfirmed_reverted",
			int64(400),
			int64(0),
			int64(400),
			now,
			"pending",
			int32(0),
			now.Add(5 * time.Minute),
			"",
			sql.NullTime{},
		},
	})
	if err != nil {
		t.Fatalf("scanPaymentReceiptStatusNotificationOutboxMessage returned error: %v", err)
	}
	if notification.PreviousStatus != valueobjects.PaymentReceiptStatusPaidUnconfirmed {
		t.Fatalf("unexpected previous status: got %q", notification.PreviousStatus)
	}
	if notification.CurrentStatus != valueobjects.PaymentReceiptStatusPaidUnconfirmedReverted {
		t.Fatalf("unexpected current status: got %q", notification.CurrentStatus)
	}
}

func TestPaymentReceiptStatusNotificationOutboxSaveDeliveryResultValidation(t *testing.T) {
	outboxStore := NewPaymentReceiptStatusNotificationOutboxStore(&stubNotificationExecutor{})

	err := outboxStore.SaveDeliveryResult(
		context.Background(),
		policies.PaymentReceiptStatusNotificationDeliveryResult{
			Status: valueobjects.PaymentReceiptNotificationDeliveryStatusSent,
		},
	)
	if err == nil {
		t.Fatal("expected delivered at validation error")
	}
}

func TestScanPaymentReceiptStatusNotificationOutboxMessageRejectsUnsupportedStatuses(t *testing.T) {
	_, err := scanPaymentReceiptStatusNotificationOutboxMessage(stubScanner{
		values: []any{
			int64(1),
			int64(11),
			"order-1",
			"invalid",
			"paid_confirmed",
			int64(1000),
			int64(1000),
			int64(0),
			time.Date(2026, 3, 6, 16, 0, 0, 0, time.UTC),
			"pending",
			int32(0),
			time.Date(2026, 3, 6, 16, 5, 0, 0, time.UTC),
			"",
			sql.NullTime{},
		},
	})
	if err == nil {
		t.Fatal("expected invalid status error")
	}
}

func TestPaymentReceiptStatusNotificationOutboxSaveDeliveryResultPendingSuccess(t *testing.T) {
	executor := &stubNotificationExecutor{
		execResult: stubSQLResult{rowsAffected: 1},
	}
	outboxStore := NewPaymentReceiptStatusNotificationOutboxStore(executor)
	nextAttemptAt := time.Date(2026, 3, 6, 17, 0, 0, 0, time.UTC)

	err := outboxStore.SaveDeliveryResult(
		context.Background(),
		policies.PaymentReceiptStatusNotificationDeliveryResult{
			NotificationID: 99,
			Status:         valueobjects.PaymentReceiptNotificationDeliveryStatusPending,
			Attempts:       2,
			LastError:      "timeout",
			NextAttemptAt:  &nextAttemptAt,
		},
	)
	if err != nil {
		t.Fatalf("SaveDeliveryResult returned error: %v", err)
	}
	if !strings.Contains(executor.lastQuery, "delivery_status = 'pending'") {
		t.Fatalf("unexpected query: %s", executor.lastQuery)
	}
}

func TestPaymentReceiptStatusNotificationOutboxSaveDeliveryResultFailedSuccess(t *testing.T) {
	executor := &stubNotificationExecutor{
		execResult: stubSQLResult{rowsAffected: 1},
	}
	outboxStore := NewPaymentReceiptStatusNotificationOutboxStore(executor)

	err := outboxStore.SaveDeliveryResult(
		context.Background(),
		policies.PaymentReceiptStatusNotificationDeliveryResult{
			NotificationID: 99,
			Status:         valueobjects.PaymentReceiptNotificationDeliveryStatusFailed,
			Attempts:       3,
			LastError:      "webhook returned status 500",
		},
	)
	if err != nil {
		t.Fatalf("SaveDeliveryResult returned error: %v", err)
	}
	if !strings.Contains(executor.lastQuery, "delivery_status = 'failed'") {
		t.Fatalf("unexpected query: %s", executor.lastQuery)
	}
}
