# Payrune Receipt Webhook Dispatcher Worker

This Worker runs the existing Go webhook dispatch use case on Cloudflare Workers via Go/Wasm.

## Deploy

- `make cf-up`

The deploy script can:

- sync the required Worker secrets `POSTGRES_CONNECTION_STRING`,
  `PAYMENT_RECEIPT_WEBHOOK_SECRET`

`PAYMENT_RECEIPT_WEBHOOK_INSECURE_SKIP_VERIFY` is intentionally not supported in the Cloudflare
worker runtime.

Repo root `.env.cloudflare` is auto-loaded before deploy and migrate flows.
Shell env still wins over values from `.env.cloudflare`.

Non-secret defaults live in `wrangler.toml`:

- `RECEIPT_WEBHOOK_DISPATCH_BATCH_SIZE = "50"`
- `RECEIPT_WEBHOOK_DISPATCH_CLAIM_TTL = "30s"`
- `RECEIPT_WEBHOOK_DISPATCH_MAX_ATTEMPTS = "10"`
- `RECEIPT_WEBHOOK_DISPATCH_RETRY_DELAY = "1m"`
- `PAYMENT_RECEIPT_WEBHOOK_TIMEOUT = "10s"`
- `crons = ["* * * * *"]`

Dispatcher always targets the internal Cloudflare mock binding mode:

- binding: `RECEIPT_WEBHOOK_MOCK`
- path: `/receipt-status`

Cloudflare observability logs are enabled by default in `wrangler.toml`.

## Migration

```bash
make cf-migrate
```

Run this separately before deploy when the target PostgreSQL schema needs to be updated.

## Delete

- `make cf-down`

## Runtime shape

- `deployments/cloudflare/payrune-webhook-dispatcher/` is only the Worker shell
- Go/Wasm runtime lives under `cmd/webhook-dispatcher-worker/`
- Cloudflare runtime wiring lives under `internal/infrastructure/di/`
- future webhook dispatch feature work should usually stay in Go code, not in this deployment shell

## Worker-to-worker calls

This worker uses a Cloudflare service binding for `receipt-webhook-mock`. It does not need to call
`payrune-api` or `payrune-poller`.
