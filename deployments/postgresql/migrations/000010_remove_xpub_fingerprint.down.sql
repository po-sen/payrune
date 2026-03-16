DROP INDEX IF EXISTS idx_address_policy_allocations_policy_xpub_reserved_at;
DROP INDEX IF EXISTS idx_address_policy_allocations_policy_xpub_index;

ALTER TABLE address_policy_allocations
  ADD COLUMN IF NOT EXISTS xpub_fingerprint_algo TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS xpub_fingerprint TEXT NOT NULL DEFAULT '';

UPDATE address_policy_allocations
   SET xpub_fingerprint_algo = 'raw-account-public-key-v1',
       xpub_fingerprint = account_public_key
 WHERE account_public_key IS NOT NULL
   AND COALESCE(xpub_fingerprint_algo, '') = ''
   AND COALESCE(xpub_fingerprint, '') = '';

ALTER TABLE address_policy_allocations
  ALTER COLUMN xpub_fingerprint_algo DROP DEFAULT,
  ALTER COLUMN xpub_fingerprint DROP DEFAULT;

ALTER TABLE address_policy_allocations
  DROP COLUMN IF EXISTS account_public_key;

ALTER TABLE address_policy_allocations
  ADD CONSTRAINT uq_address_policy_allocations_policy_fingerprint_index
  UNIQUE (address_policy_id, xpub_fingerprint_algo, xpub_fingerprint, derivation_index);

CREATE INDEX IF NOT EXISTS idx_address_policy_allocations_policy_fp_reserved_at
  ON address_policy_allocations (address_policy_id, xpub_fingerprint_algo, xpub_fingerprint, reserved_at DESC);

DROP TABLE IF EXISTS address_policy_cursors;

CREATE TABLE IF NOT EXISTS address_policy_cursors (
  address_policy_id TEXT NOT NULL,
  xpub_fingerprint_algo TEXT NOT NULL,
  xpub_fingerprint TEXT NOT NULL,
  next_index BIGINT NOT NULL CHECK (next_index >= 0 AND next_index <= 2147483648),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (address_policy_id, xpub_fingerprint_algo, xpub_fingerprint)
);

INSERT INTO address_policy_cursors (
  address_policy_id,
  xpub_fingerprint_algo,
  xpub_fingerprint,
  next_index
)
SELECT
  address_policy_id,
  xpub_fingerprint_algo,
  xpub_fingerprint,
  MAX(derivation_index) + 1
FROM address_policy_allocations
GROUP BY address_policy_id, xpub_fingerprint_algo, xpub_fingerprint
ON CONFLICT (address_policy_id, xpub_fingerprint_algo, xpub_fingerprint) DO NOTHING;
