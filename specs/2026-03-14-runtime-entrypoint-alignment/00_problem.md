---
doc: 00_problem
spec_date: 2026-03-14
slug: runtime-entrypoint-alignment
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-10-cloudflare-workers-postgres
  - 2026-03-11-cloudflare-poller-workers
  - 2026-03-12-api-worker-naming-unification
  - 2026-03-13-cloudflare-webhook-dispatcher-worker
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background:
  - The repo now has API HTTP runtimes, standalone scheduler runtimes, and Cloudflare worker
    runtimes that all enter the application through `cmd/`, `internal/bootstrap`, and inbound
    adapters.
- Users or stakeholders:
  - payrune maintainers who need the runtime entrypoint structure to be easy to reason about.
- Why now:
  - The runtime entrypoint cleanup happened incrementally, which left the spec history fragmented
    across multiple small folders even though the implementation is one coherent refactor.

## Constraints (optional)

- Technical constraints:
  - API routes, worker payload contracts, scheduler config semantics, and application behavior must
    remain unchanged.
  - The refactor must stay within existing `cmd/`, `internal/bootstrap`, `internal/adapters`, and
    `internal/infrastructure/di` boundaries.

## Problem statement

- Current pain:
  - Before the cleanup, inbound HTTP and scheduled worker responsibilities were mixed, standalone
    and Cloudflare scheduler runtimes duplicated one-cycle mapping, worker `cmd/*` packages knew too
    much about adapters/DI, and bootstrap naming drifted across `app`/`api` and
    `Dispatch`/`Dispatcher`.
- Evidence or examples:
  - Cloudflare API bridge and scheduler handlers previously lived together.
  - Standalone scheduler loops previously mapped request DTOs directly in bootstrap.
  - Worker commands previously imported inbound adapters and DI directly.
  - Bootstrap previously used generic `Run` in `app.go` and mixed webhook dispatcher nouns.

## Goals

- G1:
  - Make HTTP and scheduler inbound boundaries explicit and shared across runtimes.
- G2:
  - Keep `cmd/*` entrypoints uniformly thin and route runtime orchestration through
    `internal/bootstrap`.
- G3:
  - Normalize bootstrap naming around the runtime nouns `api`, `poller`, and
    `receipt webhook dispatcher`.

## Non-goals (out of scope)

- NG1:
  - Changing use-case behavior, worker JSON field names, route shapes, or scheduler cadence.
- NG2:
  - Introducing generic frameworks or abstraction layers beyond the concrete runtimes already in
    the repo.

## Assumptions

- A1:
  - The appropriate consistency target is thin `cmd/*` entrypoints, bootstrap-owned runtime
    orchestration, and transport-specific inbound adapters.
- A2:
  - One merged spec is a clearer source of truth than multiple same-day micro-specs for the same
    refactor.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - Runtime entrypoint responsibilities are split consistently across `cmd/`, `internal/bootstrap`,
    and inbound adapters.
- Target:
  - Shared HTTP handler assembly, shared scheduler cycle handlers, thin worker commands, and
    normalized bootstrap names are all present in the final code.
- Metric:
  - Spec tracking is consolidated.
- Target:
  - One `2026-03-14-runtime-entrypoint-alignment` spec folder replaces the fragmented same-day
    specs for this refactor.
