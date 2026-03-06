ALTER TABLE payment_receipt_trackings
  ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ;

ALTER TABLE payment_receipt_trackings
  ADD COLUMN IF NOT EXISTS lease_until TIMESTAMPTZ;

UPDATE payment_receipt_trackings
SET expires_at = COALESCE(issued_at, created_at) + INTERVAL '7 days'
WHERE expires_at IS NULL;

ALTER TABLE payment_receipt_trackings
  ALTER COLUMN expires_at SET NOT NULL;

DROP INDEX IF EXISTS idx_payment_receipt_trackings_active_expires_at;
CREATE INDEX idx_payment_receipt_trackings_active_expires_at
  ON payment_receipt_trackings (expires_at ASC)
  WHERE receipt_status IN ('watching', 'partially_paid', 'paid_unconfirmed', 'double_spend_suspected')
    AND expires_at IS NOT NULL;

DROP INDEX IF EXISTS idx_payment_receipt_trackings_active_lease_until;
CREATE INDEX idx_payment_receipt_trackings_active_lease_until
  ON payment_receipt_trackings (lease_until ASC)
  WHERE receipt_status IN ('watching', 'partially_paid', 'paid_unconfirmed', 'double_spend_suspected')
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
      'paid_confirmed',
      'double_spend_suspected',
      'failed_expired'
    )
  );
