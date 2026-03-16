CREATE TABLE IF NOT EXISTS evm_factories (
  id BIGSERIAL PRIMARY KEY,
  network TEXT NOT NULL,
  factory_address TEXT NOT NULL UNIQUE,
  collector_address TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('active', 'retired')),
  deployment_tx_hash TEXT,
  deployed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_evm_factories_active_network
  ON evm_factories (network)
  WHERE status = 'active';
