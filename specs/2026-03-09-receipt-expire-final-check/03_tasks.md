---
doc: 03_tasks
spec_date: 2026-03-09
slug: receipt-expire-final-check
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-05-blockchain-receipt-polling-service
  - 2026-03-09-sticky-paid-unconfirmed-status
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Task Plan

## Mode decision

- Selected mode: Full
- Rationale:
  - This change affects receipt polling flow ordering and failure behavior inside an async worker,
    so design and explicit tests are needed.
- Upstream dependencies (`depends_on`):
  - `2026-03-05-blockchain-receipt-polling-service`
  - `2026-03-09-sticky-paid-unconfirmed-status`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - Not applicable.
  - What would trigger switching to Full mode:
    - Already selected.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Not applicable.

## Milestones

- M1:
  - Specify the new final-check expiry behavior.
- M2:
  - Update claim and polling flow, then validate domain, use case, and store behavior.

## Tasks (ordered)

1. T-001 - Specify final-check expiry behavior
   - Scope:
     - Capture the new rule that expiry is evaluated only in a due poll cycle after the final
       observation finishes.
   - Output:
     - `specs/2026-03-09-receipt-expire-final-check/*.md`
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, NFR-002, NFR-005, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-09-receipt-expire-final-check" bash scripts/spec-lint.sh`
     - [x] Expected result: spec lint passes and the final-check expiry rule is explicit.
     - [x] Logs/metrics to check (if applicable): none
2. T-002 - Move expiry evaluation to the post-observation path
   - Scope:
     - Remove query-side early expiry claims and reorder the polling use case so expiry is decided
       only after observation succeeds.
   - Output:
     - `internal/adapters/outbound/persistence/postgres/payment_receipt_tracking_store.go`
     - `internal/application/use_cases/run_receipt_polling_cycle_use_case.go`
     - `internal/domain/policies/payment_receipt_tracking_lifecycle.go`
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, NFR-001, NFR-002, NFR-005, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `GOCACHE=/tmp/go-build go test ./internal/application/use_cases ./internal/domain/policies ./internal/adapters/outbound/persistence/postgres -count=1`
     - [x] Expected result: due claims follow `next_poll_at`, successful final observations can prevent expiry, and observer failures after expiry remain retryable.
     - [x] Logs/metrics to check (if applicable): none
3. T-003 - Add targeted regression coverage
   - Scope:
     - Cover final-check expiry timing in unit and integration-level tests.
   - Output:
     - `internal/application/use_cases/run_receipt_polling_cycle_use_case_test.go`
     - `internal/domain/policies/payment_receipt_tracking_lifecycle_test.go`
     - `internal/adapters/outbound/persistence/postgres/payment_receipt_tracking_store_test.go`
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, NFR-002, NFR-005
   - Validation:
     - [x] How to verify (manual steps or command): `GOCACHE=/tmp/go-build go test ./internal/application/use_cases ./internal/domain/policies ./internal/adapters/outbound/persistence/postgres -count=1`
     - [x] Expected result: new tests fail without the flow change and pass with it.
     - [x] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001, T-002, T-003
- FR-002 -> T-001, T-002, T-003
- FR-003 -> T-001, T-002, T-003
- FR-004 -> T-001, T-002, T-003
- NFR-001 -> T-002
- NFR-002 -> T-001, T-002, T-003
- NFR-005 -> T-001, T-002, T-003
- NFR-006 -> T-001, T-002

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Restore query-side expiry claims and pre-observation expiry ordering if the new final-check rule
    proves operationally undesirable.
