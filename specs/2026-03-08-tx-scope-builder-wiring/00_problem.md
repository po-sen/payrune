---
doc: 00_problem
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

# Tx Scope Builder Wiring - Problem & Goals

## Context

- Background:
  - Postgres `UnitOfWork` currently depends on injected tx-scope factory wiring even though the set of tx-scoped Postgres stores is fixed inside this adapter.
  - The user wants `New*Container` constructors to stay as simple assigns without function-valued or factory-style tx-scope wiring.
- Users or stakeholders:
  - Backend maintainers reading transaction wiring.
  - Reviewers trying to keep composition decisions in the composition root.
- Why now:
  - The user explicitly wants the container code cleaned up so tx-scope wiring disappears from the constructor body.

## Constraints (optional)

- Technical constraints:
  - Keep transaction semantics unchanged.
  - Do not pre-create tx-bound store instances before a `*sql.Tx` exists.
  - Preserve clean architecture boundaries and current store types.
- Timeline/cost constraints:
  - Prefer a small behavior-preserving refactor.
- Compliance/security constraints:
  - None.

## Problem statement

- Current pain:
  - For a fixed Postgres adapter, injecting a tx-scope builder/factory adds ceremony without adding meaningful flexibility.
  - The constructor code becomes cluttered by plumbing that is not actually configurable in practice.
- Evidence or examples:
  - The current tx-scope contents are always the same four Postgres stores.
  - No production path provides an alternative tx-scope composition.

## Goals

- G1:
  - Remove tx-scope builder/factory wiring from container constructors.
- G2:
  - Keep container constructors to simple assign-style wiring.
- G3:
  - Preserve current callback-visible `TxScope` contents and commit/rollback behavior.

## Non-goals (out of scope)

- NG1:
  - Changing the shape of `TxScope`.
- NG2:
  - Replacing `UnitOfWork` with a different transaction abstraction.
- NG3:
  - Refactoring every adapter/store constructor pattern in this change.

## Assumptions

- A1:
  - For this adapter, letting `UnitOfWork` assemble its fixed Postgres tx scope is simpler than injecting a configurable builder.
- A2:
  - Tests that currently pass a builder directly to `NewUnitOfWork` can do the same after the helper removal.

## Open questions

- Q1:
  - None for this scope.

## Success metrics

- Metric:
  - Tx-scope construction is visible in the composition root without changing runtime behavior.
- Target:
  - Container constructors use `NewUnitOfWork(db)` directly, and transaction behavior remains unchanged.
