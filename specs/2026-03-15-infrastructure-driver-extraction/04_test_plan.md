---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - Concrete PostgreSQL driver extraction for standalone runtime connection setup.
  - Cloudflare Postgres raw JS bridge extraction into infrastructure drivers.
  - Cloudflare webhook raw JS bridge extraction into infrastructure drivers.
- Not covered:
  - Bitcoin Cloudflare bridge extraction.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, FR-002, NFR-002
  - Steps:
    - Run unit tests for the new PostgreSQL driver helper.
  - Expected:
    - Successful connection creation is returned when the helper can open and ping a database, and
      invalid input paths return explicit errors.
- TC-002:
  - Linked requirements: FR-001, FR-002, NFR-001
  - Steps:
    - Run DI container tests that exercise the new helper-backed construction path.
  - Expected:
    - Container construction still succeeds with valid dependencies and surfaces startup errors.

### Integration

- TC-101:
  - Linked requirements: FR-003, FR-005, FR-007, NFR-001, NFR-002
  - Steps:
    - Run:
      - `go test ./internal/infrastructure/drivers/... ./internal/adapters/outbound/persistence/cloudflarepostgres ./internal/adapters/outbound/webhook ./internal/infrastructure/di ./internal/bootstrap ./cmd/api ./cmd/poller ./cmd/webhook-dispatcher ./cmd/webhook-dispatcher-worker`
      - `go list ./...`
      - `go test ./...`
  - Expected:
    - Standalone and Cloudflare runtime wiring continue to compile and all tests remain green.

## Edge cases and failure modes

- Case:
  - `DATABASE_URL` is missing or blank.
- Expected behavior:
  - The driver helper returns a clear error before any container-specific wiring proceeds.

## NFR verification

- Reliability:
  - Targeted and full Go verification commands pass.
- Maintainability:
  - Repeated standalone PostgreSQL open-and-ping code is removed from DI containers, and the
    Cloudflare Postgres and webhook raw bridges no longer live in adapter packages.
