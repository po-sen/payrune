---
doc: 01_requirements
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

# Requirements

## Out-of-scope behaviors

- OOS1:
  - Changing worker code, Go/Wasm handlers, or use-case behavior.
- OOS2:
  - Changing deploy/delete scripts or `.env.cloudflare` contract.

## Functional requirements

### FR-001 - Staggered 15-minute poller schedules

- Description:
  - Both poller workers must run every 15 minutes and be staggered so `testnet4` runs before
    `mainnet`.
- Acceptance criteria:
  - [ ] `env.testnet4.triggers.crons = ["*/15 * * * *"]`.
  - [ ] `env.mainnet.triggers.crons = ["5,20,35,50 * * * *"]`.
  - [ ] The `testnet4` schedule fires before the `mainnet` schedule in each 15-minute sequence.
- Notes:
  - Use explicit cron expressions rather than runtime offsets.

### FR-002 - Staggered 15-minute dispatcher schedule

- Description:
  - The webhook dispatcher must also run every 15 minutes and fire after `mainnet`.
- Acceptance criteria:
  - [ ] `triggers.crons = ["10,25,40,55 * * * *"]` in dispatcher `wrangler.toml`.
  - [ ] The dispatcher schedule fires after the `mainnet` poller schedule.
- Notes:
  - The required runtime order is `testnet4`, then `mainnet`, then `dispatcher`.

### FR-003 - Poller batch default of ten

- Description:
  - Cloudflare poller workers must default `POLL_BATCH_SIZE` to `10`.
- Acceptance criteria:
  - [ ] `env.testnet4.vars.POLL_BATCH_SIZE = "10"`.
  - [ ] `env.mainnet.vars.POLL_BATCH_SIZE = "10"`.
- Notes:
  - Runtime fallback code does not need to change.

### FR-004 - Dispatcher batch default of twenty

- Description:
  - Cloudflare webhook dispatcher must default `RECEIPT_WEBHOOK_DISPATCH_BATCH_SIZE` to `20`.
- Acceptance criteria:
  - [ ] Dispatcher `wrangler.toml` sets `RECEIPT_WEBHOOK_DISPATCH_BATCH_SIZE = "20"`.
- Notes:
  - Other dispatcher defaults remain unchanged.

## Non-functional requirements

- Maintainability (NFR-001):
  - Keep the change limited to Cloudflare deployment config and matching documentation/specs.
- Operability (NFR-002):
  - The final cron expressions must make the runtime order obvious during review.

## Dependencies and integrations

- External systems:
  - Cloudflare Workers / Wrangler configuration.
- Internal services:
  - Existing Cloudflare poller and webhook dispatcher runtimes.
