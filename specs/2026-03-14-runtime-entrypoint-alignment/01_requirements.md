---
doc: 01_requirements
spec_date: 2026-03-14
slug: runtime-entrypoint-alignment
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-10-cloudflare-workers-postgres
  - 2026-03-11-cloudflare-poller-workers
  - 2026-03-12-api-worker-naming-unification
  - 2026-03-13-cloudflare-webhook-dispatcher-worker
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Requirements

## Out-of-scope behaviors

- OOS1:
  - Changing API response schemas, worker JSON field names, or environment variable names.
- OOS2:
  - Changing domain rules, use-case orchestration semantics, or outbound bridge mechanics.

## Functional requirements

### FR-001 - Public HTTP entrypoints share one HTTP assembly boundary

- Description:
  - Standalone API and Cloudflare API worker runtimes must build their public HTTP surface from one
    explicit inbound HTTP assembly point.
- Acceptance criteria:
  - [ ] A shared handler builder exists under `internal/adapters/inbound/http`.
  - [ ] The Cloudflare API request bridge lives under the HTTP inbound namespace.
  - [ ] Standalone API bootstrap and the Cloudflare API worker both use the shared HTTP handler
        assembly point.

### FR-002 - Scheduler cycle mapping lives in one inbound package

- Description:
  - One-cycle scheduler mapping must live in `internal/adapters/inbound/scheduler` and be reused by
    both Cloudflare worker runtimes and standalone scheduler runtimes.
- Acceptance criteria:
  - [ ] Poller and webhook dispatcher cycle mapping lives in `internal/adapters/inbound/scheduler`.
  - [ ] Standalone bootstrap loops invoke scheduler handlers rather than mapping use-case DTOs
        directly.
  - [ ] Standalone DI containers expose the scheduler handlers needed by bootstrap.

### FR-003 - Worker command packages delegate through bootstrap

- Description:
  - Cloudflare worker `cmd/*` entrypoints must delegate request orchestration through
    `internal/bootstrap` instead of importing inbound adapters and DI directly.
- Acceptance criteria:
  - [ ] `cmd/api-worker/main.go` delegates through bootstrap.
  - [ ] `cmd/poller-worker/main.go` delegates through bootstrap.
  - [ ] `cmd/webhook-dispatcher-worker/main.go` delegates through bootstrap.
  - [ ] Worker command packages no longer import inbound adapter or DI packages directly.

### FR-004 - Bootstrap naming uses consistent runtime nouns

- Description:
  - Bootstrap file names, exported functions, and config types must use consistent runtime nouns.
- Acceptance criteria:
  - [ ] API bootstrap uses `api.go` and `RunAPI`.
  - [ ] Receipt webhook dispatcher bootstrap naming uses `ReceiptWebhookDispatcher` consistently
        across config types and worker helpers.
  - [ ] Old mixed names (`app`, `Dispatch`, missing `Receipt`) are removed from direct code
        references.

### FR-005 - Runtime behavior stays stable

- Description:
  - The refactor must preserve runtime behavior and contracts.
- Acceptance criteria:
  - [ ] Public API route registration still exposes the same paths and methods.
  - [ ] Worker request/response JSON contracts remain unchanged.
  - [ ] Scheduler loops keep the same lifecycle responsibility and output counters.

## Non-functional requirements

- Reliability (NFR-001):
  - Targeted Go tests, `go list ./...`, and full `go test ./...` must pass after the refactor.
- Security/Privacy (NFR-002):
  - Public HTTP middleware and validation behavior must remain aligned across API runtimes.
- Maintainability (NFR-003):
  - Runtime entrypoint responsibilities and bootstrap naming should be predictable without extra
    explanation.

## Dependencies and integrations

- External systems:
  - Cloudflare Worker JS/WASM invocation payloads.
  - `net/http` public API transport.
- Internal services:
  - `internal/bootstrap`
  - `internal/adapters/inbound/http`
  - `internal/adapters/inbound/scheduler`
  - `internal/infrastructure/di`
