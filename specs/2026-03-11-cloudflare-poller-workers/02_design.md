---
doc: 02_design
spec_date: 2026-03-11
slug: cloudflare-poller-workers
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-05-blockchain-receipt-polling-service
  - 2026-03-09-shared-tip-height-polling
  - 2026-03-09-poller-interval-separation
  - 2026-03-09-receipt-expire-final-check
  - 2026-03-10-cloudflare-workers-postgres
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Technical Design

## High-level approach

- Summary:
  - Add a standalone Cloudflare Worker poller runtime that executes the existing Go receipt polling use case through Go/Wasm.
  - Replace process-local ticker lifecycle with Cloudflare `scheduled()` handlers.
  - Reuse the Worker-compatible PostgreSQL adapter pattern and add a Worker-compatible Bitcoin Esplora observer bridge.
  - Keep deployment shell thin and network-specific deployment explicit.
- Key decisions:
  - Use one shared codebase and thin deployment shell for both Bitcoin pollers.
  - Deploy mainnet and testnet4 as separate Cloudflare worker names or Wrangler environments.
  - Keep chain/network scope explicit in the scheduled runtime; no inferred scope.
  - Preserve current `RunReceiptPollingCycleUseCase` orchestration and domain behavior.

## System context

- Components:
  - `cmd/poller-worker/`
    - Go/Wasm poller entrypoint only.
  - `internal/adapters/inbound/cloudflareworker/`
    - scheduled poller request/response mapping and use-case invocation from already-built poller use cases.
  - `internal/adapters/outbound/persistence/cloudflarepostgres/`
    - Worker-compatible PostgreSQL persistence path for claim/save operations.
  - `internal/adapters/outbound/bitcoin/`
    - new Cloudflare-compatible Esplora observer adapter or bridge-backed variant.
  - `internal/infrastructure/di/`
    - Cloudflare Worker runtime wiring, env parsing, and construction of concrete poller dependencies.
  - `deployments/cloudflare/payrune-poller/`
    - Wrangler config, Worker JS shell, Go-Wasm loader, JS bridge(s), deploy wiring.
- Interfaces:
  - Cloudflare `scheduled()` handler as the runtime trigger.
  - PostgreSQL bridge interface.
  - Bitcoin Esplora bridge interface.

## Key flows

- Flow 1: Scheduled mainnet poll cycle

1. Cloudflare triggers `scheduled()` for the mainnet poller Worker.
2. JS shell registers bridge context(s) and forwards a scheduled-event envelope into Go/Wasm.
3. Cloudflare DI wiring builds Worker poller dependencies with `chain=bitcoin`, `network=mainnet`.
4. Inbound poller handler maps the scheduled request into `RunReceiptPollingCycleUseCase.Execute(...)`.
5. Use case performs due-claim, shared tip-height fetch, observation, save, and outbox enqueue decisions using existing Go behavior.
6. Worker logs cycle counters and exits.

- Flow 2: Scheduled testnet4 poll cycle

  1. Same as Flow 1, but `network=testnet4`.

- Flow 3: PostgreSQL path

  1. Go polling use case calls Worker-compatible PostgreSQL stores through outbound ports.
  2. Go adapter calls the JS PostgreSQL bridge.
  3. JS bridge executes SQL via `pg` and returns rows/results/errors.

- Flow 4: Bitcoin observation path
  1. Go polling use case calls the Worker-compatible Bitcoin observer port.
  2. The Worker-compatible observer calls a JS fetch bridge or a Worker-compatible concrete adapter.
  3. JS bridge performs Esplora HTTP requests and returns normalized observation data.
  4. Go use case continues with existing lifecycle logic.

## Data model

- Entities:
  - Existing receipt tracking rows remain unchanged.
- Schema changes or migrations:
  - No schema change is required for the initial Worker poller migration.
  - Deploy flow still runs migrations defensively when a PostgreSQL connection string is provided.
- Consistency and idempotency:
  - Existing PostgreSQL claim/lease semantics remain the concurrency control source of truth.
  - Scheduled overlap safety remains in database locking and claim TTL logic rather than in Worker runtime.

## API or contracts

- Scheduled contract:
  - No public HTTP API is introduced for poller execution.
  - Worker shell forwards a scheduled-event envelope to Go/Wasm.
- Deployment/runtime config:
  - Mainnet worker name: `payrune-poller-mainnet`
  - Testnet4 worker name: `payrune-poller-testnet4`
  - Runtime config still includes:
    - `POLL_RESCHEDULE_INTERVAL`
    - `POLL_BATCH_SIZE`
    - `POLL_CLAIM_TTL`
    - Bitcoin Esplora network-specific secrets
  - Cloudflare deployment defaults are tuned for per-minute scheduled execution headroom:
    - `POLL_RESCHEDULE_INTERVAL = 10m`
    - `POLL_BATCH_SIZE = 1`
    - `POLL_CLAIM_TTL = 2m`
  - Cloudflare schedule frequency is configured separately from receipt reschedule interval.

## Backward compatibility (optional)

- API compatibility:
  - Not applicable; this is background runtime.
- Data migration compatibility:
  - Preserved; no schema redesign is required.
- Operational compatibility:
  - Existing compose pollers can continue to exist during migration if rollout prefers side-by-side validation.

## Failure modes and resiliency

- Missing `POSTGRES_CONNECTION_STRING`:
  - Scheduled run fails fast and logs deterministic error.
- Missing Esplora config for target network:
  - Scheduled run records deterministic observer failure through existing polling error path.
- Overlapping scheduled runs:
  - Existing database claim/lease logic prevents duplicate work from becoming inconsistent.
- Worker bootstrap failure:
  - Scheduled run fails with explicit bootstrap error in Worker logs.
- Bitcoin observer bridge failure:
  - Existing use case persists processing errors and reschedules according to current behavior.

## Observability

- Logs:
  - Scheduled runs log the same summary counters currently emitted by `bootstrap.RunPoller`.
  - Scope (`chain`, `network`) should be present in poller Worker logs.
- Metrics:
  - Existing output counters remain the logical source for operational metrics.
- Traces:
  - Not introduced in this slice.
- Alerts:
  - Existing worker/log-based monitoring can alert on repeated poll cycle failure.

## Security

- Authentication/authorization:
  - Not applicable to scheduled pollers.
- Secrets:
  - PostgreSQL and Esplora credentials live in Worker secrets.
- Abuse cases:
  - Public abuse surface is not in scope because this is scheduled runtime, not a public HTTP endpoint.

## Alternatives considered

- Option A:
  - Keep pollers on compose/VM and move only the API to Cloudflare.
- Option B:
  - Rewrite poller behavior in standalone JS Worker code.
- Option C:
  - Move pollers to standalone Cloudflare Workers but keep Go use cases as the source of truth.
- Why chosen:
  - Option C satisfies the Cloudflare-only goal without creating a second polling behavior implementation.

## Risks

- Risk:
  - Worker-compatible Bitcoin observer path may be more complex than the PostgreSQL bridge because it replaces current Go HTTP client assumptions.
- Mitigation:
  - Keep the JS shell generic and push observer behavior into a concrete Go adapter boundary plus explicit bridge contract.
- Risk:
  - Cloudflare Cron minimum granularity could create operational confusion versus `POLL_RESCHEDULE_INTERVAL`.
- Mitigation:
  - Keep scheduled frequency and receipt reschedule interval as clearly separate concepts in runtime config and docs.
