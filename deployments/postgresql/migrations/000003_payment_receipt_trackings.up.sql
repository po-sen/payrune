CREATE TABLE IF NOT EXISTS payment_receipt_trackings (
  id BIGSERIAL PRIMARY KEY,
  payment_address_id BIGINT NOT NULL UNIQUE REFERENCES address_policy_allocations(id) ON DELETE CASCADE,
  address_policy_id TEXT NOT NULL,
  chain TEXT NOT NULL,
  network TEXT NOT NULL,
  address TEXT NOT NULL,
  issued_at TIMESTAMPTZ,
  expected_amount_minor BIGINT NOT NULL CHECK (expected_amount_minor > 0),
  required_confirmations INTEGER NOT NULL CHECK (required_confirmations >= 1),
  receipt_status TEXT NOT NULL CHECK (
    receipt_status IN (
      'watching',
      'partially_paid',
      'paid_unconfirmed',
      'paid_confirmed',
      'double_spend_suspected'
    )
  ),
  observed_total_minor BIGINT NOT NULL DEFAULT 0 CHECK (observed_total_minor >= 0),
  confirmed_total_minor BIGINT NOT NULL DEFAULT 0 CHECK (confirmed_total_minor >= 0),
  unconfirmed_total_minor BIGINT NOT NULL DEFAULT 0 CHECK (unconfirmed_total_minor >= 0),
  conflict_total_minor BIGINT NOT NULL DEFAULT 0 CHECK (conflict_total_minor >= 0),
  last_observed_block_height BIGINT NOT NULL DEFAULT 0 CHECK (last_observed_block_height >= 0),
  first_observed_at TIMESTAMPTZ,
  paid_at TIMESTAMPTZ,
  confirmed_at TIMESTAMPTZ,
  last_polled_at TIMESTAMPTZ,
  next_poll_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  last_error TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (chain, address)
);

CREATE INDEX IF NOT EXISTS idx_payment_receipt_trackings_due
  ON payment_receipt_trackings (receipt_status, next_poll_at ASC);

CREATE INDEX IF NOT EXISTS idx_payment_receipt_trackings_address_policy
  ON payment_receipt_trackings (address_policy_id, created_at DESC);
