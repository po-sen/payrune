---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Tx scope:
  - The bundle of transaction-scoped Postgres stores exposed inside one `UnitOfWork` callback.

## Out-of-scope behaviors

- OOS1:
  - New transaction lifecycle features.
- OOS2:
  - New runtime configuration or env handling.

## Functional requirements

### FR-001 - Remove tx-scope wiring from container constructors

- Description:
  - Production wiring must not require containers to assemble or inject tx-scope builders/factories for the fixed Postgres transaction scope.
- Acceptance criteria:
  - [ ] `internal/infrastructure/di/container.go`, `poller_container.go`, and `receipt_webhook_dispatcher_container.go` call `postgresadapter.NewUnitOfWork(db)` directly.
  - [ ] Production code no longer relies on tx-scope builders or factories.
  - [ ] `UnitOfWork` still assembles the same tx-scoped stores currently required by `outport.TxScope`.
- Notes:
  - Tx-scope construction still happens only after `WithinTransaction` has created a `*sql.Tx`.

### FR-002 - Preserve transaction orchestration behavior

- Description:
  - The refactor must not change `UnitOfWork` commit/rollback semantics or the stores exposed inside the callback.
- Acceptance criteria:
  - [ ] `WithinTransaction` still begins a transaction, invokes the callback with a populated `TxScope`, commits on success, and rolls back on callback error.
  - [ ] Unit tests continue to assert non-nil allocation, idempotency, receipt tracking, and outbox stores in the callback scope.
- Notes:
  - This is a behavior-preserving refactor.

## Non-functional requirements

- Performance (NFR-001):
  - No extra DB round trips or transaction steps are introduced.
- Availability/Reliability (NFR-002):
  - Existing transaction success/failure behavior remains unchanged.
- Security/Privacy (NFR-003):
  - No new secrets or sensitive data handling is introduced.
- Compliance (NFR-004):
  - None.
- Observability (NFR-005):
  - Existing transaction logging/behavior is unchanged.
- Maintainability (NFR-006):
  - Constructor code should stay assign-oriented and avoid transaction-scope wiring ceremony when the underlying adapter composition is fixed.

## Dependencies and integrations

- External systems:
  - None.
- Internal services:
  - `internal/adapters/outbound/persistence/postgres`
  - `internal/infrastructure/di`
