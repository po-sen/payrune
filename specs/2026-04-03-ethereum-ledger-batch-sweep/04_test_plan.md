---
doc: 04_test_plan
spec_date: 2026-04-03
slug: ethereum-ledger-batch-sweep
mode: Full
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-02-sweep-material-redesign
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
  - Solidity contract build wiring
  - Factory deploy script validation behavior
  - Unified sweep helper validation behavior
  - Removal of private-key-based CREATE2 helper paths
  - Documentation and regression checks for the new flow
- Not covered:
  - Real network mainnet/sepolia deployment in this spec

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-004 / FR-006 / NFR-006
  - Steps:
    - Run `bash scripts/ethereum_create2_build_artifacts.sh`
  - Expected:
    - The updated factory contract compiles, exposes CREATE2-aware batch recovery, and no separate
      batch sweeper artifact is emitted.
- TC-002:
  - Linked requirements: FR-007 / FR-005 / NFR-002 / NFR-005
  - Steps:
    - Run the factory deploy script in dry-run and mocked broadcast modes.
  - Expected:
    - Dry-run prints one deterministic deploy command and broadcast updates the expected metadata
      file with the deployed factory address.
- TC-003:
  - Linked requirements: FR-003 / FR-005 / FR-007 / NFR-002 / NFR-005
  - Steps:
    - Run targeted script tests or fixture-based validation for malformed selectors, mixed networks,
      missing `sweep_material_json`, stale payload factory mismatch against active metadata,
      mismatched `init_code_hash`, mismatched computed CREATE2 address, wrong deployed receiver
      collector, and zero-balance receivers.
  - Expected:
    - The script fails closed before broadcast for each invalid case.

### Integration

- TC-101:
  - Linked requirements: FR-001 / FR-002 / FR-003 / FR-005 / FR-007 / NFR-001 / NFR-003
  - Steps:
    - Run the unified sweep helper in dry-run mode with multiple valid issued rows that all target
      the same factory and a mocked/stubbed Ledger address command.
  - Expected:
    - One deterministic factory-based CREATE2 recovery command is printed and no broadcast occurs.
- TC-102:
  - Linked requirements: FR-001 / FR-002 / FR-005 / FR-006 / NFR-003 / NFR-005
  - Steps:
    - Run the unified sweep helper after the refactor with exactly one selected row.
  - Expected:
    - The same helper works for a batch of size 1.
- TC-103:
  - Linked requirements: FR-006 / NFR-003
  - Steps:
    - Search the repo for `verify-chain`, `operator-private-key`,
      `ETHEREUM_CREATE2_VERIFY_OPERATOR_PRIVATE_KEY`, `ETHEREUM_SWEEP_BATCH_CALLER_ADDRESS`,
      `SweepBatchCaller`, and `ethereum_create2_batch_sweep`.
  - Expected:
    - Legacy private-key and extra-contract operator helper paths are gone.
- TC-104:
  - Linked requirements: FR-001 / FR-002 / NFR-002 / NFR-005
  - Steps:
    - Run the batch helper in broadcast mode with stubbed `psql`, `cast wallet address --ledger`,
      and `cast send`.
  - Expected:
    - The script invokes `cast send` with flags before the batch call positionals, so the command
      is accepted by the current Foundry CLI.

### E2E (if applicable)

- Scenario 1:
  - Optional operator rehearsal on a test network using Ledger, deployed CREATE2 factory metadata,
    and two funded receivers.

## Edge cases and failure modes

- Case:
  - Duplicate receiver addresses are selected.
- Expected behavior:
  - The script rejects the invocation before rendering the command.
- Case:
  - The selected rows span more than one network.
- Expected behavior:
  - The script rejects the invocation.
- Case:
  - The active metadata factory has no code on-chain.
- Expected behavior:
  - The script rejects the invocation before broadcast.
- Case:
  - The recorded recovery payload points at a different factory than the active metadata factory.
- Expected behavior:
  - The script rejects the invocation before broadcast.
- Case:
  - The recorded `init_code_hash` does not equal `keccak(init_code_hex)`.
- Expected behavior:
  - The script rejects the invocation before broadcast.
- Case:
  - The recorded receiver address does not equal the CREATE2 address computed from the recorded
    factory, salt, and init code.
- Expected behavior:
  - The script rejects the invocation before broadcast.
- Case:
  - A receiver is already deployed, but its `collector()` does not match the recorded
    `collector_address`.
- Expected behavior:
  - The script rejects the invocation before broadcast.
- Case:
  - One selected receiver already has zero on-chain balance.
- Expected behavior:
  - The script rejects the invocation before broadcast.
- Case:
  - One receiver call fails on-chain.
- Expected behavior:
  - The whole batch transaction reverts.
- Case:
  - A selected CREATE2 address still has ETH balance but no deployed receiver code.
- Expected behavior:
  - The same batch recovery path still works because the factory deploys the missing receiver
    before calling `sweep()`.

## NFR verification

- Performance:
  - Dry-run remains fast and deterministic for a moderate batch size such as 25 receivers.
- Reliability:
  - All validation happens before Ledger broadcast.
- Security:
  - No private-key path or second singleton contract exists anywhere in the CREATE2 operator
    workflow after the change.

## Execution evidence

- `TC-001`
  - Executed via `bash scripts/ethereum_create2_build_artifacts.sh`
  - Result: passed
- `TC-002`
  - Executed via mocked factory deploy dry-run and broadcast rehearsal
  - Result: passed
- `TC-003`
  - Executed via duplicate-selector, stale-factory mismatch, `init_code_hash` mismatch,
    computed-address mismatch, wrong deployed collector, and zero-balance dry-run rehearsals
  - Result: passed
- `TC-101`
  - Executed via mocked `psql` plus mocked Ledger sender dry-run rehearsal with two receivers
  - Result: passed
- `TC-102`
  - Executed via mocked one-row sweep rehearsal
  - Result: passed
- `TC-103`
  - Executed via `rg -n "verify-chain|ETHEREUM_CREATE2_VERIFY_OPERATOR_PRIVATE_KEY|operator-private-key|ETHEREUM_SWEEP_BATCH_CALLER_ADDRESS|SweepBatchCaller|ethereum_create2_batch_sweep" cmd scripts README.md internal`
  - Result: passed
- `TC-104`
  - Executed via mocked broadcast rehearsal with stubbed `psql` and `cast send`
  - Result: passed
