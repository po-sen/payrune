DROP INDEX IF EXISTS idx_payment_receipt_status_notifications_pending_lease_until;
DROP INDEX IF EXISTS idx_payment_receipt_status_notifications_pending_next_attempt;

ALTER TABLE payment_receipt_status_notifications
  DROP CONSTRAINT IF EXISTS payment_receipt_status_notifications_delivery_attempts_check;

ALTER TABLE payment_receipt_status_notifications
  DROP COLUMN IF EXISTS updated_at;

ALTER TABLE payment_receipt_status_notifications
  DROP COLUMN IF EXISTS delivered_at;

ALTER TABLE payment_receipt_status_notifications
  DROP COLUMN IF EXISTS last_error;

ALTER TABLE payment_receipt_status_notifications
  DROP COLUMN IF EXISTS lease_until;

ALTER TABLE payment_receipt_status_notifications
  DROP COLUMN IF EXISTS next_attempt_at;

ALTER TABLE payment_receipt_status_notifications
  DROP COLUMN IF EXISTS delivery_attempts;
