# Receipt Webhook Mock Worker

This is a Cloudflare-only mock webhook target for testing `payrune-webhook-dispatcher`.

## Behavior

- `POST /receipt-status`
  - verifies `X-Payrune-Signature-256` with `PAYMENT_RECEIPT_WEBHOOK_SECRET` when configured
  - logs the received request
  - returns `204` on success
  - returns `401` on invalid signature
- `GET /health`
  - returns `200 ok`

## Deploy

- `make cf-up`

The default Cloudflare stack deploys this worker before `payrune-webhook-dispatcher`.

## Delete

- `make cf-down`

## Notes

- This worker is intentionally JS-only because it is a test helper, not a core business runtime.
- Dispatcher targets this worker through a Cloudflare service binding by default.
