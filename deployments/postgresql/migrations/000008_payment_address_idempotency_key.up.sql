CREATE TABLE IF NOT EXISTS payment_address_idempotency_keys (
  chain TEXT NOT NULL,
  idempotency_key TEXT NOT NULL,
  address_policy_id TEXT NOT NULL,
  expected_amount_minor BIGINT NOT NULL CHECK (expected_amount_minor > 0),
  customer_reference TEXT,
  payment_address_id BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT pk_payment_address_idempotency_keys PRIMARY KEY (chain, idempotency_key),
  CONSTRAINT fk_payment_address_idempotency_keys_payment_address
    FOREIGN KEY (payment_address_id)
    REFERENCES address_policy_allocations (id)
);
