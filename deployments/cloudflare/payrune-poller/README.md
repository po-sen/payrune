# Payrune Poller Worker

This Worker runs the existing Go receipt polling use case on Cloudflare Workers via Go/Wasm.

## Deploy

- `make cf-up`

The deploy script can:

- optionally sync the Worker secret `POSTGRES_CONNECTION_STRING`
- sync optional target-network Esplora auth secrets

Repo root `.env.cloudflare` is auto-loaded before deploy and migrate flows.
Shell env still wins over values from `.env.cloudflare`.

Non-secret defaults now live in `wrangler.toml`:

- `POLL_RESCHEDULE_INTERVAL = "10m"`
- `POLL_BATCH_SIZE = "10"`
- `POLL_CLAIM_TTL = "2m"`
- `env.testnet4.triggers.crons = ["*/15 * * * *"]`
- `env.mainnet.triggers.crons = ["5,20,35,50 * * * *"]`
- `BITCOIN_MAINNET_ESPLORA_URL = "https://mempool.space/api"`
- `BITCOIN_MAINNET_ESPLORA_TIMEOUT = "10s"`
- `BITCOIN_TESTNET4_ESPLORA_URL = "https://mempool.space/testnet4/api"`
- `BITCOIN_TESTNET4_ESPLORA_TIMEOUT = "10s"`

Cloudflare observability logs are enabled by default in `wrangler.toml`.

## Migration

```bash
make cf-migrate
```

Run this separately before deploy when the target PostgreSQL schema needs to be updated.

## Delete

- `make cf-down`

## Runtime shape

- `deployments/cloudflare/payrune-poller/` is only the Worker shell
- Go/Wasm runtime lives under `cmd/poller-worker/`
- Cloudflare runtime wiring lives under `internal/infrastructure/di/`
- future poller feature work should usually stay in Go code, not in this deployment shell
