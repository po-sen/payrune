---
doc: 04_test_plan
spec_date: 2026-03-09
slug: shared-tip-height-polling
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

# Shared Tip Height Polling - Test Plan

## Scope

- Covered:
  - Poll-cycle latest-block-height caching.
  - Multi-chain routing of separately fetched latest block height.
  - Bitcoin Esplora observer behavior when latest block height is supplied by the caller.
- Not covered:
  - Provider billing measurement outside unit and integration verification.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, FR-002, NFR-001, NFR-002
  - Steps:
    - Run receipt polling use-case tests covering successful updates, observer failures, and repeated addresses on the same network.
  - Expected:
    - The polling use case fetches latest block height once per chain/network pair and reuses it across address observations.
- TC-002:
  - Linked requirements: FR-002, NFR-005, NFR-006
  - Steps:
    - Run multi-chain observer and Bitcoin observer tests with the new shared tip-height input flow.
  - Expected:
    - Routing and confirmation calculations still behave correctly when latest block height is supplied externally.

### Integration

- TC-101:
  - Linked requirements: FR-002, NFR-002, NFR-005
  - Steps:
    - Run poller DI tests and package compilation checks.
  - Expected:
    - Poller container wiring remains valid after the observer port change.

## Edge cases and failure modes

- Case:
  - Latest block height fetch fails for one claimed chain/network pair.
  - Expected behavior:
    - Trackings on that pair are marked with polling errors using existing retry behavior, without repeated tip-height calls for that pair in the same cycle.

## NFR verification

- Performance:
  - Confirm one poll cycle does not repeatedly call latest-block-height fetch for multiple addresses on the same chain/network.
- Reliability:
  - Confirm status transitions and polling-error persistence remain unchanged.
- Security:
  - Confirm no new external data types or credentials are introduced.
