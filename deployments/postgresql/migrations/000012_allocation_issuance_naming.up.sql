ALTER TABLE address_policy_allocations
  RENAME COLUMN address_source_ref TO address_space_ref;

ALTER TABLE address_policy_allocations
  RENAME COLUMN derivation_index TO slot_index;

ALTER TABLE address_policy_allocations
  RENAME COLUMN address_reference TO issuance_ref;

ALTER TABLE address_policy_allocations
  ADD COLUMN issuance_ref_kind TEXT;

UPDATE address_policy_allocations
   SET issuance_ref = CASE
         WHEN scheme = 'create2'
              AND POSITION('/' IN COALESCE(issuance_ref, '')) > 0
           THEN split_part(issuance_ref, '/', 2)
         ELSE issuance_ref
       END,
       issuance_ref_kind = CASE
         WHEN scheme = 'create2'
              AND COALESCE(issuance_ref, '') <> ''
           THEN 'create2_salt'
         WHEN chain = 'bitcoin'
              AND COALESCE(issuance_ref, '') <> ''
           THEN 'hd_path_absolute'
         ELSE issuance_ref_kind
       END
 WHERE allocation_status = 'issued';

ALTER TABLE address_policy_allocations
  ADD CONSTRAINT chk_address_policy_allocations_issuance_ref_kind
  CHECK (issuance_ref_kind IS NULL OR issuance_ref_kind IN ('hd_path_absolute', 'create2_salt'));

ALTER TABLE address_policy_allocations
  ADD CONSTRAINT chk_address_policy_allocations_issued_issuance_ref
  CHECK (
    allocation_status <> 'issued'
    OR (
      issuance_ref_kind IS NOT NULL
      AND issuance_ref IS NOT NULL
    )
  );

DROP INDEX IF EXISTS idx_address_policy_allocations_policy_source_ref_index;
CREATE UNIQUE INDEX IF NOT EXISTS idx_address_policy_allocations_policy_space_ref_slot_index
  ON address_policy_allocations (address_policy_id, address_space_ref, slot_index);

DROP INDEX IF EXISTS idx_address_policy_allocations_policy_source_ref_reserved_at;
CREATE INDEX IF NOT EXISTS idx_address_policy_allocations_policy_space_ref_reserved_at
  ON address_policy_allocations (address_policy_id, address_space_ref, reserved_at DESC);

ALTER TABLE address_policy_cursors
  RENAME COLUMN address_source_ref TO address_space_ref;
