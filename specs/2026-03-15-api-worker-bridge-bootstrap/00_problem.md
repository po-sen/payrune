---
doc: 00_problem
spec_date: 2026-03-15
slug: api-worker-bridge-bootstrap
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-14-runtime-entrypoint-alignment
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
  - Runtime entrypoints were aligned so `cmd/*` delegates through `internal/bootstrap`, shared HTTP
    assembly lives under `internal/adapters/inbound/http`, and scheduler one-cycle mapping lives
    under `internal/adapters/inbound/scheduler`.
- Users or stakeholders:
  - payrune maintainers keeping inbound adapter boundaries strict and easy to explain.
- Why now:
  - The Cloudflare API request bridge currently lives under
    `internal/adapters/inbound/http/cloudflare`, but it bridges a Cloudflare worker envelope into a
    local `http.Handler` rather than adapting transport input directly to an application inbound
    port.

## Constraints (optional)

- Technical constraints:
  - Cloudflare API worker JSON request/response contracts must remain unchanged.
  - Public API handler behavior must remain unchanged.

## Problem statement

- Current pain:
  - `internal/adapters/inbound/http/cloudflare/bridge.go` sits in `inbound`, but its role is
    transport-to-transport bridging inside API worker runtime orchestration rather than a direct
    inbound adapter to application ports.
- Evidence or examples:
  - `internal/bootstrap/api_worker.go` already owns Cloudflare API payload decoding and DI
    orchestration, but still calls the separate bridge package for the final handler invocation.

## Goals

- G1:
  - Move the Cloudflare API request bridge logic into `internal/bootstrap`.
- G2:
  - Remove the now-misplaced `internal/adapters/inbound/http/cloudflare` package.

## Non-goals (out of scope)

- NG1:
  - Changing API routes, middleware, or response mapping behavior.
- NG2:
  - Reworking scheduler handlers or other runtime entrypoints.

## Assumptions

- A1:
  - The Cloudflare API request bridge is better modeled as runtime orchestration support in
    `bootstrap` than as an inbound adapter package.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - The API worker bridge no longer lives under `internal/adapters/inbound`.
- Target:
  - `internal/adapters/inbound/http/cloudflare` is removed, and equivalent logic exists in
    `internal/bootstrap`.
