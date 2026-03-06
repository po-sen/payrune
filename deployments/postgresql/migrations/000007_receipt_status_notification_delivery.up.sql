ALTER TABLE payment_receipt_status_notifications
  ADD COLUMN IF NOT EXISTS delivery_attempts INTEGER NOT NULL DEFAULT 0;

ALTER TABLE payment_receipt_status_notifications
  ADD COLUMN IF NOT EXISTS next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE payment_receipt_status_notifications
  ADD COLUMN IF NOT EXISTS lease_until TIMESTAMPTZ;

ALTER TABLE payment_receipt_status_notifications
  ADD COLUMN IF NOT EXISTS last_error TEXT;

ALTER TABLE payment_receipt_status_notifications
  ADD COLUMN IF NOT EXISTS delivered_at TIMESTAMPTZ;

ALTER TABLE payment_receipt_status_notifications
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE payment_receipt_status_notifications
  DROP CONSTRAINT IF EXISTS payment_receipt_status_notifications_delivery_attempts_check;

ALTER TABLE payment_receipt_status_notifications
  ADD CONSTRAINT payment_receipt_status_notifications_delivery_attempts_check
  CHECK (delivery_attempts >= 0);

DROP INDEX IF EXISTS idx_payment_receipt_status_notifications_pending_next_attempt;
CREATE INDEX idx_payment_receipt_status_notifications_pending_next_attempt
  ON payment_receipt_status_notifications (next_attempt_at ASC)
  WHERE delivery_status = 'pending';

DROP INDEX IF EXISTS idx_payment_receipt_status_notifications_pending_lease_until;
CREATE INDEX idx_payment_receipt_status_notifications_pending_lease_until
  ON payment_receipt_status_notifications (lease_until ASC)
  WHERE delivery_status = 'pending'
    AND lease_until IS NOT NULL;
