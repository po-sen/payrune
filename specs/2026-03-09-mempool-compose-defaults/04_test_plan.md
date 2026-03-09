---
doc: 04_test_plan
spec_date: 2026-03-09
slug: mempool-compose-defaults
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

# Mempool Compose Defaults - Test Plan

## Scope

- Covered:
  - Compose fallback endpoint consistency.
  - Poller DI config-loading tests.
- Not covered:
  - Live endpoint health checks against public providers.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-002, NFR-002, NFR-006
  - Steps:
    - Run poller container config-loading tests after updating the mainnet example endpoint.
  - Expected:
    - Tests pass and reflect mempool.space as the mainnet example endpoint.

### Integration

- TC-101:
  - Linked requirements: FR-001, FR-002, NFR-002
  - Steps:
    - Run DI tests and package compilation checks.
  - Expected:
    - Updated compose defaults and config tests compile and pass cleanly.

## Edge cases and failure modes

- Case:
  - A deployment explicitly overrides `BITCOIN_MAINNET_ESPLORA_URL`.
  - Expected behavior:
    - The explicit env override still wins; only the default fallback changes.

## NFR verification

- Performance:
  - No runtime performance impact.
- Reliability:
  - Env parsing behavior remains unchanged aside from the fallback URL string.
- Security:
  - No new secrets or auth flows are introduced.
