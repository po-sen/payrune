package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/value_objects"
)

type PaymentReceiptStatusNotificationRepository struct {
	executor Executor
}

func NewPaymentReceiptStatusNotificationRepository(executor Executor) *PaymentReceiptStatusNotificationRepository {
	return &PaymentReceiptStatusNotificationRepository{executor: executor}
}

func (r *PaymentReceiptStatusNotificationRepository) EnqueueStatusChanged(
	ctx context.Context,
	input outport.EnqueuePaymentReceiptStatusChangedInput,
) error {
	if input.PaymentAddressID <= 0 {
		return errors.New("payment address id must be greater than zero")
	}
	if input.PreviousStatus == "" {
		return errors.New("previous status is required")
	}
	if input.CurrentStatus == "" {
		return errors.New("current status is required")
	}
	if input.PreviousStatus == input.CurrentStatus {
		return errors.New("status change is required")
	}
	if input.ObservedTotalMinor < 0 {
		return errors.New("observed total minor must be greater than or equal to zero")
	}
	if input.ConfirmedTotalMinor < 0 {
		return errors.New("confirmed total minor must be greater than or equal to zero")
	}
	if input.UnconfirmedTotalMinor < 0 {
		return errors.New("unconfirmed total minor must be greater than or equal to zero")
	}
	if input.ConflictTotalMinor < 0 {
		return errors.New("conflict total minor must be greater than or equal to zero")
	}
	if input.StatusChangedAt.IsZero() {
		return errors.New("status changed at is required")
	}

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
		     conflict_total_minor,
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
		          $8,
		          'pending'
		   FROM address_policy_allocations a
		   WHERE a.id = $1`,
		input.PaymentAddressID,
		string(input.PreviousStatus),
		string(input.CurrentStatus),
		input.ObservedTotalMinor,
		input.ConfirmedTotalMinor,
		input.UnconfirmedTotalMinor,
		input.ConflictTotalMinor,
		input.StatusChangedAt.UTC(),
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("payment address allocation is not found: %d", input.PaymentAddressID)
	}

	return nil
}

func (r *PaymentReceiptStatusNotificationRepository) ClaimPending(
	ctx context.Context,
	input outport.ClaimPaymentReceiptStatusNotificationsInput,
) ([]entities.PaymentReceiptStatusNotification, error) {
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
		     n.conflict_total_minor,
		     n.status_changed_at,
		     n.delivery_status,
		     n.delivery_attempts,
		     n.next_attempt_at,
		     COALESCE(n.last_error, ''),
		     n.delivered_at`,
		input.Now,
		input.Limit,
		input.ClaimUntil,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make([]entities.PaymentReceiptStatusNotification, 0, input.Limit)
	for rows.Next() {
		notification, err := scanPaymentReceiptStatusNotification(rows)
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

func (r *PaymentReceiptStatusNotificationRepository) MarkSent(
	ctx context.Context,
	notificationID int64,
	deliveredAt time.Time,
) error {
	if notificationID <= 0 {
		return errors.New("notification id must be greater than zero")
	}
	if deliveredAt.IsZero() {
		return errors.New("delivered at is required")
	}

	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE payment_receipt_status_notifications
		 SET delivery_status = 'sent',
		     delivered_at = $2,
		     lease_until = NULL,
		     last_error = NULL,
		     updated_at = NOW()
		 WHERE id = $1`,
		notificationID,
		deliveredAt.UTC(),
	)
	if err != nil {
		return err
	}
	return ensureNotificationRowsAffected(result)
}

func (r *PaymentReceiptStatusNotificationRepository) MarkRetryScheduled(
	ctx context.Context,
	input outport.MarkPaymentReceiptStatusNotificationRetryInput,
) error {
	if input.NotificationID <= 0 {
		return errors.New("notification id must be greater than zero")
	}
	if input.Attempts <= 0 {
		return errors.New("attempts must be greater than zero")
	}
	if strings.TrimSpace(input.LastError) == "" {
		return errors.New("last error is required")
	}
	if input.NextAttemptAt.IsZero() {
		return errors.New("next attempt at is required")
	}

	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE payment_receipt_status_notifications
		 SET delivery_status = 'pending',
		     delivery_attempts = $2,
		     next_attempt_at = $3,
		     lease_until = NULL,
		     last_error = $4,
		     updated_at = NOW()
		 WHERE id = $1`,
		input.NotificationID,
		input.Attempts,
		input.NextAttemptAt.UTC(),
		strings.TrimSpace(input.LastError),
	)
	if err != nil {
		return err
	}
	return ensureNotificationRowsAffected(result)
}

func (r *PaymentReceiptStatusNotificationRepository) MarkFailed(
	ctx context.Context,
	input outport.MarkPaymentReceiptStatusNotificationFailureInput,
) error {
	if input.NotificationID <= 0 {
		return errors.New("notification id must be greater than zero")
	}
	if input.Attempts <= 0 {
		return errors.New("attempts must be greater than zero")
	}
	if strings.TrimSpace(input.LastError) == "" {
		return errors.New("last error is required")
	}

	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE payment_receipt_status_notifications
		 SET delivery_status = 'failed',
		     delivery_attempts = $2,
		     lease_until = NULL,
		     last_error = $3,
		     updated_at = NOW()
		 WHERE id = $1`,
		input.NotificationID,
		input.Attempts,
		strings.TrimSpace(input.LastError),
	)
	if err != nil {
		return err
	}
	return ensureNotificationRowsAffected(result)
}

func scanPaymentReceiptStatusNotification(scanner interface {
	Scan(dest ...any) error
}) (entities.PaymentReceiptStatusNotification, error) {
	var (
		notificationID        int64
		paymentAddressID      int64
		customerReference     string
		previousStatusRaw     string
		currentStatusRaw      string
		observedTotalMinor    int64
		confirmedTotalMinor   int64
		unconfirmedTotalMinor int64
		conflictTotalMinor    int64
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
		&conflictTotalMinor,
		&statusChangedAt,
		&deliveryStatusRaw,
		&deliveryAttempts,
		&nextAttemptAt,
		&lastError,
		&deliveredAt,
	); err != nil {
		return entities.PaymentReceiptStatusNotification{}, err
	}

	previousStatus, ok := value_objects.ParsePaymentReceiptStatus(previousStatusRaw)
	if !ok {
		return entities.PaymentReceiptStatusNotification{}, fmt.Errorf("unsupported previous status: %s", previousStatusRaw)
	}
	currentStatus, ok := value_objects.ParsePaymentReceiptStatus(currentStatusRaw)
	if !ok {
		return entities.PaymentReceiptStatusNotification{}, fmt.Errorf("unsupported current status: %s", currentStatusRaw)
	}
	deliveryStatus, ok := value_objects.ParsePaymentReceiptNotificationDeliveryStatus(deliveryStatusRaw)
	if !ok {
		return entities.PaymentReceiptStatusNotification{}, fmt.Errorf("unsupported delivery status: %s", deliveryStatusRaw)
	}

	notification := entities.PaymentReceiptStatusNotification{
		NotificationID:        notificationID,
		PaymentAddressID:      paymentAddressID,
		CustomerReference:     customerReference,
		PreviousStatus:        previousStatus,
		CurrentStatus:         currentStatus,
		ObservedTotalMinor:    observedTotalMinor,
		ConfirmedTotalMinor:   confirmedTotalMinor,
		UnconfirmedTotalMinor: unconfirmedTotalMinor,
		ConflictTotalMinor:    conflictTotalMinor,
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

func ensureNotificationRowsAffected(result sql.Result) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("payment receipt status notification is not found")
	}
	return nil
}

var _ outport.PaymentReceiptStatusNotificationRepository = (*PaymentReceiptStatusNotificationRepository)(nil)
