---
doc: 00_problem
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

# Container Store Injection Refactor - Problem & Goals

## Context

- Background:
  - The API container currently instantiates DB-scoped Postgres stores that exist only to satisfy use-case constructor wiring.
  - `AllocatePaymentAddressUseCase` already owns a `UnitOfWork`, but it still injects allocation/idempotency stores for transaction-external replay lookups.
  - `GetPaymentAddressStatusUseCase` uses a dedicated read-side finder, but the current container also keeps a separate local assignment for it.
- Users or stakeholders:
  - Backend maintainers reviewing container wiring and use-case dependencies.
- Why now:
  - The user wants to remove unnecessary store injections visible in the container and simplify the dependency shape.

## Constraints (optional)

- Technical constraints:
  - Preserve current API behavior for payment-address creation, idempotent replay, and payment-status reads.
  - Keep clean architecture boundaries explicit; do not hide the issue behind more DI helper layers.
  - Do not add new persistence models or migrations.
- Timeline/cost constraints:
  - Prefer a targeted refactor over a broad persistence redesign.
- Compliance/security constraints:
  - None.

## Problem statement

- Current pain:
  - The container visibly constructs `allocationStore`, `paymentAddressStatusFinder`, and `idempotencyStore` even though at least part of that wiring exists only because of use-case constructor shape rather than true independent dependencies.
  - `AllocatePaymentAddressUseCase` currently mixes `UnitOfWork` with direct store injections, which makes ownership of persistence access less clear.
- Evidence or examples:
  - `AllocatePaymentAddressUseCase` only uses the injected allocation/idempotency stores for replay reads outside the transaction boundary.
  - The status finder is only used once by the status use case and does not need its own named variable in the container.

## Goals

- G1:
  - Remove direct allocation/idempotency store injection from `AllocatePaymentAddressUseCase`.
- G2:
  - Keep payment-address replay lookup behavior intact by routing those reads through `UnitOfWork`.
- G3:
  - Simplify container wiring so only meaningful dependencies remain as named assignments.

## Non-goals (out of scope)

- NG1:
  - Redesigning all persistence ports in the payment flow.
- NG2:
  - Changing API contracts or idempotency semantics.
- NG3:
  - Reworking the payment status read model beyond what is needed for cleaner container wiring.

## Assumptions

- A1:
  - Opening a transaction for replay lookup is acceptable for this refactor because correctness and dependency clarity matter more than preserving the old non-transactional fast path.
- A2:
  - Inlining the status finder construction at its single call site is sufficient; a bigger abstraction is unnecessary.

## Open questions

- Q1:
  - None for this scope.

## Success metrics

- Metric:
  - The API container no longer needs separate allocation/idempotency store assignments for the allocate use case.
- Target:
  - `AllocatePaymentAddressUseCase` constructor depends on `UnitOfWork` rather than transaction-external store injections, and payment behavior remains unchanged.
