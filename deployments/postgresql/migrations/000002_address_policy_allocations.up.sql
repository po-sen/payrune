CREATE TABLE IF NOT EXISTS address_policy_cursors (
  address_policy_id TEXT NOT NULL,
  xpub_fingerprint_algo TEXT NOT NULL,
  xpub_fingerprint TEXT NOT NULL,
  next_index BIGINT NOT NULL CHECK (next_index >= 0 AND next_index <= 2147483648),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (address_policy_id, xpub_fingerprint_algo, xpub_fingerprint)
);

CREATE TABLE IF NOT EXISTS address_policy_allocations (
  id BIGSERIAL PRIMARY KEY,
  address_policy_id TEXT NOT NULL,
  xpub_fingerprint_algo TEXT NOT NULL,
  xpub_fingerprint TEXT NOT NULL,
  derivation_index BIGINT NOT NULL CHECK (derivation_index >= 0 AND derivation_index <= 2147483647),
  expected_amount_minor BIGINT NOT NULL CHECK (expected_amount_minor > 0),
  customer_reference TEXT,
  chain TEXT,
  network TEXT,
  scheme TEXT,
  address TEXT,
  derivation_path TEXT,
  allocation_status TEXT NOT NULL CHECK (allocation_status IN ('reserved', 'issued', 'derivation_failed')),
  failure_reason TEXT,
  reserved_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  issued_at TIMESTAMPTZ,
  UNIQUE (address_policy_id, xpub_fingerprint_algo, xpub_fingerprint, derivation_index),
  UNIQUE (chain, address)
);

CREATE INDEX IF NOT EXISTS idx_address_policy_allocations_chain_address
  ON address_policy_allocations (chain, address);

CREATE INDEX IF NOT EXISTS idx_address_policy_allocations_policy_fp_reserved_at
  ON address_policy_allocations (address_policy_id, xpub_fingerprint_algo, xpub_fingerprint, reserved_at DESC);

CREATE INDEX IF NOT EXISTS idx_address_policy_allocations_customer_reference
  ON address_policy_allocations (customer_reference, reserved_at DESC)
  WHERE customer_reference IS NOT NULL;
