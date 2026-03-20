ALTER TABLE address_policy_allocations
  RENAME COLUMN account_public_key TO address_source_ref;

ALTER TABLE address_policy_allocations
  RENAME COLUMN derivation_path TO address_reference;

DROP INDEX IF EXISTS idx_address_policy_allocations_policy_xpub_index;
CREATE UNIQUE INDEX IF NOT EXISTS idx_address_policy_allocations_policy_source_ref_index
  ON address_policy_allocations (address_policy_id, address_source_ref, derivation_index);

DROP INDEX IF EXISTS idx_address_policy_allocations_policy_xpub_reserved_at;
CREATE INDEX IF NOT EXISTS idx_address_policy_allocations_policy_source_ref_reserved_at
  ON address_policy_allocations (address_policy_id, address_source_ref, reserved_at DESC);

ALTER TABLE address_policy_cursors
  RENAME COLUMN account_public_key TO address_source_ref;
