---
doc: 00_problem
spec_date: 2026-03-11
slug: cloudflare-poller-workers
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-05-blockchain-receipt-polling-service
  - 2026-03-09-shared-tip-height-polling
  - 2026-03-09-poller-interval-separation
  - 2026-03-09-receipt-expire-final-check
  - 2026-03-10-cloudflare-workers-postgres
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
  - Payrune today runs `payrune-poller-mainnet` and `payrune-poller-testnet4` as separate long-running Go processes.
  - The public API has already been moved toward a standalone Cloudflare Worker runtime using Go/Wasm and a Worker-compatible PostgreSQL adapter.
  - Receipt polling still depends on process-local ticker loops and compose/runtime deployment.
- Users or stakeholders:
  - Payrune operators who want both Bitcoin pollers to run on Cloudflare instead of separate servers.
  - Backend teams who need polling behavior to stay aligned with the existing Go receipt lifecycle and polling use case.
- Why now:
  - The next Cloudflare slice is to move the two Bitcoin pollers off standalone processes while preserving current polling semantics and keeping future poller feature work in Go.

## Constraints (optional)

- Technical constraints:
  - The resulting pollers must run as standalone Cloudflare Workers with scheduled triggers.
  - Existing Go polling use case and receipt lifecycle remain the source of truth.
  - `deployments/cloudflare/` must stay a thin shell; future poller changes should primarily happen outside `deployments/`.
  - Plain Workers are preferred; no Hono.
  - PostgreSQL remains external.
- Timeline/cost constraints:
  - Reuse the existing receipt polling domain and application code rather than redesigning the product behavior.
  - Keep the diff tightly scoped to Cloudflare poller runtime code.
- Compliance/security constraints:
  - Security hardening beyond the minimum runnable Worker poller can be handled later.

## Problem statement

- Current pain:
  - `payrune-poller-mainnet` and `payrune-poller-testnet4` still require separate long-running process runtime and deployment shape.
  - A Cloudflare move that reimplements polling behavior in JS would drift from the current Go polling use case.
  - Polling also depends on Bitcoin Esplora observer behavior, so a Cloudflare runtime path needs both PostgreSQL and blockchain observation bridges.
- Evidence or examples:
  - `cmd/poller` currently runs `bootstrap.RunPoller`, which owns ticker-based process lifecycle.
  - `RunReceiptPollingCycleUseCase` already owns batch claim, shared tip-height reuse, final-check expiry ordering, and status persistence logic.

## Goals

- G1:
  - Run `payrune-poller-mainnet` and `payrune-poller-testnet4` directly on Cloudflare Workers with scheduled triggers.
- G2:
  - Execute the existing Go receipt polling use case instead of writing a parallel JS polling implementation.
- G3:
  - Keep `deployments/cloudflare/` as runtime shell only.
- G4:
  - Ensure future poller feature work usually changes Go code, not deployment-shell code.
- G5:
  - Preserve current polling semantics for scope, claim/reschedule, shared tip height, and final-check expiry.

## Non-goals (out of scope)

- NG1:
  - Moving receipt webhook dispatcher into Cloudflare in this slice.
- NG2:
  - Replacing PostgreSQL with D1 or SQLite.
- NG3:
  - Rewriting receipt polling business logic in standalone JS.
- NG4:
  - Changing receipt state semantics, payment expiry policy, or Esplora business rules.
- NG5:
  - Adding edge auth, abuse controls, or other security hardening in this slice.

## Assumptions

- A1:
  - Cloudflare Cron Trigger minimum granularity of one minute is acceptable because current receipt reschedule defaults are much larger than one minute.
- A2:
  - Operators can provide PostgreSQL and Esplora secrets to Cloudflare Workers.
- A3:
  - Two deployed Worker environments or worker names are acceptable as long as mainnet and testnet4 remain independently runnable.

## Open questions

- Q1:
  - None for this scope.

## Success metrics

- Metric:
  - Standalone runtime coverage.
- Target:
  - Both mainnet and testnet4 receipt pollers can be deployed and scheduled on Cloudflare without a separately running Go poller process.
- Metric:
  - Go logic reuse.
- Target:
  - Polling behavior runs through the existing Go polling use case and existing receipt lifecycle rules.
- Metric:
  - Deployment-shell stability.
- Target:
  - Future poller behavior changes generally do not require edits under `deployments/cloudflare/`.
