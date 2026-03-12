---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - Cloudflare worker configuration defaults for API and poller.
- Not covered:
  - Runtime verification inside deployed Cloudflare environments.

## Tests

### Unit

- TC-001:

  - Linked requirements: FR-001, FR-002, NFR-002
  - Steps:
    - Parse `deployments/cloudflare/payrune-poller/wrangler.toml` with a TOML parser.
  - Expected:
    - Both environments use `*/5 * * * *` and `POLL_BATCH_SIZE = "5"`.

- TC-002:
  - Linked requirements: FR-003, FR-004, NFR-001, NFR-002
  - Steps:
    - Parse both `wrangler.toml` files with a TOML parser.
  - Expected:
    - Both workers explicitly enable observability/logging settings.

## Integration

- TC-101:
  - Linked requirements: FR-001, FR-002, FR-003, FR-004
  - Steps:
    - Run the existing worker test suites after the config updates.
  - Expected:
    - API and poller worker tests still pass.

## Edge cases and failure modes

- Case:
  - Cloudflare observability keys are misspelled or placed under the wrong section.
- Expected behavior:
  - TOML/config validation catches the mistake before deploy.

## NFR verification

- Observability:
  - Confirm both workers now have explicit observability/logging settings.
- Maintainability:
  - Confirm the diff stays limited to Cloudflare deployment config and directly related docs/specs.
