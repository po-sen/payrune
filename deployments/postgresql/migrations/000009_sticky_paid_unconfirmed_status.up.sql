DROP INDEX IF EXISTS idx_payment_receipt_trackings_active_expires_at;
CREATE INDEX idx_payment_receipt_trackings_active_expires_at
  ON payment_receipt_trackings (expires_at ASC)
  WHERE receipt_status IN ('watching', 'partially_paid')
    AND paid_at IS NULL
    AND expires_at IS NOT NULL;

DROP INDEX IF EXISTS idx_payment_receipt_trackings_active_lease_until;
CREATE INDEX idx_payment_receipt_trackings_active_lease_until
  ON payment_receipt_trackings (lease_until ASC)
  WHERE receipt_status IN (
      'watching',
      'partially_paid',
      'paid_unconfirmed',
      'paid_unconfirmed_reverted'
    )
    AND lease_until IS NOT NULL;

ALTER TABLE payment_receipt_trackings
  DROP CONSTRAINT IF EXISTS payment_receipt_trackings_receipt_status_check;

ALTER TABLE payment_receipt_trackings
  ADD CONSTRAINT payment_receipt_trackings_receipt_status_check
  CHECK (
    receipt_status IN (
      'watching',
      'partially_paid',
      'paid_unconfirmed',
      'paid_unconfirmed_reverted',
      'paid_confirmed',
      'failed_expired'
    )
  );

ALTER TABLE payment_receipt_trackings
  DROP COLUMN IF EXISTS conflict_total_minor;

ALTER TABLE payment_receipt_status_notifications
  DROP CONSTRAINT IF EXISTS payment_receipt_status_notifications_previous_status_check;

ALTER TABLE payment_receipt_status_notifications
  ADD CONSTRAINT payment_receipt_status_notifications_previous_status_check
  CHECK (
    previous_status IN (
      'watching',
      'partially_paid',
      'paid_unconfirmed',
      'paid_unconfirmed_reverted',
      'paid_confirmed',
      'failed_expired'
    )
  );

ALTER TABLE payment_receipt_status_notifications
  DROP CONSTRAINT IF EXISTS payment_receipt_status_notifications_current_status_check;

ALTER TABLE payment_receipt_status_notifications
  ADD CONSTRAINT payment_receipt_status_notifications_current_status_check
  CHECK (
    current_status IN (
      'watching',
      'partially_paid',
      'paid_unconfirmed',
      'paid_unconfirmed_reverted',
      'paid_confirmed',
      'failed_expired'
    )
  );

ALTER TABLE payment_receipt_status_notifications
  DROP COLUMN IF EXISTS conflict_total_minor;
