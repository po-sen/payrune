---
doc: 03_tasks
spec_date: 2026-04-11
slug: api-error-contract-alignment
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-07-ethereum-contract-readiness
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale:
  - This is a focused API-contract cleanup: controller mappings, controller tests, and OpenAPI alignment with no schema, migration, or new runtime integration.
- Upstream dependencies (`depends_on`):
  - `2026-04-07-ethereum-contract-readiness`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - No new flow or structural design work is needed.
  - What would trigger switching to Full mode:
    - If the change expanded into API versioning or a new response-envelope standard.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Validation steps are listed under each task below.

## Milestones

- M1:
  - Align controller status mappings for address-policy errors.
- M2:
  - Normalize `/health` error responses and update tests.
- M3:
  - Bring OpenAPI in line with the final runtime contract.

## Tasks (ordered)

1. T-001 - Align controller error mappings
   - Scope:
     - Update address-related controller status mappings so disabled policies stop returning `501`, and unsupported-but-valid policy identifiers no longer return `400`.
   - Output:
     - Updated controller mappings and controller tests for `/v1/chains/{chain}/addresses` and `/v1/chains/{chain}/payment-addresses`.
   - Linked requirements: FR-001, NFR-002, NFR-003, NFR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command):
       - `go test ./internal/adapters/inbound/http/controllers`
     - [ ] Expected result:
       - Controller tests assert `404` for `ErrAddressPolicyNotFound` and `409` for `ErrAddressPolicyNotEnabled`.
     - [ ] Logs/metrics to check (if applicable):
       - Mapped controller error logs still emit the chosen public status.
2. T-002 - Normalize health-controller error responses
   - Scope:
     - Replace plain-text health errors with the shared JSON error shape and update tests.
   - Output:
     - `/health` returns JSON error responses consistently for `405` and `500`.
   - Linked requirements: FR-002, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command):
       - `go test ./internal/adapters/inbound/http/controllers`
     - [ ] Expected result:
       - Health-controller tests pass with JSON error assertions.
     - [ ] Logs/metrics to check (if applicable):
       - Not applicable.
3. T-003 - Align OpenAPI error contract
   - Scope:
     - Update `deployments/swagger/openapi.yaml` to reflect the final controller mapping and improve shared error schema documentation.
   - Output:
     - OpenAPI with accurate `404` / `409` behavior, JSON health errors, and clearer error examples.
   - Linked requirements: FR-003, NFR-003, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command):
       - `bash scripts/precommit-run.sh`
       - Inspect `deployments/swagger/openapi.yaml`
     - [ ] Expected result:
       - Swagger validation passes and the documented contract matches the controllers.
     - [ ] Logs/metrics to check (if applicable):
       - Not applicable.

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-002
- FR-003 -> T-003
- NFR-002 -> T-001, T-002
- NFR-003 -> T-001, T-003
- NFR-005 -> T-001
- NFR-006 -> T-001, T-002, T-003

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Revert the controller/OpenAPI mapping commit if downstream clients depend on the old status contract.

## Validation evidence

- `go test ./internal/adapters/inbound/http/controllers`
- `go test ./...`
- `SPEC_DIR="specs/2026-04-11-api-error-contract-alignment" bash scripts/spec-lint.sh`
- `bash scripts/precommit-run.sh`
