---
doc: 04_test_plan
spec_date: 2026-03-06
slug: write-through-receipt-tracking
mode: Full
status: DONE
owners:
  - payrune-team
depends_on: []
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Test Plan

## Scope

- Covered:
  - Write-through registration contract, allocation use case transaction flow, poller cycle behavior after register-step removal.
- Not covered:
  - Live node integration behavior for blockchain observer.

## Tests

### Unit

- TC-001:

  - Linked requirements: FR-001, NFR-002
  - Steps:
    - Run allocation use case success path.
  - Expected:
    - `Complete` and receipt registration are both called once inside same transaction.

- TC-002:

  - Linked requirements: FR-002, FR-005, NFR-001, NFR-005
  - Steps:
    - Run poller cycle use case success/validation paths.
  - Expected:
    - No register call exists; `ClaimDue` still executes; output/log fields only include active counters.

- TC-003:
  - Linked requirements: FR-006, NFR-002
  - Steps:
    - Run allocation use case with network-specific required-confirmations config.
  - Expected:
    - Mainnet/testnet4 allocations persist different `required_confirmations` values as configured.

### Integration

- TC-101:

  - Linked requirements: FR-003
  - Steps:
    - Inspect migration SQL and run repository adapter tests.
  - Expected:
    - Backfill query is idempotent and scope-limited to issued allocations with usable chain/network/address values.

- TC-102:
  - Linked requirements: FR-006
  - Steps:
    - Validate DI env parsing for `BITCOIN_MAINNET_REQUIRED_CONFIRMATIONS` and `BITCOIN_TESTNET4_REQUIRED_CONFIRMATIONS`.
  - Expected:
    - Invalid/non-positive values fail fast; missing values fall back to `1`.

## Edge cases and failure modes

- Case:
  - Tracking repository missing in allocation UoW repos.
- Expected behavior:
  - Use case returns explicit configuration error and does not silently skip registration.

## NFR verification

- Performance:
  - Poller cycle path excludes allocation-scan registration query.
- Reliability:
  - Allocation issue + registration remains atomic via one transaction.
- Security:
  - No new secret/config surfaces introduced.
