---
doc: 04_test_plan
spec_date: 2026-03-08
slug: container-store-injection-refactor
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-08-payment-address-idempotency-key
  - 2026-03-08-payment-address-status-api
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Container Store Injection Refactor - Test Plan

## Scope

- Covered:
  - Allocate use-case constructor/dependency refactor.
  - Replay lookup and duplicate-claim recovery behavior after moving replay reads under `UnitOfWork`.
  - API container wiring cleanup and compile/test validation.
- Not covered:
  - New business behavior, because none is intended.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, NFR-001, NFR-002, NFR-006
  - Steps:
    - Run allocate use-case tests covering fresh success, replay hit, conflicting replay, duplicate-claim recovery, and derivation failure.
  - Expected:
    - All existing payment-address behaviors remain unchanged after removing direct store injection.

### Integration

- TC-101:
  - Linked requirements: FR-002, NFR-005, NFR-006
  - Steps:
    - Run DI, controller, and Postgres persistence suites after cleaning up container wiring.
  - Expected:
    - The API container compiles and wires payment flows correctly without named local allocation/idempotency store assignments.

## Edge cases and failure modes

- Case:
  - A duplicate idempotency claim occurs before the original transaction completes.
  - Expected behavior:
    - Recovery path still returns the replayed allocation when the completed record becomes available, or a deterministic internal error if the idempotency record remains incomplete.

## NFR verification

- Performance:
  - Confirm the fresh issuance path still uses one main write transaction and no extra persistence work beyond replay lookup behavior.
- Reliability:
  - Confirm replay and conflict behavior remain unchanged.
- Security:
  - Confirm no new data is exposed or persisted.
