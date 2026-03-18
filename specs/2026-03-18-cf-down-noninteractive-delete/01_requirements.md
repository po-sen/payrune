---
doc: 01_requirements
spec_date: 2026-03-18
slug: cf-down-noninteractive-delete
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

# Requirements

## Out-of-scope behaviors

- OOS1:
  - Any change to Cloudflare worker deploy scripts, migrate scripts, or `Makefile` target names.
- OOS2:
  - Any change to worker source code or Wrangler configuration files.

## Functional requirements

### FR-001 - Non-interactive cf-down delete flow

- Description:
  - Every worker delete script used by `make cf-down` must call Wrangler delete in a way that does
    not ask the operator for an extra confirmation prompt.
- Acceptance criteria:
  - [ ] `scripts/cf-api-worker-delete.sh` invokes `wrangler delete --force`.
  - [ ] `scripts/cf-poller-worker-delete.sh` invokes `wrangler delete --force`.
  - [ ] `scripts/cf-receipt-webhook-mock-worker-delete.sh` invokes `wrangler delete --force`.
  - [ ] `scripts/cf-webhook-dispatcher-worker-delete.sh` invokes `wrangler delete --force`.
- Notes:
  - Use Wrangler-supported CLI flags rather than piping `yes` into stdin.

### FR-002 - Preserve current teardown entrypoints and arguments

- Description:
  - The change must preserve the current `make cf-down` entrypoint, script names, target-specific
    arguments, and delete order.
- Acceptance criteria:
  - [ ] `Makefile` continues to call the same delete scripts in the same order.
  - [ ] `scripts/cf-poller-worker-delete.sh` still passes the requested `mainnet|testnet4`
        environment to Wrangler.
- Notes:
  - Operators should not need to learn a new teardown command.

## Non-functional requirements

- Reliability (NFR-001):
  - The non-interactive change must use Wrangler's supported delete flag so the flow does not depend
    on shell TTY behavior or stdin piping.
- Maintainability (NFR-002):
  - Keep the code change limited to the delete scripts used by `make cf-down` plus directly related
    spec files.

## Dependencies and integrations

- External systems:
  - Cloudflare Wrangler CLI.
- Internal services:
  - `Makefile` `cf-down` target and the existing delete scripts under `scripts/`.
