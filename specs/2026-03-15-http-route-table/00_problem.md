---
doc: 00_problem
spec_date: 2026-03-15
slug: http-route-table
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on: []
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background:
  - The inbound HTTP route composition file had been named
    `internal/adapters/inbound/http/handler.go`, while the chain-address controller acted like a
    nested router for `/v1/chains/...`.
- Users or stakeholders:
  - payrune maintainers who want the public API surface to be readable from one place.
- Why now:
  - The current split makes it hard to see what HTTP paths exist without reading
    controller-internal route parsing logic, and the current `handler` naming is weaker than a
    routing-specific name.

## Constraints (optional)

- Technical constraints:
  - Public API paths and response behavior should remain unchanged.
- Timeline/cost constraints:
  - None.
- Compliance/security constraints:
  - None.

## Problem statement

- Current pain:
  - The old `handler.go` file did not show the actual API route table, and
    `chain_address_controller.go` manually parsed `/v1/chains/...` path segments internally.
- Evidence or examples:
  - A reader had to jump between `handler.go` and `parseChainRoute(...)` to understand the HTTP
    surface area, and the file/function names still said `handler` even though the file owned
    routing composition.

## Goals

- G1:
  - Make the inbound HTTP route composition live in a single routing file.
- G2:
  - Remove controller-internal sub-routing for `/v1/chains/...`.
- G3:
  - Align file and function naming with router semantics rather than generic handler semantics.

## Non-goals (out of scope)

- NG1:
  - Changing public API path names or request/response contracts.
- NG2:
  - Introducing a new top-level router framework or third-party dependency.

## Assumptions

- A1:
  - `net/http.ServeMux` path patterns and path values are sufficient for this refactor.
- A2:
  - Controllers should keep request parsing and response mapping, but not own nested route tables.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - Public HTTP routes are listed in one place.
- Target:
  - `router.go` contains the concrete route registrations, controller path parsing helpers for
    `/v1/chains/...` are removed, exported routing names use `router` terminology, and tests remain
    green.
