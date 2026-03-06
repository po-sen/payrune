UPDATE payment_receipt_trackings
SET receipt_status = 'watching'
WHERE receipt_status = 'failed_expired';

DROP INDEX IF EXISTS idx_payment_receipt_trackings_active_expires_at;
DROP INDEX IF EXISTS idx_payment_receipt_trackings_active_lease_until;

ALTER TABLE payment_receipt_trackings
  DROP CONSTRAINT IF EXISTS payment_receipt_trackings_receipt_status_check;

ALTER TABLE payment_receipt_trackings
  ADD CONSTRAINT payment_receipt_trackings_receipt_status_check
  CHECK (
    receipt_status IN (
      'watching',
      'partially_paid',
      'paid_unconfirmed',
      'paid_confirmed',
      'double_spend_suspected'
    )
  );

ALTER TABLE payment_receipt_trackings
  DROP COLUMN IF EXISTS lease_until;

ALTER TABLE payment_receipt_trackings
  DROP COLUMN IF EXISTS expires_at;
