---
doc: 00_problem
spec_date: 2026-03-13
slug: cloudflare-worker-defaults-logging
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-10-cloudflare-workers-postgres
  - 2026-03-11-cloudflare-poller-workers
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
  - `payrune-api` and `payrune-poller` run as standalone Cloudflare Workers.
- Users or stakeholders:
  - payrune maintainers operating Cloudflare worker deployments.
- Why now:
  - Poller cadence and batch defaults need adjusting, and both workers should emit Cloudflare
    observability logs by default.

## Constraints (optional)

- Technical constraints:
  - Changes must stay inside the Cloudflare deployment configuration.
  - Poller business semantics must remain unchanged.

## Problem statement

- Current pain:
  - Pollers currently run every minute with `POLL_BATCH_SIZE = 1`, and worker observability settings
    are not explicitly enabled in `wrangler.toml`.
- Evidence or examples:
  - Operator requested a slower five-minute cron cadence, larger batch size, and enabled logs for
    both API and poller workers.

## Goals

- G1:
  - Change Cloudflare poller trigger cadence to every five minutes.
- G2:
  - Change Cloudflare poller batch default to `5`.
- G3:
  - Enable Cloudflare worker logging/observability for both API and poller runtimes.

## Non-goals (out of scope)

- NG1:
  - Any change to worker runtime code paths or business logic.
- NG2:
  - Any Cloudflare security, routing, or secret-management redesign.

## Assumptions

- A1:
  - Enabling Cloudflare observability logs in `wrangler.toml` is the desired meaning of "enable
    log".
- A2:
  - Existing deploy flows continue to read the same `wrangler.toml` files without script changes.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - Poller deployment config reflects the requested cadence and batch defaults.
- Target:
  - `crons = ["*/5 * * * *"]` and `POLL_BATCH_SIZE = "5"` for both poller environments.
- Metric:
  - Worker observability logs are enabled in both worker configs.
- Target:
  - API and poller `wrangler.toml` include enabled observability/logging settings.
