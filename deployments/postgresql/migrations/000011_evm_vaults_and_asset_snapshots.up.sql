ALTER TABLE evm_factories
  ADD COLUMN IF NOT EXISTS vault_creation_code_hash TEXT;

ALTER TABLE address_policy_allocations
  ADD COLUMN IF NOT EXISTS asset_code TEXT,
  ADD COLUMN IF NOT EXISTS asset_type TEXT,
  ADD COLUMN IF NOT EXISTS token_address TEXT,
  ADD COLUMN IF NOT EXISTS minor_unit TEXT,
  ADD COLUMN IF NOT EXISTS decimals INTEGER,
  ADD COLUMN IF NOT EXISTS issuance_method TEXT;

UPDATE address_policy_allocations
SET asset_code = COALESCE(asset_code, 'btc'),
    asset_type = COALESCE(asset_type, 'native'),
    minor_unit = COALESCE(minor_unit, 'satoshi'),
    decimals = COALESCE(decimals, 8),
    issuance_method = COALESCE(issuance_method, 'xpub_derivation')
WHERE asset_code IS NULL
   OR asset_type IS NULL
   OR minor_unit IS NULL
   OR decimals IS NULL
   OR issuance_method IS NULL;

ALTER TABLE address_policy_allocations
  ALTER COLUMN asset_code SET NOT NULL,
  ALTER COLUMN asset_type SET NOT NULL,
  ALTER COLUMN minor_unit SET NOT NULL,
  ALTER COLUMN decimals SET NOT NULL,
  ALTER COLUMN issuance_method SET NOT NULL;

ALTER TABLE payment_receipt_trackings
  ADD COLUMN IF NOT EXISTS asset_code TEXT,
  ADD COLUMN IF NOT EXISTS asset_type TEXT,
  ADD COLUMN IF NOT EXISTS token_address TEXT,
  ADD COLUMN IF NOT EXISTS minor_unit TEXT,
  ADD COLUMN IF NOT EXISTS decimals INTEGER,
  ADD COLUMN IF NOT EXISTS issuance_method TEXT;

UPDATE payment_receipt_trackings
SET asset_code = COALESCE(asset_code, 'btc'),
    asset_type = COALESCE(asset_type, 'native'),
    minor_unit = COALESCE(minor_unit, 'satoshi'),
    decimals = COALESCE(decimals, 8),
    issuance_method = COALESCE(issuance_method, 'xpub_derivation')
WHERE asset_code IS NULL
   OR asset_type IS NULL
   OR minor_unit IS NULL
   OR decimals IS NULL
   OR issuance_method IS NULL;

ALTER TABLE payment_receipt_trackings
  ALTER COLUMN asset_code SET NOT NULL,
  ALTER COLUMN asset_type SET NOT NULL,
  ALTER COLUMN minor_unit SET NOT NULL,
  ALTER COLUMN decimals SET NOT NULL,
  ALTER COLUMN issuance_method SET NOT NULL;

CREATE TABLE IF NOT EXISTS evm_payment_vaults (
  payment_address_id BIGINT PRIMARY KEY REFERENCES address_policy_allocations(id) ON DELETE CASCADE,
  network TEXT NOT NULL,
  factory_id BIGINT NOT NULL REFERENCES evm_factories(id),
  factory_address TEXT NOT NULL,
  collector_address TEXT NOT NULL,
  token_address TEXT,
  salt_hex TEXT NOT NULL UNIQUE,
  predicted_address TEXT NOT NULL UNIQUE,
  deploy_status TEXT NOT NULL CHECK (deploy_status IN ('predicted', 'deployed')),
  sweep_status TEXT NOT NULL CHECK (sweep_status IN ('pending', 'submitted', 'succeeded', 'failed')),
  deploy_tx_hash TEXT,
  last_sweep_tx_hash TEXT,
  last_sweep_error TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
