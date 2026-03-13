---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - Cloudflare cron/default config for poller and dispatcher workers.
- Not covered:
  - Live production scheduling behavior in Cloudflare after deploy.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-003, FR-004
  - Steps:
    - Review `wrangler.toml` values for poller and dispatcher defaults.
  - Expected:
    - Poller batch size is `10` and dispatcher batch size is `20`.

### Integration

- TC-101:
  - Linked requirements: FR-001, FR-002, FR-003, FR-004, NFR-001, NFR-002
  - Steps:
    - Run `wrangler deploy --dry-run` for poller and dispatcher workers.
  - Expected:
    - Worker configs parse correctly and show:
      - `testnet4 = */15`
      - `mainnet = 5,20,35,50`
      - `dispatcher = 10,25,40,55`
      - `poller batch = 10`
      - `dispatcher batch = 20`

## Edge cases and failure modes

- Case:
  - Cron expressions accidentally overlap.
- Expected behavior:
  - Review/dry-run should make the ordering obvious before deploy.

## NFR verification

- Maintainability:
  - Confirm the change stays limited to Cloudflare deployment configuration and docs/spec.
- Operability:
  - Confirm the final cron expressions clearly communicate `testnet4 -> mainnet -> dispatcher`.
