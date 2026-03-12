## Payrune API Worker

This directory is the Cloudflare deployment shell for the standalone Payrune API Worker.

The actual API behavior lives in Go:

- `cmd/api-worker/`
- `internal/adapters/inbound/cloudflareworker/`
- `internal/infrastructure/di/`
- `internal/adapters/outbound/persistence/cloudflarepostgres/`
- existing Go use cases under `internal/application/usecases/`

`deployments/cloudflare/payrune-api/` only owns:

- Wrangler config
- Worker shell / Go-Wasm loader
- PostgreSQL JS bridge
- deploy/test wiring

The deployment shell imports the generated `.wasm` file directly instead of base64-wrapping it into
JavaScript so the Worker bundle stays under Cloudflare's free-plan size limit.

Future `/v1/...` API work should usually happen in Go, not in this directory.

### Required Worker secret

- `POSTGRES_CONNECTION_STRING`

### Optional Worker secrets

- `BITCOIN_MAINNET_LEGACY_XPUB`
- `BITCOIN_MAINNET_SEGWIT_XPUB`
- `BITCOIN_MAINNET_NATIVE_SEGWIT_XPUB`
- `BITCOIN_MAINNET_TAPROOT_XPUB`
- `BITCOIN_TESTNET4_LEGACY_XPUB`
- `BITCOIN_TESTNET4_SEGWIT_XPUB`
- `BITCOIN_TESTNET4_NATIVE_SEGWIT_XPUB`
- `BITCOIN_TESTNET4_TAPROOT_XPUB`

### Non-secret Worker defaults

These now live in `wrangler.toml`:

- `BITCOIN_MAINNET_REQUIRED_CONFIRMATIONS = "2"`
- `BITCOIN_TESTNET4_REQUIRED_CONFIRMATIONS = "2"`
- `BITCOIN_MAINNET_RECEIPT_EXPIRES_AFTER = "24h"`
- `BITCOIN_TESTNET4_RECEIPT_EXPIRES_AFTER = "24h"`

### Deploy

```bash
make cf-up
```

Repo root `.env.cloudflare` is auto-loaded before deploy and migrate flows.
Shell env still wins over values from `.env.cloudflare`.

The deploy flow will:

- explicitly tell you whether `POSTGRES_CONNECTION_STRING` Worker secret sync will run or be skipped
- build the Go/Wasm worker binary
- sync provided secrets to Wrangler
- deploy the Worker

### Migration

```bash
make cf-migrate
```

Run this separately before deploy when the target PostgreSQL schema needs to be updated.

`make cf-migrate` also auto-loads repo root `.env.cloudflare`.

### Teardown

```bash
make cf-down
```
