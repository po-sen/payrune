ALTER TABLE payment_receipt_trackings
  DROP CONSTRAINT chk_payment_receipt_trackings_asset_reference;

ALTER TABLE address_policy_allocations
  DROP CONSTRAINT chk_address_policy_allocations_asset_reference;

ALTER TABLE payment_receipt_trackings
  DROP COLUMN asset_reference;

ALTER TABLE address_policy_allocations
  DROP COLUMN asset_reference;
