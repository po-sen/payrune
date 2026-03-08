---
doc: 03_tasks
spec_date: 2026-03-08
slug: payment-address-status-api
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-04-policy-payment-address-allocation
  - 2026-03-05-blockchain-receipt-polling-service
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Payment Address Status API - Task Plan

## Mode decision

- Selected mode: Quick
- Rationale:
  - The change adds one read endpoint and a read-side query port without a migration, new external integration, or risky async behavior change.
- Upstream dependencies (`depends_on`):
  - `2026-03-04-policy-payment-address-allocation`
  - `2026-03-05-blockchain-receipt-polling-service`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - The data already exists, and the new work is a straightforward read model plus controller/use-case wiring.
  - What would trigger switching to Full mode:
    - A new persistence model, authentication flow, or webhook/query consistency mechanism.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): not applicable; `04_test_plan.md` is included.

## Milestones

- M1:
  - Define the API contract and read-model boundaries for payment status lookup.
- M2:
  - Implement and verify the read endpoint, use case, persistence query, and Swagger contract.

## Tasks (ordered)

1. T-001 - Finalize the payment status read API spec
   - Scope:
     - Capture endpoint shape, payload fields, and error behavior for querying payment status by `paymentAddressId`.
   - Output:
     - `specs/2026-03-08-payment-address-status-api/*.md`
   - Linked requirements: FR-001, FR-002, FR-003, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-08-payment-address-status-api" bash scripts/spec-lint.sh`
     - [x] Expected result: spec lint passes and documents consistently describe the read endpoint.
     - [x] Logs/metrics to check (if applicable): none
2. T-002 - Implement read-side query port, use case, controller route, and Swagger contract
   - Scope:
     - Add a query-style outbound port and Postgres finder for one payment status view.
     - Add a use case and HTTP handler for `GET /v1/chains/{chain}/payment-addresses/{paymentAddressId}`.
     - Document the response in Swagger.
   - Output:
     - Application, adapter, DI, and Swagger changes for the new endpoint.
   - Linked requirements: FR-001, FR-002, FR-003, NFR-001, NFR-002, NFR-003, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `GOCACHE=/tmp/go-build go test ./internal/application/use_cases ./internal/adapters/inbound/http/controllers ./internal/adapters/outbound/persistence/postgres ./internal/infrastructure/di -count=1`
     - [x] Expected result: targeted suites pass and cover success, not-found, invalid-id, and controller mapping behavior.
     - [x] Logs/metrics to check (if applicable): none
3. T-003 - Verify contract docs and final status behavior
   - Scope:
     - Run contract validation and ensure the response shape matches the implemented endpoint.
   - Output:
     - Verified spec and Swagger behavior for the payment status API.
   - Linked requirements: FR-001, FR-002, FR-003, NFR-005
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-08-payment-address-status-api" bash scripts/spec-lint.sh` and `GOCACHE=/tmp/go-build go list ./...`
     - [x] Expected result: spec lint and package listing pass after the new endpoint is wired.
     - [x] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001, T-002, T-003
- FR-002 -> T-001, T-002, T-003
- FR-003 -> T-001, T-002, T-003
- NFR-001 -> T-002
- NFR-002 -> T-002
- NFR-003 -> T-002
- NFR-005 -> T-003
- NFR-006 -> T-001, T-002

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Revert the new read endpoint wiring and Swagger contract if behavior is incorrect.
