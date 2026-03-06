CREATE TABLE IF NOT EXISTS payment_receipt_status_notifications (
  id BIGSERIAL PRIMARY KEY,
  payment_address_id BIGINT NOT NULL REFERENCES address_policy_allocations(id) ON DELETE CASCADE,
  customer_reference TEXT,
  previous_status TEXT NOT NULL CHECK (
    previous_status IN (
      'watching',
      'partially_paid',
      'paid_unconfirmed',
      'paid_confirmed',
      'double_spend_suspected',
      'failed_expired'
    )
  ),
  current_status TEXT NOT NULL CHECK (
    current_status IN (
      'watching',
      'partially_paid',
      'paid_unconfirmed',
      'paid_confirmed',
      'double_spend_suspected',
      'failed_expired'
    )
  ),
  observed_total_minor BIGINT NOT NULL CHECK (observed_total_minor >= 0),
  confirmed_total_minor BIGINT NOT NULL CHECK (confirmed_total_minor >= 0),
  unconfirmed_total_minor BIGINT NOT NULL CHECK (unconfirmed_total_minor >= 0),
  conflict_total_minor BIGINT NOT NULL CHECK (conflict_total_minor >= 0),
  status_changed_at TIMESTAMPTZ NOT NULL,
  delivery_status TEXT NOT NULL CHECK (delivery_status IN ('pending', 'sent', 'failed')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_receipt_status_notifications_delivery_created
  ON payment_receipt_status_notifications (delivery_status, created_at ASC);

CREATE INDEX IF NOT EXISTS idx_payment_receipt_status_notifications_address_created
  ON payment_receipt_status_notifications (payment_address_id, created_at DESC);
