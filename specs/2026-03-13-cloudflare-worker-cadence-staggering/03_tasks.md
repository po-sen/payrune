---
doc: 03_tasks
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

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale:
  - This is a small Cloudflare deployment-default change with no runtime code, schema, or
    integration-shape change.
- Upstream dependencies (`depends_on`):
  - `2026-03-11-cloudflare-poller-workers`
  - `2026-03-13-cloudflare-webhook-dispatcher-worker`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - Only `wrangler.toml` schedule/default values and related docs are changing.
  - What would trigger switching to Full mode:
    - Any runtime, secret-model, or deploy-flow change.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Not skipped.

## Milestones

- M1:
  - Update staggered Cloudflare cron defaults.
- M2:
  - Update batch defaults and validate docs/config.

## Tasks (ordered)

1. T-001 - Stagger worker cron schedules

   - Scope:
     - Update poller and dispatcher cron expressions to 15-minute cadence with explicit ordering:
       `testnet4`, `mainnet`, `dispatcher`.
   - Output:
     - Updated `deployments/cloudflare/payrune-poller/wrangler.toml`
     - Updated `deployments/cloudflare/payrune-webhook-dispatcher/wrangler.toml`
   - Linked requirements: FR-001, FR-002, NFR-001, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): inspect `wrangler.toml` cron expressions and
           run `wrangler deploy --dry-run`.
     - [ ] Expected result: `testnet4` uses `*/15`, `mainnet` uses `5,20,35,50`, and
           `dispatcher` uses `10,25,40,55`.
     - [ ] Logs/metrics to check (if applicable): none

1. T-002 - Update worker batch defaults
   - Scope:
     - Change Cloudflare poller batch size to `10` and dispatcher dispatch batch size to `20`.
   - Output:
     - Updated `wrangler.toml` defaults and matching Cloudflare README/spec text.
   - Linked requirements: FR-003, FR-004, NFR-001
   - Validation:
     - [ ] How to verify (manual steps or command): inspect config values and run worker tests.
     - [ ] Expected result: poller uses batch `10` and dispatcher uses batch `20`.
     - [ ] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-001
- FR-003 -> T-002
- FR-004 -> T-002
- NFR-001 -> T-001, T-002
- NFR-002 -> T-001

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Restore previous `wrangler.toml` cron and batch values.
