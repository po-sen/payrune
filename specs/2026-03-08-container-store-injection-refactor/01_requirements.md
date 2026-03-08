---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Replay lookup:
  - The lookup path that finds an existing idempotency record and issued allocation to return a replayed payment-address response.

## Out-of-scope behaviors

- OOS1:
  - Renaming every persistence port in the payment flow.
- OOS2:
  - Introducing new read-side infrastructure beyond the existing status finder.

## Functional requirements

### FR-001 - Remove direct replay-store injection from allocate use case

- Description:
  - `AllocatePaymentAddressUseCase` must no longer require DB-scoped allocation/idempotency stores in its constructor.
- Acceptance criteria:
  - [ ] `NewAllocatePaymentAddressUseCase` no longer accepts `PaymentAddressAllocationStore` or `PaymentAddressIdempotencyStore`.
  - [ ] Replay lookup and duplicate-claim recovery still return the same replayed allocation response or `409 Conflict` behavior as before.
  - [ ] Transaction-internal issuance logic continues to use tx-scoped stores from `UnitOfWork`.
- Notes:
  - This refactor targets dependency shape, not externally visible behavior.

### FR-002 - Simplify container wiring for payment flows

- Description:
  - The API container should stop creating named local variables for stores that are no longer independently needed.
- Acceptance criteria:
  - [ ] `container.go` no longer has named `allocationStore` or `idempotencyStore` assignments for the allocate use case.
  - [ ] The payment status finder is either inlined at its single use site or otherwise reduced to one clear read-side dependency.
- Notes:
  - Keeping one explicit read-side dependency for the status API is acceptable.

## Non-functional requirements

- Performance (NFR-001):
  - The refactor must not add extra persistence round trips on the successful fresh-allocation path beyond what is needed to preserve behavior.
- Availability/Reliability (NFR-002):
  - Idempotent replay, duplicate-claim recovery, and successful issuance behavior must remain unchanged.
- Security/Privacy (NFR-003):
  - No new data exposure or secret handling is introduced.
- Compliance (NFR-004):
  - None.
- Observability (NFR-005):
  - Existing error mapping and logging behavior remain unchanged.
- Maintainability (NFR-006):
  - Use-case constructor dependencies should better reflect true responsibilities: `UnitOfWork` for write flow orchestration, dedicated read port only where actually needed.

## Dependencies and integrations

- External systems:
  - None.
- Internal services:
  - `internal/application/use_cases`
  - `internal/adapters/outbound/persistence/postgres`
  - `internal/infrastructure/di`
