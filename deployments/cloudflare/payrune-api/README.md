## Payrune API Worker

This directory is the Cloudflare deployment shell for the standalone Payrune API Worker.

The actual API behavior lives in Go:

- `cmd/payrune-api-worker/`
- `internal/adapters/inbound/cloudflareworker/`
- `internal/adapters/outbound/persistence/cloudflarepostgres/`
- existing Go use cases under `internal/application/use_cases/`

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
- `BITCOIN_MAINNET_REQUIRED_CONFIRMATIONS`
- `BITCOIN_TESTNET4_REQUIRED_CONFIRMATIONS`
- `BITCOIN_MAINNET_RECEIPT_EXPIRES_AFTER`
- `BITCOIN_TESTNET4_RECEIPT_EXPIRES_AFTER`

### Deploy

```bash
make cf-api-deploy
```

The deploy flow will:

- explicitly tell you whether PostgreSQL migration will run or be skipped
- optionally prompt for `POSTGRES_CONNECTION_STRING`
- optionally prompt for xpub / confirmations / expiry values
- run migrations when a PostgreSQL connection string is provided
- build the Go/Wasm worker binary
- sync provided secrets to Wrangler
- deploy the Worker

### Teardown

```bash
make cf-api-delete
```
