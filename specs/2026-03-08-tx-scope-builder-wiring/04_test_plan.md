---
doc: 04_test_plan
spec_date: 2026-03-08
slug: tx-scope-builder-wiring
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-07-architecture-naming-refactor
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Tx Scope Builder Wiring - Test Plan

## Scope

- Covered:
  - Postgres `UnitOfWork` transaction behavior after internalizing tx-scope assembly.
  - Container wiring compile/test coverage after removing tx-scope builder/factory injection.
- Not covered:
  - New functional behavior, because none is introduced.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-002, NFR-002, NFR-006
  - Steps:
    - Run `UnitOfWork` tests after internalizing tx-scope assembly.
  - Expected:
    - Success path commits, error path rolls back, and callback scope still contains the required stores.

### Integration

- TC-101:
  - Linked requirements: FR-001, FR-002, NFR-001, NFR-006
  - Steps:
    - Run Postgres persistence and DI tests after removing tx-scope builder/factory injection from containers.
  - Expected:
    - Production wiring still constructs the same tx-bound stores and packages compile cleanly.

## Edge cases and failure modes

- Case:
  - `UnitOfWork` is created without a database.
  - Expected behavior:
    - `WithinTransaction` returns a deterministic configuration error.

## NFR verification

- Performance:
  - Confirm no extra query or transaction step is introduced.
- Reliability:
  - Confirm commit/rollback tests are unchanged in behavior.
- Security:
  - Confirm the refactor does not add new secret or env handling.
