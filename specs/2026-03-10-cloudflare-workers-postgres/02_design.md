---
doc: 02_design
spec_date: 2026-03-10
slug: cloudflare-workers-postgres
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-08-payment-address-idempotency-key
  - 2026-03-08-payment-address-status-api
  - 2026-03-09-receipt-expire-final-check
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Design

## Summary

- Cloudflare Worker directly serves the public Payrune API.
- Go/Wasm is the application runtime inside the Worker.
- PostgreSQL stays external and is accessed through a JS bridge plus a Go Cloudflare-specific outbound adapter.
- `deployments/cloudflare/payrune-api/` stays a deployment shell.

## Code placement

### Deployment shell

`deployments/cloudflare/payrune-api/` owns:

- `wrangler.toml`
- thin `src/index.mjs`
- `src/go-wasm-runtime.mjs`
- `src/postgres-bridge.mjs`
- package metadata
- deployment-focused README
- Worker-focused test runner wiring
- generated Wasm bootstrap artifacts

The deployment shell imports the generated `.wasm` artifact directly. It must not wrap the Wasm
payload into base64 JavaScript because that inflates the bundle beyond Cloudflare's Worker size
limit.

### Go application runtime

`cmd/payrune-api-worker/` owns:

- Go-Wasm entrypoint
- Worker-specific composition root
- env-to-Go wiring for Cloudflare runtime

`internal/adapters/inbound/cloudflareworker/` owns:

- Worker request/response envelope mapping
- translation from Worker requests into Go HTTP handler invocation

`internal/adapters/outbound/persistence/cloudflarepostgres/` owns:

- Worker-compatible PostgreSQL adapter
- transaction bridge
- API-facing persistence stores and finder implementations needed by Go use cases

This keeps future API work in Go and out of `deployments/`.

## Runtime model

### Request flow

1. Worker `fetch()` receives the request.
2. JS shell snapshots string env values and registers a request-scoped PostgreSQL bridge context.
3. JS shell forwards a request envelope into Go/Wasm.
4. Go/Wasm builds the Worker HTTP handler and dispatches through existing Go controllers and use cases.
5. When a use case needs persistence, the Cloudflare Postgres adapter calls the JS PostgreSQL bridge.
6. JS bridge executes SQL via `pg`, returns rows/results/errors, and Go continues the use case flow.
7. Go/Wasm returns a response envelope to JS, and JS creates the final Worker `Response`.

### PostgreSQL access

- Worker uses `POSTGRES_CONNECTION_STRING`.
- PostgreSQL access uses `pg` with Worker Node.js compatibility enabled.
- The JS bridge maintains pooled connections by connection string and request-scoped transaction contexts.
- The Go adapter owns the outbound port implementations used by the existing Go use cases.

## API behavior

The Worker runs the existing Go controllers and use cases for:

- `GET /health`
- `GET /v1/chains/bitcoin/address-policies`
- `GET /v1/chains/bitcoin/addresses`
- `POST /v1/chains/bitcoin/payment-addresses`
- `GET /v1/chains/bitcoin/payment-addresses/{paymentAddressId}`

The JS shell is route-agnostic for `/v1/...`; future `/v1/...` additions should not require deployment-shell changes.

## Failure modes

- Missing `POSTGRES_CONNECTION_STRING` for DB-backed routes:
  - Worker returns `500 {"error":"internal server error"}`
- Missing xpub for a disabled policy:
  - address generation/create return `501` with the current contract error
- PostgreSQL connection/query failure:
  - Worker returns `500 {"error":"internal server error"}`
- Unexpected transaction inconsistency:
  - Worker returns `500 {"error":"internal server error"}`
- Go-Wasm bootstrap failure:
  - Worker returns `500 {"error":"internal server error"}`

## Observability

- Worker JS shell logs unexpected bootstrap and PostgreSQL bridge failures with `console.error`.
- Go code continues to return current controller/use-case error mappings.

## Security

- This slice only covers functional standalone deployment.
- Edge auth, rate limiting, and origin hardening remain out of scope.

## Removed designs

This slice explicitly excludes:

- thin-edge proxy/origin forwarding
- standalone JS business-logic reimplementation
