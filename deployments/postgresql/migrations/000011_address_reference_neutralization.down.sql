ALTER TABLE address_policy_cursors
  RENAME COLUMN address_source_ref TO account_public_key;

DROP INDEX IF EXISTS idx_address_policy_allocations_policy_source_ref_reserved_at;
DROP INDEX IF EXISTS idx_address_policy_allocations_policy_source_ref_index;

ALTER TABLE address_policy_allocations
  RENAME COLUMN address_reference TO derivation_path;

ALTER TABLE address_policy_allocations
  RENAME COLUMN address_source_ref TO account_public_key;

CREATE UNIQUE INDEX IF NOT EXISTS idx_address_policy_allocations_policy_xpub_index
  ON address_policy_allocations (address_policy_id, account_public_key, derivation_index);

CREATE INDEX IF NOT EXISTS idx_address_policy_allocations_policy_xpub_reserved_at
  ON address_policy_allocations (address_policy_id, account_public_key, reserved_at DESC);
