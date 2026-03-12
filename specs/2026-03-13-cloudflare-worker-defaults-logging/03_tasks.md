---
doc: 03_tasks
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

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale:
  - This is a small Cloudflare deployment-default change with no schema, integration-shape, or
    business-logic change.
- Upstream dependencies (`depends_on`):
  - `2026-03-10-cloudflare-workers-postgres`
  - `2026-03-11-cloudflare-poller-workers`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - Only `wrangler.toml` defaults and associated docs are changing.
  - What would trigger switching to Full mode:
    - Any change to runtime code, secret model, or deploy flow.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Not skipped.

## Milestones

- M1:
  - Update Cloudflare worker defaults.
- M2:
  - Validate configuration syntax and document the new defaults.

## Tasks (ordered)

1. T-001 - Update poller cadence and batch defaults

   - Scope:
     - Change Cloudflare poller cron trigger to every five minutes and set `POLL_BATCH_SIZE` to
       `5`.
   - Output:
     - Updated `deployments/cloudflare/payrune-poller/wrangler.toml`.
   - Linked requirements: FR-001, FR-002, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): parse `wrangler.toml` and inspect env values.
     - [ ] Expected result: both poller envs show `*/5` cron and `POLL_BATCH_SIZE = "5"`.
     - [ ] Logs/metrics to check (if applicable): none

2. T-002 - Enable worker observability logs
   - Scope:
     - Enable Cloudflare observability/logging in API and poller worker configs.
   - Output:
     - Updated `deployments/cloudflare/payrune-api/wrangler.toml`
     - Updated `deployments/cloudflare/payrune-poller/wrangler.toml`
   - Linked requirements: FR-003, FR-004, NFR-001, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): parse both `wrangler.toml` files and inspect
           observability keys.
     - [ ] Expected result: both workers explicitly enable observability/logs/invocation logs.
     - [ ] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-001
- FR-003 -> T-002
- FR-004 -> T-002
- NFR-001 -> T-002
- NFR-002 -> T-001, T-002

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Restore previous `wrangler.toml` values for cadence, batch size, and observability settings.
