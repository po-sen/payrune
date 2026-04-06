ALTER TABLE address_policy_allocations
  ADD COLUMN asset_reference TEXT;

ALTER TABLE payment_receipt_trackings
  ADD COLUMN asset_reference TEXT;

UPDATE payment_receipt_trackings pr
SET asset_reference = a.asset_reference
FROM address_policy_allocations a
WHERE a.id = pr.payment_address_id
  AND a.asset_reference IS NOT NULL;

ALTER TABLE address_policy_allocations
  ADD CONSTRAINT chk_address_policy_allocations_asset_reference
  CHECK (
    asset_reference IS NULL
    OR BTRIM(asset_reference) <> ''
  );

ALTER TABLE payment_receipt_trackings
  ADD CONSTRAINT chk_payment_receipt_trackings_asset_reference
  CHECK (
    asset_reference IS NULL
    OR BTRIM(asset_reference) <> ''
  );
