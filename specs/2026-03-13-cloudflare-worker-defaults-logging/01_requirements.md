---
doc: 01_requirements
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

# Requirements

## Out-of-scope behaviors

- OOS1:
  - Changing worker source code, handlers, bridges, or use-case behavior.
- OOS2:
  - Changing deploy/delete scripts or `.env.cloudflare` contract.

## Functional requirements

### FR-001 - Five-minute poller cadence

- Description:
  - Cloudflare poller workers must run on a five-minute cron schedule instead of every minute.
- Acceptance criteria:
  - [ ] `env.mainnet.triggers.crons` is set to `["*/5 * * * *"]`.
  - [ ] `env.testnet4.triggers.crons` is set to `["*/5 * * * *"]`.
- Notes:
  - Only Cloudflare worker deployment defaults are changed.

### FR-002 - Poller batch default of five

- Description:
  - Cloudflare poller workers must default `POLL_BATCH_SIZE` to `5`.
- Acceptance criteria:
  - [ ] `env.mainnet.vars.POLL_BATCH_SIZE` is `"5"`.
  - [ ] `env.testnet4.vars.POLL_BATCH_SIZE` is `"5"`.
- Notes:
  - Runtime fallback code does not need to change.

### FR-003 - Enable API worker observability logs

- Description:
  - `payrune-api` Cloudflare worker configuration must explicitly enable Cloudflare observability
    logs.
- Acceptance criteria:
  - [ ] API `wrangler.toml` enables observability.
  - [ ] API `wrangler.toml` enables logs/invocation logs.
- Notes:
  - Use Cloudflare-supported `wrangler.toml` fields only.

### FR-004 - Enable poller worker observability logs

- Description:
  - `payrune-poller` Cloudflare worker configuration must explicitly enable Cloudflare
    observability logs.
- Acceptance criteria:
  - [ ] Poller `wrangler.toml` enables observability.
  - [ ] Poller `wrangler.toml` enables logs/invocation logs.
- Notes:
  - Apply at the shared worker config so both envs inherit it.

## Non-functional requirements

- Observability (NFR-001):
  - Use official Cloudflare Wrangler observability/logging configuration keys supported by current
    docs.
- Maintainability (NFR-002):
  - Keep changes limited to Cloudflare deployment configuration and related docs/specs.

## Dependencies and integrations

- External systems:
  - Cloudflare Workers / Wrangler configuration.
- Internal services:
  - Existing Cloudflare API and poller deploy flows.
