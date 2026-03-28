---
doc: 01_requirements
spec_date: 2026-03-28
slug: allocate-usecase-readability
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-26-allocate-usecase-decomposition
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Requirements

## Glossary (optional)

- Side-effect error state:
- A mutable variable outside the transaction callback that is used later to decide the final outward error, instead of returning that error directly through the main flow.

## Out-of-scope behaviors

- OOS1: No outbound contract change for allocate payment address
- OOS2: No UnitOfWork redesign

## Functional requirements

### FR-001 - Allocation transaction flow must read as one straight-line workflow

- Description: `issueAllocation(...)` should keep claim/reserve/derive/issue/persist/tracking/idempotency-complete in one visible order, without hidden side-effect state for final error selection.
- Acceptance criteria:
  - [x] Derivation failure handling no longer depends on a mutable outer error variable to decide the final return value.
  - [x] The transaction callback returns the same final outward error that the caller should observe after failure persistence succeeds.
- Notes: keep the transaction boundary and stage ordering intact.

### FR-002 - Readability refactor must preserve existing behavior

- Description: the cleanup must not change idempotency, allocation reservation, derivation failure persistence, or receipt tracking behavior.
- Acceptance criteria:
  - [x] Existing allocate payment address tests continue to pass without intended behavior drift.
  - [x] Full repo tests and precommit validation pass after the refactor.
- Notes: this is readability cleanup, not policy or persistence redesign.

## Non-functional requirements

- Performance (NFR-001): No extra DB calls, no extra transactions, and no new retries beyond the existing flow.
- Availability/Reliability (NFR-002): Existing allocate-related tests and full repo tests pass unchanged in intent.
- Security/Privacy (NFR-003): No new outward error detail is introduced.
- Compliance (NFR-004):
- Observability (NFR-005): Existing tests remain the primary regression signal for the allocation workflow.
- Maintainability (NFR-006): A reviewer can read `issueAllocation(...)` top-to-bottom without tracking a second error channel outside the transaction callback.

## Dependencies and integrations

- External systems:
- Internal services: `internal/application/usecases/allocate_payment_address_use_case.go`, allocate-related usecase tests
