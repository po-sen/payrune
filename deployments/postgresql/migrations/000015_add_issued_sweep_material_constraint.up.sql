ALTER TABLE address_policy_allocations
  DROP CONSTRAINT IF EXISTS chk_address_policy_allocations_issued_sweep_material;

ALTER TABLE address_policy_allocations
  ADD CONSTRAINT chk_address_policy_allocations_issued_sweep_material
  CHECK (
    allocation_status <> 'issued'
    OR sweep_material_json IS NOT NULL
  );
