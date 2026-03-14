---
doc: 01_requirements
spec_date: 2026-03-15
slug: api-worker-bridge-bootstrap
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-14-runtime-entrypoint-alignment
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
  - Changing the shared public HTTP handler assembly under `internal/adapters/inbound/http`.
- OOS2:
  - Changing Cloudflare worker JSON field names or API worker JS global names.

## Functional requirements

### FR-001 - API worker request bridge lives in bootstrap

- Description:
  - The Cloudflare API request bridging logic must live in `internal/bootstrap` rather than
    `internal/adapters/inbound`.
- Acceptance criteria:
  - [ ] `internal/bootstrap` contains the request/response bridge logic needed by the API worker.
  - [ ] `internal/bootstrap/api_worker.go` no longer imports
        `internal/adapters/inbound/http/cloudflare`.
- Notes:
  - The bridge may stay in the same bootstrap file or a nearby bootstrap-specific file.

### FR-002 - The old inbound bridge package is removed

- Description:
  - The old `internal/adapters/inbound/http/cloudflare` bridge package must be removed.
- Acceptance criteria:
  - [ ] `internal/adapters/inbound/http/cloudflare/bridge.go` no longer exists.
  - [ ] `internal/adapters/inbound/http/cloudflare/bridge_test.go` no longer exists.
  - [ ] No Go source imports `internal/adapters/inbound/http/cloudflare`.
- Notes:
  - This is a package-boundary cleanup, not a behavior change.

### FR-003 - API worker runtime behavior remains stable

- Description:
  - The API worker must continue to decode the same payload and return the same response envelope.
- Acceptance criteria:
  - [ ] The API worker still accepts `request`, `env`, and `bridgeId` in the JSON payload.
  - [ ] The API worker still returns `status`, `headers`, and `body` in the response envelope.
  - [ ] Existing targeted and full Go tests pass after the move.
- Notes:
  - The helper may be re-tested under `internal/bootstrap`.

## Non-functional requirements

- Reliability (NFR-001):
  - Targeted tests and full `go test ./...` must pass.
- Maintainability (NFR-002):
  - `internal/adapters/inbound` should contain direct inbound transport adapters rather than
    bootstrap/runtime bridges.

## Dependencies and integrations

- External systems:
  - Cloudflare Worker JS/WASM API worker payload contract.
- Internal services:
  - `internal/bootstrap`
  - `internal/infrastructure/di`
  - `http.Handler`-based public API transport
