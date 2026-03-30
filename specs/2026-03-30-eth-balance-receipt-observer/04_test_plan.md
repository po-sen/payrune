---
doc: 04_test_plan
spec_date: 2026-03-30
slug: eth-balance-receipt-observer
mode: Full
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-20-create2-eth-payment-receiving
  - 2026-03-30-eth-poller-stall-fix
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
  - ETH balance-snapshot observer behavior
  - ETH observer failure handling
  - Allocation/bootstrap non-regression
- Not covered:
  - Live-network deploy-and-sweep collection
  - ERC-20 receipt observation

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001 / FR-003 / NFR-001
  - Steps:
    - Run ETH observer against a test RPC server with exact balances for latest and confirmed
      block heights.
  - Expected:
    - Observer returns observed, confirmed, and unconfirmed totals from the two balance snapshots.
- TC-002:
  - Linked requirements: FR-003 / NFR-002
  - Steps:
    - Run ETH observer where confirmed snapshot exceeds the observed snapshot.
  - Expected:
    - Observer returns failure.
- TC-003:
  - Linked requirements: FR-001 / NFR-001
  - Steps:
    - Run ETH observer with `requiredConfirmations` greater than the latest block height.
  - Expected:
    - Observer returns `confirmed_total_minor = 0` and leaves all observed value in unconfirmed.

### Integration

- TC-101:
  - Linked requirements: FR-002 / FR-004 / NFR-006
  - Steps:
    - Run allocation and bootstrap tests after removing ETH baseline coupling.
  - Expected:
    - Allocation/bootstrap behavior remains unchanged.
- TC-102:
  - Linked requirements: NFR-003 / NFR-005 / NFR-006
  - Steps:
    - Run
      `SPEC_DIR="specs/2026-03-30-eth-balance-receipt-observer" bash scripts/spec-lint.sh`
      and `bash scripts/precommit-run.sh`.
  - Expected:
    - Spec lint and repo validation pass.

### E2E

- Scenario 1:
  - Optional follow-up local sepolia check after deployment to confirm ETH rows progress without
    block rescans.

## Edge cases and failure modes

- Case:
  - Provider returns inconsistent balance snapshots.
- Expected behavior:
  - Observer fails the row.
- Case:
  - Required confirmations exceed current chain height.
- Expected behavior:
  - Confirmed total stays zero.

## NFR verification

- Performance:
  - Confirm ETH observer uses at most two balance reads per row.
- Reliability:
  - Confirm provider and inconsistency failures remain row-scoped.
- Security:
  - Review that no new public fields, persistence fields, or sensitive logs were added.
