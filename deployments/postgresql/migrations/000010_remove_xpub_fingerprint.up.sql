TRUNCATE TABLE address_policy_allocations RESTART IDENTITY CASCADE;

ALTER TABLE address_policy_allocations
  ADD COLUMN IF NOT EXISTS account_public_key TEXT;

ALTER TABLE address_policy_allocations
  DROP COLUMN IF EXISTS xpub_fingerprint_algo,
  DROP COLUMN IF EXISTS xpub_fingerprint;

ALTER TABLE address_policy_allocations
  ALTER COLUMN account_public_key SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_address_policy_allocations_policy_xpub_index
  ON address_policy_allocations (address_policy_id, account_public_key, derivation_index);

CREATE INDEX IF NOT EXISTS idx_address_policy_allocations_policy_xpub_reserved_at
  ON address_policy_allocations (address_policy_id, account_public_key, reserved_at DESC);

DROP TABLE IF EXISTS address_policy_cursors;

CREATE TABLE IF NOT EXISTS address_policy_cursors (
  address_policy_id TEXT NOT NULL,
  account_public_key TEXT NOT NULL,
  next_index BIGINT NOT NULL CHECK (next_index >= 0 AND next_index <= 2147483648),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (address_policy_id, account_public_key)
);
