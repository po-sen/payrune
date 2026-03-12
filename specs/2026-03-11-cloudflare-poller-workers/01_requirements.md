---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Standalone poller Worker:
  - A Cloudflare Worker that runs scheduled receipt polling directly and does not depend on a separately running Go poller process.
- Polling cycle:
  - One execution of `RunReceiptPollingCycleUseCase` for a configured chain/network scope.

## Out-of-scope behaviors

- OOS1:
  - Receipt webhook dispatching on Cloudflare.
- OOS2:
  - Replacing PostgreSQL or Esplora with new providers.

## Functional requirements

### FR-001 - Standalone Cloudflare poller workers

- Description:
  - Payrune must run both Bitcoin pollers directly on Cloudflare Workers.
- Acceptance criteria:
  - [ ] A mainnet Worker deployment exists for `payrune-poller-mainnet`.
  - [ ] A testnet4 Worker deployment exists for `payrune-poller-testnet4`.
  - [ ] Neither deployment requires a separately running Go poller process.
- Notes:
  - These may share one codebase and deploy as separate worker names or Wrangler environments.

### FR-002 - Execute existing Go polling use case

- Description:
  - Cloudflare pollers must execute the existing Go application layer for receipt polling behavior.
- Acceptance criteria:
  - [ ] Polling cycles are driven by the existing `RunReceiptPollingCycleUseCase`.
  - [ ] Current receipt lifecycle semantics remain in Go rather than being reimplemented in standalone JS.
  - [ ] Shared tip-height reuse, final-check expiry ordering, and current status persistence continue to come from the existing Go logic.
- Notes:
  - This slice should reuse the current Go polling source of truth just as the Cloudflare API Worker reuses Go API use cases.

### FR-003 - Scheduled runtime model

- Description:
  - Cloudflare must trigger the pollers through scheduled handlers instead of process-local tickers.
- Acceptance criteria:
  - [ ] The Worker runtime supports Cloudflare scheduled execution.
  - [ ] A mainnet scheduled invocation runs one polling cycle scoped to `chain=bitcoin`, `network=mainnet`.
  - [ ] A testnet4 scheduled invocation runs one polling cycle scoped to `chain=bitcoin`, `network=testnet4`.
  - [ ] Poll scheduling remains separate from receipt `next_poll_at` reschedule semantics.
- Notes:
  - The Worker trigger frequency is operational scheduling; it must not replace `next_poll_at` logic.

### FR-004 - Deployment shell only

- Description:
  - Cloudflare poller deployment code must remain a thin shell around the actual runtime behavior.
- Acceptance criteria:
  - [ ] `deployments/cloudflare/payrune-poller/` contains Wrangler config, thin bootstrap entrypoint, Go-Wasm loader, JS bridge(s), package metadata, and deployment-focused tests/docs only.
  - [ ] Polling behavior, scope resolution, and orchestration live outside `deployments/`.
- Notes:
  - Future poller feature work should primarily change Go code.

### FR-005 - Deploy and teardown entrypoints

- Description:
  - The repo must provide simple deploy and delete flows for both poller workers.
- Acceptance criteria:
- [ ] `make cf-up` runs the shared Cloudflare migration and deploys both poller workers as part of the unified Cloudflare rollout flow.
- [ ] `make cf-down` deletes both poller workers as part of the unified Cloudflare teardown flow.
- [ ] `make cf-migrate` runs the shared Cloudflare PostgreSQL migration independently.
- [ ] Cloudflare deploy/migrate scripts auto-load repo-local `.env.cloudflare` when present.
- [ ] The repo provides `.env.cloudflare.example` as the local Cloudflare env template.
- [ ] Non-sensitive poller cadence and Esplora endpoint defaults live in Worker config, not in deploy-time secret prompts.
- [ ] Cloudflare Worker defaults prioritize scheduled-run CPU headroom over per-run throughput, with `POLL_BATCH_SIZE=1` as the baseline default.
- [ ] `POSTGRES_CONNECTION_STRING` is required for `make cf-up` and `make cf-migrate`; missing it must fail fast instead of prompting.
- [ ] Optional Esplora auth secrets may be left blank and skipped during Worker secret sync.
- [ ] Cloudflare deployment docs and ignore rules reference only repo-root `.env.cloudflare`, not deployment-local `.env.local` files.
- [ ] Deploy clearly announces that `POSTGRES_CONNECTION_STRING` Worker secret sync will run before build/test/deploy steps.
- [ ] The deploy flow builds the Go-Wasm poller artifact before publishing the Worker.

### FR-006 - Worker-side PostgreSQL adapter reuse

- Description:
  - The Worker poller must use a Cloudflare-compatible PostgreSQL adapter so the Go polling use case can run without `database/sql` at runtime.
- Acceptance criteria:
  - [ ] Claim/save polling paths run through a Worker-compatible PostgreSQL adapter.
  - [ ] The poller Worker runtime does not depend on `database/sql` or a separately running Go process.
- Notes:
  - Existing `cloudflarepostgres` work should be reused where appropriate.

### FR-007 - Worker-side Bitcoin observer adapter

- Description:
  - The poller Worker must provide a Cloudflare-compatible Bitcoin Esplora observer path for the existing polling use case.
- Acceptance criteria:
  - [ ] The Worker runtime can fetch latest block height for a configured Bitcoin network.
  - [ ] The Worker runtime can observe address receipt totals using the current Esplora-compatible contract.
  - [ ] Missing Esplora configuration still produces deterministic poller errors rather than silent success.
- Notes:
  - This slice needs the equivalent of the current Bitcoin observer, but runnable inside Worker runtime.

### FR-008 - Poll-cycle output parity

- Description:
  - Scheduled poller runs must preserve current output counters and scope-aware logging.
- Acceptance criteria:
  - [ ] Mainnet/testnet4 scheduled runs log or emit current counters for `claimed`, `updated`, `terminal_failed`, and `processing_errors`.
  - [ ] Scheduled runs preserve current scope-specific behavior rather than mixing networks.

## Non-functional requirements

- Deployment decoupling (NFR-001):
  - Future poller behavior changes should usually not require editing `deployments/cloudflare/payrune-poller/`.
- Runtime simplicity (NFR-002):
  - The standalone poller should use the smallest practical Cloudflare runtime model while still executing Go polling use cases.
- Verification (NFR-003):
  - The slice must be verifiable with focused local checks, Worker syntax/tests, and dry-run deploy checks.
- Reliability (NFR-004):
  - Scheduled overlap must continue to be safe because claim/lease semantics remain enforced in PostgreSQL.
- Operability (NFR-005):
  - Mainnet and testnet4 pollers must remain independently deployable and independently deletable.

## Dependencies and integrations

- External systems:
  - Cloudflare Workers Cron Triggers.
  - PostgreSQL.
  - Bitcoin Esplora-compatible HTTP APIs.
- Internal services:
  - `RunReceiptPollingCycleUseCase`
  - existing receipt lifecycle/domain rules
  - existing poller persistence contracts
