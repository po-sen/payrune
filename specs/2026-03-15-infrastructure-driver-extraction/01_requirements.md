---
doc: 01_requirements
spec_date: 2026-03-15
slug: infrastructure-driver-extraction
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on: []
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Requirements

## Out-of-scope behaviors

- OOS1:
  - Changing Cloudflare worker runtime contracts or scheduler request formats.
- OOS2:
  - Moving outbound PostgreSQL adapters out of `internal/adapters/outbound/persistence/postgres`.
- OOS3:
  - Moving Bitcoin Cloudflare transport bridges in this iteration.

## Functional requirements

### FR-001 - Standalone PostgreSQL connection setup lives in drivers

- Description:
  - The low-level PostgreSQL connection open-and-ping logic used by standalone DI containers must
    live in `internal/infrastructure/drivers`.
- Acceptance criteria:
  - [ ] `internal/infrastructure/drivers` contains a concrete PostgreSQL connection helper.
  - [ ] `internal/infrastructure/di/container.go` no longer calls `sql.Open` directly.
  - [ ] `internal/infrastructure/di/poller_container.go` no longer calls `sql.Open` directly.
  - [ ] `internal/infrastructure/di/receipt_webhook_dispatcher_container.go` no longer calls
        `sql.Open` directly.
- Notes:
  - The helper may still rely on the standard library `database/sql` and the pq driver import.

### FR-002 - DI retains composition ownership

- Description:
  - DI containers must continue to own adapter construction, use-case wiring, and non-driver env
    parsing.
- Acceptance criteria:
  - [ ] DI containers still assemble repositories/adapters/use cases after obtaining a database
        handle from the driver helper.
  - [ ] The new driver package does not import application, domain, or adapter packages.
- Notes:
  - This refactor should move only low-level connection setup.

### FR-003 - Standalone runtime behavior remains unchanged

- Description:
  - The standalone API, poller, and receipt webhook dispatcher runtimes must keep the same behavior
    after the driver extraction.
- Acceptance criteria:
  - [ ] Relevant targeted Go tests pass after the extraction.
  - [ ] `go list ./...` passes.
  - [ ] Full `go test ./...` passes.
- Notes:
  - This is a maintainability refactor, not a runtime contract change.

### FR-004 - Cloudflare Postgres raw bridge implementation lives in drivers

- Description:
  - The raw Cloudflare Postgres JS bridge implementation must live in
    `internal/infrastructure/drivers`, not in the adapter package.
- Acceptance criteria:
  - [ ] The Cloudflare Postgres bridge interface remains available to the adapter package.
  - [ ] The JS/WASM and unsupported `NewJSBridge` implementation no longer lives under
        `internal/adapters/outbound/persistence/cloudflarepostgres`.
  - [ ] Cloudflare DI builders obtain the raw bridge from an infrastructure driver package and
        inject it into the Cloudflare Postgres adapter.
- Notes:
  - This move covers only the raw bridge implementation, not the executor, unit of work, or stores.

### FR-005 - Cloudflare runtime behavior remains unchanged

- Description:
  - Cloudflare API, poller, and webhook dispatcher runtimes must keep the same behavior after the
    raw bridge move.
- Acceptance criteria:
  - [ ] `internal/infrastructure/di/cloudflare_api_worker.go` still builds a working runtime.
  - [ ] `internal/infrastructure/di/cloudflare_poller_worker.go` still builds a working runtime.
  - [ ] `internal/infrastructure/di/cloudflare_webhook_dispatcher_worker.go` still builds a
        working runtime.
  - [ ] Relevant targeted Go tests and full `go test ./...` pass after the move.
- Notes:
  - This is still a boundary correction, not a contract change.

### FR-006 - Cloudflare webhook raw bridge implementation lives in drivers

- Description:
  - The raw Cloudflare webhook JS bridge implementation must live in
    `internal/infrastructure/drivers`, not in the adapter package.
- Acceptance criteria:
  - [ ] The Cloudflare webhook bridge interface remains available to the adapter package.
  - [ ] The JS/WASM and unsupported webhook bridge implementation no longer lives under
        `internal/adapters/outbound/webhook`.
  - [ ] Cloudflare webhook DI wiring obtains the raw bridge from an infrastructure driver package
        and injects it into the notifier adapter.
- Notes:
  - This move covers only the raw JS transport bridge, not the notifier adapter.

### FR-007 - Cloudflare webhook runtime behavior remains unchanged

- Description:
  - The Cloudflare webhook dispatcher runtime must keep the same behavior after the raw webhook
    bridge move.
- Acceptance criteria:
  - [ ] `internal/infrastructure/di/cloudflare_webhook_dispatcher_worker.go` still builds a working
        runtime.
  - [ ] Relevant targeted Go tests and full `go test ./...` pass after the move.
- Notes:
  - This is also a boundary correction, not a runtime contract change.

## Non-functional requirements

- Reliability (NFR-001):
  - Targeted and full Go verification commands must pass after the extraction.
- Maintainability (NFR-002):
  - Repeated or misplaced low-level driver code should move behind concrete infrastructure driver
    packages without pulling adapter semantics into infrastructure.

## Dependencies and integrations

- External systems:
  - PostgreSQL via `database/sql` and the pq driver.
  - Cloudflare Worker JS/WASM bridge globals for Postgres operations.
  - Cloudflare Worker JS/WASM bridge globals for webhook posting.
- Internal services:
  - `internal/infrastructure/di`
  - `internal/infrastructure/drivers`
