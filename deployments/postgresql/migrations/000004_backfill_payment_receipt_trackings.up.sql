INSERT INTO payment_receipt_trackings (
  payment_address_id,
  address_policy_id,
  chain,
  network,
  address,
  issued_at,
  expected_amount_minor,
  required_confirmations,
  receipt_status,
  next_poll_at
)
SELECT
  a.id,
  a.address_policy_id,
  a.chain,
  a.network,
  a.address,
  a.issued_at,
  a.expected_amount_minor,
  1,
  'watching',
  NOW()
FROM address_policy_allocations a
WHERE a.allocation_status = 'issued'
  AND a.network IS NOT NULL
  AND a.address IS NOT NULL
ON CONFLICT (payment_address_id) DO NOTHING;
