# Payrune Worker

This directory is the unified Cloudflare deployment shell for the Payrune service worker.

The actual API, poller, and dispatcher behavior lives in Go:

- `cmd/payrune-worker/`
- `internal/bootstrap/`
- `internal/infrastructure/di/`
- existing Go use cases under `internal/application/usecases/`

`deployments/cloudflare/payrune/` only owns:

- Wrangler config
- unified Worker shell / Go-Wasm loader
- PostgreSQL JS bridge
- Bitcoin observer bridge
- webhook notifier bridge
- deploy/test wiring

Cloudflare observability logs are enabled by default in `wrangler.toml`.

## Public API

- `GET /health`
- `/v1/...`

All other `fetch()` paths return `404`.

## Scheduled jobs

- Ethereum mainnet poller: `2,17,32,47 * * * *`
- Bitcoin mainnet poller: `5,20,35,50 * * * *`
- Ethereum sepolia poller: `8,23,38,53 * * * *`
- Bitcoin testnet4 poller: `*/15 * * * *`
- Receipt webhook dispatcher: `10,25,40,55 * * * *`

## Required Worker secrets

- `POSTGRES_CONNECTION_STRING`
- `PAYMENT_RECEIPT_WEBHOOK_SECRET`

## Optional env-backed Worker values

These are the optional values currently supported by `make cf-up`. When non-empty, the deploy
script syncs them as Wrangler secrets.

- `BITCOIN_MAINNET_LEGACY_XPUB`
- `BITCOIN_MAINNET_SEGWIT_XPUB`
- `BITCOIN_MAINNET_NATIVE_SEGWIT_XPUB`
- `BITCOIN_MAINNET_TAPROOT_XPUB`
- `BITCOIN_TESTNET4_LEGACY_XPUB`
- `BITCOIN_TESTNET4_SEGWIT_XPUB`
- `BITCOIN_TESTNET4_NATIVE_SEGWIT_XPUB`
- `BITCOIN_TESTNET4_TAPROOT_XPUB`
- `ETHEREUM_MAINNET_CREATE2_COLLECTOR_ADDRESS`
- `ETHEREUM_MAINNET_CREATE2_DERIVATION_KEY`
- `ETHEREUM_MAINNET_USDT_ASSET_REFERENCE`
- `ETHEREUM_SEPOLIA_CREATE2_COLLECTOR_ADDRESS`
- `ETHEREUM_SEPOLIA_CREATE2_DERIVATION_KEY`
- `ETHEREUM_SEPOLIA_USDT_ASSET_REFERENCE`
- `BITCOIN_MAINNET_ESPLORA_USER`
- `BITCOIN_MAINNET_ESPLORA_PASSWORD`
- `BITCOIN_TESTNET4_ESPLORA_USER`
- `BITCOIN_TESTNET4_ESPLORA_PASSWORD`
- `ETHEREUM_MAINNET_RPC_USER`
- `ETHEREUM_MAINNET_RPC_PASSWORD`
- `ETHEREUM_SEPOLIA_RPC_USER`
- `ETHEREUM_SEPOLIA_RPC_PASSWORD`

CREATE2 derivation keys must be 32-byte hex strings such as
`0x1111111111111111111111111111111111111111111111111111111111111111`.

## Non-secret Worker defaults

These live in `wrangler.toml`:

- API confirmation / receipt-expiry defaults
- Ethereum CREATE2 confirmation / receipt-expiry defaults
- poller cadence and batch defaults
- network-specific Esplora endpoint defaults
- network-specific Ethereum JSON-RPC endpoint defaults
- webhook dispatcher timeout and retry defaults

Current public RPC defaults:

- Ethereum mainnet: `https://ethereum-rpc.publicnode.com`
- Ethereum sepolia: `https://ethereum-sepolia-rpc.publicnode.com`

These URL values come from checked-in Worker config in `wrangler.toml`, not from Wrangler secrets.
RPC usernames and passwords remain Wrangler secrets.

Factory addresses and receiver init-code hashes are not Worker secrets. They are expected to come
from checked-in deployment metadata and contract artifacts once the CREATE2 contracts land.

## Deploy

```bash
make cf-up
```

Repo root `.env.cloudflare` is auto-loaded before deploy and migrate flows.
Shell env still wins over values from `.env.cloudflare`.

The default Cloudflare stack deploys `receipt-webhook-mock` first, then deploys this unified
worker.

## Migration

```bash
make cf-migrate
```

Run this separately before deploy when the target PostgreSQL schema needs to be updated.

## Delete

- `make cf-down`

## Worker-to-worker calls

This worker uses a Cloudflare service binding for `receipt-webhook-mock` during webhook dispatcher
execution.
