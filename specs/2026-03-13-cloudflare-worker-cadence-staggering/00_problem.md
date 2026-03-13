---
doc: 00_problem
spec_date: 2026-03-13
slug: cloudflare-worker-cadence-staggering
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-11-cloudflare-poller-workers
  - 2026-03-13-cloudflare-webhook-dispatcher-worker
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
  - `payrune-poller-mainnet`, `payrune-poller-testnet4`, and
    `payrune-webhook-dispatcher` already run as standalone Cloudflare Workers.
- Users or stakeholders:
  - payrune maintainers operating the Cloudflare-based deployment.
- Why now:
  - Current worker cadence is too frequent and not intentionally staggered across runtimes.

## Constraints (optional)

- Technical constraints:
  - Changes must stay in Cloudflare deployment defaults and docs.
  - Runtime logic, use cases, and deploy scripts must remain unchanged.

## Problem statement

- Current pain:
  - Pollers and dispatcher run too often and without a clear staggered order, increasing overlap
    and noise in Cloudflare scheduling.
- Evidence or examples:
  - Operator requested a 15-minute cadence, strict order `testnet4 -> mainnet -> dispatcher`, and
    new batch defaults for poller and dispatcher.

## Goals

- G1:
  - Run all three Cloudflare scheduled workers every 15 minutes.
- G2:
  - Stagger schedules by runtime in the explicit order `testnet4`, then `mainnet`, then
    `dispatcher`.
- G3:
  - Increase Cloudflare poller batch default to `10` and dispatcher batch default to `20`.

## Non-goals (out of scope)

- NG1:
  - Changing worker runtime code or Cloudflare service topology.
- NG2:
  - Changing secrets, `.env.cloudflare`, or deployment automation.

## Assumptions

- A1:
  - Five-minute offsets between workers are enough to satisfy the requested staggered order while
    keeping a 15-minute cadence.
- A2:
  - Existing claim TTL and retry defaults remain acceptable.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - Worker schedules are staggered and repeat every 15 minutes.
- Target:
  - `testnet4` runs at minute `0` of each 15-minute window, `mainnet` runs five minutes later,
    and `dispatcher` runs five minutes after that.
- Metric:
  - Cloudflare batch defaults reflect the requested throughput increase.
- Target:
  - Poller uses `POLL_BATCH_SIZE = "10"` and dispatcher uses
    `RECEIPT_WEBHOOK_DISPATCH_BATCH_SIZE = "20"`.
