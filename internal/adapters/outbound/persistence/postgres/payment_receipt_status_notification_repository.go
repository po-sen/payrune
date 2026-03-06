package postgres

import (
	"context"
	"errors"
	"fmt"

	outport "payrune/internal/application/ports/out"
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

var _ outport.PaymentReceiptStatusNotificationRepository = (*PaymentReceiptStatusNotificationRepository)(nil)
