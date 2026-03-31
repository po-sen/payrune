---
doc: 04_test_plan
spec_date: 2026-03-31
slug: allocation-recovery-schema-phase1
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-29-allocation-issuance-naming
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Allocation Recovery Schema Phase 1 - Test Plan

## Scope

- Covered:
  - BTC and ETH issued-address derivation of `sweep_material_json`.
  - Allocation store read/write handling for the new column.
  - Allocation use case regression for both chains.
  - ETH sweep helper env validation, DB selection, dry-run command assembly, and Ledger sender
    consistency checks.
  - Spec lint and repo validation.
- Not covered:
  - Phase-2 internal schema cleanup.
  - Live mainnet / sepolia broadcast from CI.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-002 / NFR-003
  - Steps:
    - Run Bitcoin issued-address deriver tests with assertions on emitted `sweep_material_json`.
  - Expected:
    - New Bitcoin issued rows produce a JSON payload containing xpub, absolute HD path, scheme, and
      address.
- TC-002:
  - Linked requirements: FR-002 / NFR-003
  - Steps:
    - Run Ethereum issued-address deriver tests with assertions on emitted `sweep_material_json`.
  - Expected:
    - New Ethereum CREATE2 issued rows produce a JSON payload containing predicted address, factory,
      collector, salt, init code hex, and init code hash.
- TC-003:
  - Linked requirements: FR-002 / NFR-001
  - Steps:
    - Run allocation entity / use-case tests that call `MarkIssued(...)` and complete the
      allocation.
  - Expected:
    - New issued allocations persist non-empty `sweep_material_json` without changing reservation
      semantics.

### Integration

- TC-101:
  - Linked requirements: FR-001, FR-002 / NFR-001, NFR-003
  - Steps:
    - Run Postgres and Cloudflare Postgres allocation store tests against the updated column set.
  - Expected:
    - Stores write `sweep_material_json` on completion, clear it on reopen, and read it back for
      issued allocations.
- TC-102:
  - Linked requirements: FR-003 / NFR-001, NFR-003
  - Steps:
    - Validate the backfill migration logic in focused persistence coverage.
  - Expected:
    - Existing issued BTC and ETH rows receive `sweep_material_json` without touching cursor /
      slot continuity fields.
- TC-103:
  - Linked requirements: FR-006 / NFR-001
  - Steps:
    - Run focused allocation-path tests for both Bitcoin and Ethereum controllers / use cases.
  - Expected:
    - Both `POST /v1/chains/bitcoin/payment-addresses` and
      `POST /v1/chains/ethereum/payment-addresses` flows still succeed in test coverage.
- TC-104:
  - Linked requirements: FR-004 / NFR-002, NFR-004
  - Steps:
    - Run the ETH sweep helper in dry-run mode with stubbed `psql`, `jq`, and `cast` responses.
  - Expected:
    - The helper reads `sweep_material_json`, validates the Ledger sender, and prints the expected
      `cast send ... "sweep()"` command without requiring factory / collector / salt envs.

### E2E (if applicable)

- Scenario 1:
  - Apply migrations on a local DB, allocate one new ETH address, then run the helper dry-run using
    only DB selector + env.
- Scenario 2:
  - Apply migrations on a local DB containing older issued rows, backfill them, and confirm the
    operator can read `sweep_material_json` directly from SQL.

## Edge cases and failure modes

- Case:
  - Selected allocation row is not Ethereum or not `scheme='create2'`.
- Expected behavior:
  - Helper aborts with a clear operator error and does not attempt broadcast.
- Case:
  - `sweep_material_json` is missing or malformed on the selected row.
- Expected behavior:
  - Helper aborts and does not infer factory / collector / salt from internal compatibility fields.
- Case:
  - Ledger-derived sender does not match `ETHEREUM_SWEEP_FROM_ADDRESS`.
- Expected behavior:
  - Helper aborts before broadcast.
- Case:
  - Reopened derivation-failed reservation is retried.
- Expected behavior:
  - Stale `sweep_material_json` is cleared before the row is reissued.

## NFR verification

- Performance:
  - No new performance target in phase 1; allocation-path test runtime should remain within normal
    focused-suite bounds.
- Reliability:
  - Focused tests confirm no allocation/cursor continuity changes and no new allocation-path
    failures.
- Security:
  - Helper validation covers missing env, malformed JSON, and Ledger sender mismatch before
    broadcast.
