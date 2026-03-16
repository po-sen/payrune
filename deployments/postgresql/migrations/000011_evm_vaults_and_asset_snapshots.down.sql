DROP TABLE IF EXISTS evm_payment_vaults;

ALTER TABLE payment_receipt_trackings
  DROP COLUMN IF EXISTS issuance_method,
  DROP COLUMN IF EXISTS decimals,
  DROP COLUMN IF EXISTS minor_unit,
  DROP COLUMN IF EXISTS token_address,
  DROP COLUMN IF EXISTS asset_type,
  DROP COLUMN IF EXISTS asset_code;

ALTER TABLE address_policy_allocations
  DROP COLUMN IF EXISTS issuance_method,
  DROP COLUMN IF EXISTS decimals,
  DROP COLUMN IF EXISTS minor_unit,
  DROP COLUMN IF EXISTS token_address,
  DROP COLUMN IF EXISTS asset_type,
  DROP COLUMN IF EXISTS asset_code;

ALTER TABLE evm_factories
  DROP COLUMN IF EXISTS vault_creation_code_hash;
