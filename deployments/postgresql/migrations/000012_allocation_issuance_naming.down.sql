ALTER TABLE address_policy_cursors
  RENAME COLUMN address_space_ref TO address_source_ref;

DROP INDEX IF EXISTS idx_address_policy_allocations_policy_space_ref_reserved_at;
DROP INDEX IF EXISTS idx_address_policy_allocations_policy_space_ref_slot_index;

ALTER TABLE address_policy_allocations
  DROP CONSTRAINT IF EXISTS chk_address_policy_allocations_issued_issuance_ref;

ALTER TABLE address_policy_allocations
  DROP CONSTRAINT IF EXISTS chk_address_policy_allocations_issuance_ref_kind;

UPDATE address_policy_allocations
   SET issuance_ref = CASE
         WHEN issuance_ref_kind = 'create2_salt'
              AND COALESCE(issuance_ref, '') <> ''
           THEN address_policy_id || '/' || issuance_ref
         ELSE issuance_ref
       END
 WHERE allocation_status = 'issued';

ALTER TABLE address_policy_allocations
  DROP COLUMN IF EXISTS issuance_ref_kind;

ALTER TABLE address_policy_allocations
  RENAME COLUMN issuance_ref TO address_reference;

ALTER TABLE address_policy_allocations
  RENAME COLUMN slot_index TO derivation_index;

ALTER TABLE address_policy_allocations
  RENAME COLUMN address_space_ref TO address_source_ref;

CREATE UNIQUE INDEX IF NOT EXISTS idx_address_policy_allocations_policy_source_ref_index
  ON address_policy_allocations (address_policy_id, address_source_ref, derivation_index);

CREATE INDEX IF NOT EXISTS idx_address_policy_allocations_policy_source_ref_reserved_at
  ON address_policy_allocations (address_policy_id, address_source_ref, reserved_at DESC);
