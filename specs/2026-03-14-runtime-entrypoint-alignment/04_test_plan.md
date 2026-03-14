---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - Shared HTTP/runtime boundary cleanup, standalone scheduler handler reuse, worker cmd/bootstrap
    unification, and bootstrap naming normalization.
- Not covered:
  - Live Cloudflare deployment verification.

## Tests

### Unit

- TC-001:

  - Linked requirements: FR-001, FR-005, NFR-002, NFR-003
  - Steps:
    - Run the HTTP bridge and shared handler tests.
  - Expected:
    - Cloudflare API request bridging and shared HTTP handler assembly still return the expected
      response envelope and route behavior.

- TC-002:

  - Linked requirements: FR-002, FR-005, NFR-003
  - Steps:
    - Run the scheduler handler tests and inspect bootstrap code for direct scheduler DTO mapping.
  - Expected:
    - Scheduler request-to-use-case mapping remains in one inbound package only.

- TC-003:
  - Linked requirements: FR-003, FR-004, NFR-003
  - Steps:
    - Search worker `cmd/*` imports and bootstrap references for old names and direct adapter/DI
      imports.
  - Expected:
    - Worker commands stay thin, and old mixed bootstrap names are removed.

### Integration

- TC-101:

  - Linked requirements: FR-001, FR-002, FR-003, FR-005, NFR-001
  - Steps:
    - Run targeted Go tests for:
      - `./internal/adapters/inbound/http/...`
      - `./internal/adapters/inbound/scheduler`
      - `./internal/infrastructure/di`
      - `./internal/bootstrap`
      - `./cmd/api`
      - `./cmd/api-worker`
      - `./cmd/poller`
      - `./cmd/poller-worker`
      - `./cmd/webhook-dispatcher`
      - `./cmd/webhook-dispatcher-worker`
  - Expected:
    - Touched runtime packages compile and test cleanly.

- TC-102:
  - Linked requirements: FR-001, FR-002, FR-003, FR-004, NFR-001, NFR-003
  - Steps:
    - Run `go list ./...` and full `go test ./...`.
  - Expected:
    - The full repo import graph and test suite remain green.

## Edge cases and failure modes

- Case:
  - Shared HTTP or scheduler handlers are nil after DI/bootstrap refactors.
- Expected behavior:

  - Runtime helpers return configuration errors instead of panicking.

- Case:
  - One rename lands without all direct call-site updates.
- Expected behavior:
  - Compile/test/search checks fail immediately.

## NFR verification

- Reliability:
  - Targeted and full Go verification commands pass.
- Security:
  - Public HTTP middleware and validation remain on the shared API path.
- Maintainability:
  - Runtime entrypoint responsibilities and bootstrap names are consistent, and one merged spec
    documents the whole refactor.
