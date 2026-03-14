---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - API worker bridge relocation from inbound to bootstrap.
- Not covered:
  - Any changes to non-API-worker runtime paths.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, FR-003, NFR-001
  - Steps:
    - Run bootstrap API worker tests after moving the request bridge logic.
  - Expected:
    - Request envelope decoding and `status`/`headers`/`body` response mapping still behave the
      same.

### Integration

- TC-101:
  - Linked requirements: FR-002, FR-003, NFR-001, NFR-002
  - Steps:
    - Run:
      - `go test ./internal/bootstrap ./cmd/api-worker`
      - `go list ./...`
      - `go test ./...`
      - `rg -n "internal/adapters/inbound/http/cloudflare" cmd internal`
  - Expected:
    - The old package import path is gone and the repo still builds/tests cleanly.

## Edge cases and failure modes

- Case:
  - The handler passed to the bootstrap-local bridge is nil.
- Expected behavior:
  - The helper returns a clear configuration error.

## NFR verification

- Reliability:
  - Targeted and full Go verification commands pass.
- Maintainability:
  - The API worker bridge no longer sits in `internal/adapters/inbound`.
