ALTER TABLE address_policy_allocations
  DROP CONSTRAINT IF EXISTS chk_address_policy_allocations_issued_issuance_ref;

ALTER TABLE address_policy_allocations
  DROP CONSTRAINT IF EXISTS chk_address_policy_allocations_issuance_ref_kind;

ALTER TABLE address_policy_allocations
  DROP COLUMN IF EXISTS issuance_ref_kind,
  DROP COLUMN IF EXISTS issuance_ref;
