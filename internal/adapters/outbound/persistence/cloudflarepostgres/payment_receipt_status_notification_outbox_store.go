package cloudflarepostgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	applicationoutbox "payrune/internal/application/outbox"
	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/events"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/value_objects"
)

type PaymentReceiptStatusNotificationOutboxStore struct {
	executor Executor
}

func NewPaymentReceiptStatusNotificationOutboxStore(executor Executor) *PaymentReceiptStatusNotificationOutboxStore {
	return &PaymentReceiptStatusNotificationOutboxStore{executor: executor}
}

func (r *PaymentReceiptStatusNotificationOutboxStore) EnqueueStatusChanged(
	ctx context.Context,
	event events.PaymentReceiptStatusChanged,
) error {
	result, err := r.executor.ExecContext(
		ctx,
		`INSERT INTO payment_receipt_status_notifications (
		     payment_address_id,
		     customer_reference,
		     previous_status,
		     current_status,
		     observed_total_minor,
		     confirmed_total_minor,
		     unconfirmed_total_minor,
		     status_changed_at,
		     delivery_status
		   )
		   SELECT a.id,
		          a.customer_reference,
		          $2,
		          $3,
		          $4,
		          $5,
		          $6,
		          $7,
		          'pending'
		   FROM address_policy_allocations a
		   WHERE a.id = $1`,
		event.PaymentAddressID,
		string(event.PreviousStatus),
		string(event.CurrentStatus),
		event.ObservedTotalMinor,
		event.ConfirmedTotalMinor,
		event.UnconfirmedTotalMinor,
		event.StatusChangedAt.UTC(),
	)
	if err != nil {
		return err
	}

	return ensureNotificationRowsAffected(result)
}

func (r *PaymentReceiptStatusNotificationOutboxStore) ClaimPending(
	ctx context.Context,
	input outport.ClaimPaymentReceiptStatusNotificationsInput,
) ([]applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage, error) {
	if input.Now.IsZero() {
		return nil, errors.New("claim now is required")
	}
	if input.ClaimUntil.IsZero() {
		return nil, errors.New("claim until is required")
	}
	if input.Limit <= 0 {
		return nil, errors.New("claim limit must be greater than zero")
	}

	rows, err := r.executor.QueryContext(
		ctx,
		`WITH due AS (
		     SELECT id
		     FROM payment_receipt_status_notifications
		     WHERE delivery_status = 'pending'
		       AND next_attempt_at <= $1
		       AND (lease_until IS NULL OR lease_until <= $1)
		     ORDER BY next_attempt_at ASC, id ASC
		     FOR UPDATE SKIP LOCKED
		     LIMIT $2
		   )
		   UPDATE payment_receipt_status_notifications n
		   SET lease_until = $3,
		       updated_at = NOW()
		   FROM due
		   WHERE n.id = due.id
		   RETURNING
		     n.id,
		     n.payment_address_id,
		     COALESCE(n.customer_reference, ''),
		     n.previous_status,
		     n.current_status,
		     n.observed_total_minor,
		     n.confirmed_total_minor,
		     n.unconfirmed_total_minor,
		     n.status_changed_at,
		     n.delivery_status,
		     n.delivery_attempts,
		     n.next_attempt_at,
		     COALESCE(n.last_error, ''),
		     n.delivered_at`,
		input.Now.UTC(),
		input.Limit,
		input.ClaimUntil.UTC(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make([]applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage, 0, input.Limit)
	for rows.Next() {
		notification, err := scanPaymentReceiptStatusNotificationOutboxMessage(rows)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *PaymentReceiptStatusNotificationOutboxStore) SaveDeliveryResult(
	ctx context.Context,
	result policies.PaymentReceiptStatusNotificationDeliveryResult,
) error {
	switch result.Status {
	case value_objects.PaymentReceiptNotificationDeliveryStatusSent:
		if result.DeliveredAt == nil {
			return errors.New("delivered at is required")
		}
		execResult, err := r.executor.ExecContext(
			ctx,
			`UPDATE payment_receipt_status_notifications
			 SET delivery_status = 'sent',
			     delivered_at = $2,
			     lease_until = NULL,
			     last_error = NULL,
			     updated_at = NOW()
			 WHERE id = $1`,
			result.NotificationID,
			result.DeliveredAt.UTC(),
		)
		if err != nil {
			return err
		}
		return ensureNotificationRowsAffected(execResult)
	case value_objects.PaymentReceiptNotificationDeliveryStatusPending:
		if result.NextAttemptAt == nil {
			return errors.New("next attempt at is required")
		}
		execResult, err := r.executor.ExecContext(
			ctx,
			`UPDATE payment_receipt_status_notifications
			 SET delivery_status = 'pending',
			     delivery_attempts = $2,
			     next_attempt_at = $3,
			     lease_until = NULL,
			     last_error = $4,
			     updated_at = NOW()
			 WHERE id = $1`,
			result.NotificationID,
			result.Attempts,
			result.NextAttemptAt.UTC(),
			result.LastError,
		)
		if err != nil {
			return err
		}
		return ensureNotificationRowsAffected(execResult)
	case value_objects.PaymentReceiptNotificationDeliveryStatusFailed:
		execResult, err := r.executor.ExecContext(
			ctx,
			`UPDATE payment_receipt_status_notifications
			 SET delivery_status = 'failed',
			     delivery_attempts = $2,
			     lease_until = NULL,
			     last_error = $3,
			     updated_at = NOW()
			 WHERE id = $1`,
			result.NotificationID,
			result.Attempts,
			result.LastError,
		)
		if err != nil {
			return err
		}
		return ensureNotificationRowsAffected(execResult)
	default:
		return errors.New("delivery result status is invalid")
	}
}

func scanPaymentReceiptStatusNotificationOutboxMessage(scanner interface {
	Scan(dest ...any) error
}) (applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage, error) {
	var (
		notificationID        int64
		paymentAddressID      int64
		customerReference     string
		previousStatusRaw     string
		currentStatusRaw      string
		observedTotalMinor    int64
		confirmedTotalMinor   int64
		unconfirmedTotalMinor int64
		statusChangedAt       time.Time
		deliveryStatusRaw     string
		deliveryAttempts      int32
		nextAttemptAt         time.Time
		lastError             string
		deliveredAt           sql.NullTime
	)

	if err := scanner.Scan(
		&notificationID,
		&paymentAddressID,
		&customerReference,
		&previousStatusRaw,
		&currentStatusRaw,
		&observedTotalMinor,
		&confirmedTotalMinor,
		&unconfirmedTotalMinor,
		&statusChangedAt,
		&deliveryStatusRaw,
		&deliveryAttempts,
		&nextAttemptAt,
		&lastError,
		&deliveredAt,
	); err != nil {
		return applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage{}, err
	}

	previousStatus, ok := value_objects.ParsePaymentReceiptStatus(previousStatusRaw)
	if !ok {
		return applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage{}, fmt.Errorf("unsupported previous status: %s", previousStatusRaw)
	}
	currentStatus, ok := value_objects.ParsePaymentReceiptStatus(currentStatusRaw)
	if !ok {
		return applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage{}, fmt.Errorf("unsupported current status: %s", currentStatusRaw)
	}
	deliveryStatus, ok := value_objects.ParsePaymentReceiptNotificationDeliveryStatus(deliveryStatusRaw)
	if !ok {
		return applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage{}, fmt.Errorf("unsupported delivery status: %s", deliveryStatusRaw)
	}

	notification := applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage{
		NotificationID:        notificationID,
		PaymentAddressID:      paymentAddressID,
		CustomerReference:     customerReference,
		PreviousStatus:        previousStatus,
		CurrentStatus:         currentStatus,
		ObservedTotalMinor:    observedTotalMinor,
		ConfirmedTotalMinor:   confirmedTotalMinor,
		UnconfirmedTotalMinor: unconfirmedTotalMinor,
		StatusChangedAt:       statusChangedAt.UTC(),
		DeliveryStatus:        deliveryStatus,
		DeliveryAttempts:      deliveryAttempts,
		NextAttemptAt:         nextAttemptAt.UTC(),
		LastError:             lastError,
	}
	if deliveredAt.Valid {
		timeValue := deliveredAt.Time.UTC()
		notification.DeliveredAt = &timeValue
	}

	return notification, nil
}

func ensureNotificationRowsAffected(result Result) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("payment receipt status notification is not found")
	}
	return nil
}

var _ outport.PaymentReceiptStatusNotificationOutbox = (*PaymentReceiptStatusNotificationOutboxStore)(nil)
