---
doc: 03_tasks
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

# Task Plan

## Mode decision

- Selected mode: Full
- Rationale:
  - This change introduces a new operator flow, a new Solidity contract artifact, and non-trivial
    failure/security considerations.
- Upstream dependencies (`depends_on`):
  - 2026-04-02-sweep-material-redesign
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - Not applicable.
  - What would trigger switching to Full mode:
    - Already Full.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Not applicable. `04_test_plan.md` is required in Full mode.

## Milestones

- M1:
  - Extend the CREATE2 factory contract and artifact for batch sweep recovery.
- M2:
  - Add the Ledger-only factory deploy script and simplify sweep to one script.
- M3:
  - Validate the full one-time initialization plus batch sweep workflow.

## Tasks (ordered)

1. T-001 - Extend the CREATE2 factory contract for batch sweep recovery
   - Scope:
     - Keep one canonical CREATE2-aware recovery entry point in `Create2ReceiverFactory`.
     - Let the factory derive each receiver from `salt + init_code`, deploy missing receivers, and
       then sweep them.
     - Remove public legacy recovery entry points from the factory API.
   - Output:
     - Updated factory contract source and artifact in `internal/infrastructure/ethereumcreate2assets`.
   - Linked requirements: FR-004 / FR-006 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command):
           `bash scripts/ethereum_create2_build_artifacts.sh`
     - [x] Expected result:
           Build completes and emits the updated factory artifact with no batch caller artifact.
     - [x] Logs/metrics to check (if applicable):
           No Solidity compile errors.
2. T-002 - Implement one-time Ledger-only factory deployment
   - Scope:
     - Add a dedicated deploy helper that deploys `Create2ReceiverFactory`, resolves the target
       network, validates the Ledger sender, and writes the deployed address back into metadata on
       successful broadcast.
   - Output:
     - `scripts/ethereum_create2_factory_deploy.sh`
   - Linked requirements: FR-002 / FR-005 / FR-006 / FR-007 / NFR-002 / NFR-003 / NFR-005 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command):
           Run the deploy script in dry-run mode with mocked Ledger and RPC helpers.
     - [x] Expected result:
           One deterministic deploy command is printed, and a broadcast rehearsal can update the
           expected metadata file.
     - [x] Logs/metrics to check (if applicable):
           Dry-run output includes network, metadata file, sender, and command.
3. T-003 - Simplify the Ledger-only sweep helper to one script
   - Scope:
     - Make `scripts/ethereum_create2_sweep.sh` support one or many explicit selections and route
       recovery through the active factory recorded in checked-in metadata for the selected network.
     - Add a live on-chain balance precheck so zero-balance receivers fail closed before broadcast.
     - Make the script encode CREATE2 recovery payload into one canonical factory batch call that
       also works when selected receivers are still undeployed.
     - Recompute `init_code_hash` and CREATE2 predicted address from the recovery payload, and for
       already-deployed receivers verify `collector()` matches the recorded collector before
       broadcast.
     - Remove legacy single-row recovery branching from the script.
   - Output:
     - One clear Ledger-only sweep helper for CREATE2 recovery.
   - Linked requirements: FR-001 / FR-002 / FR-003 / FR-005 / FR-006 / FR-007 / NFR-001 / NFR-002 / NFR-003 / NFR-005 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command):
           Run the sweep script in dry-run mode against fixture/test data with one and many
           explicit selections.
     - [x] Expected result:
           One deterministic factory-based CREATE2 recovery command is printed, with no batch-caller
           env, zero-balance receivers are rejected before broadcast, undeployed receivers are still
           recoverable through the same path, and malformed recovery payload or wrong deployed
           receiver contracts are rejected before broadcast.
     - [x] Logs/metrics to check (if applicable):
           Dry-run output includes selected ids, network, count, balances, receiver deployment
           states, sender, factory address, recovery path, and command.
4. T-004 - Remove non-Ledger and extra-contract helper paths
   - Scope:
     - Remove the private-key-based verify wrapper script and CLI subcommand, plus the separate
       batch-caller contract, artifact, env, and scripts.
   - Output:
     - One clear Ledger-only operator surface for CREATE2 deployment and sweep.
   - Linked requirements: FR-002 / FR-004 / FR-006 / NFR-003 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command):
           `rg -n "verify-chain|ETHEREUM_CREATE2_VERIFY_OPERATOR_PRIVATE_KEY|operator-private-key|ETHEREUM_SWEEP_BATCH_CALLER_ADDRESS|SweepBatchCaller|ethereum_create2_batch_sweep" cmd scripts README.md internal`
     - [x] Expected result:
           No legacy private-key or extra-contract operator helper path remains.
     - [x] Logs/metrics to check (if applicable):
           None.
5. T-005 - Document the one-time initialization and sweep workflow
   - Scope:
     - Update README and Ethereum CREATE2 asset docs to explain one-time factory deployment,
       metadata update, explicit selection, dry-run review, and broadcast flow.
   - Output:
   - Updated operator documentation.
   - Linked requirements: FR-002 / FR-005 / FR-006 / FR-007 / NFR-005 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command):
           Review README snippets and env list for completeness.
     - [x] Expected result:
           Operators can follow the one-contract Ledger-only flow without hidden context.
     - [x] Logs/metrics to check (if applicable):
           None.
6. T-006 - Regression validation
   - Scope:
     - Add or update tests for contract build wiring, metadata update behavior, and simplified sweep
       script validation.
   - Output:
     - Passing repo validation and targeted tests.
   - Linked requirements: FR-001 / FR-002 / FR-003 / FR-004 / FR-005 / FR-006 / FR-007 / NFR-001 / NFR-002 / NFR-003 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command):
           `go test ./...`
     - [x] Expected result:
           Tests pass after the batch sweep additions.
     - [x] Logs/metrics to check (if applicable):
           None.

## Validation evidence

- `bash scripts/ethereum_create2_build_artifacts.sh`
- `bash -n scripts/ethereum_create2_factory_deploy.sh scripts/ethereum_create2_sweep.sh`
- Valid deploy dry-run rehearsal with mocked Ledger sender
  - printed one `cast send --create ... --ledger` command and target metadata file
- Valid deploy broadcast rehearsal with stubbed `cast send --json`
  - updated the expected metadata file with the deployed factory address
- Valid sweep dry-run rehearsal with mocked `psql` and mocked Ledger sender
  - printed one active-factory batch recovery command for two receivers and exposed the chosen
    call signature in dry-run output
- Invalid sweep dry-run rehearsal with duplicate `ETHEREUM_SWEEP_PAYMENT_ADDRESS_IDS`
  - failed closed before any broadcast attempt
- Invalid sweep dry-run rehearsal with one zero-balance receiver
  - failed closed before any broadcast attempt
- Invalid sweep dry-run rehearsal with stale row payload targeting a superseded factory address
  - failed closed before any broadcast attempt
- Invalid sweep dry-run rehearsal with mismatched `init_code_hash`
  - failed closed before any broadcast attempt
- Invalid sweep dry-run rehearsal with mismatched computed CREATE2 address
  - failed closed before any broadcast attempt
- Invalid sweep dry-run rehearsal with deployed receiver whose `collector()` does not match payload
  - failed closed before any broadcast attempt
- Sweep broadcast rehearsal with stubbed `psql`, mocked Ledger sender, and stubbed `cast send`
  - verified `--rpc-url` / `--from` / `--ledger` precede the factory address and raw calldata so
    the command is directly executable with the current Foundry CLI
- `rg -n "verify-chain|ETHEREUM_CREATE2_VERIFY_OPERATOR_PRIVATE_KEY|operator-private-key|ETHEREUM_SWEEP_BATCH_CALLER_ADDRESS|SweepBatchCaller|ethereum_create2_batch_sweep" cmd scripts README.md internal`
- `go test ./...`
- `SPEC_DIR="specs/2026-04-03-ethereum-ledger-batch-sweep" bash scripts/spec-lint.sh`
- `bash scripts/precommit-run.sh`

## Traceability (optional)

- FR-001 -> T-003, T-006
- FR-002 -> T-002, T-003, T-004, T-005, T-006
- FR-003 -> T-003, T-006
- FR-004 -> T-001, T-004, T-006
- FR-005 -> T-002, T-003, T-005, T-006
- FR-006 -> T-004, T-005, T-006
- FR-007 -> T-002, T-003, T-005, T-006
- NFR-001 -> T-003, T-006
- NFR-002 -> T-002, T-003, T-006
- NFR-003 -> T-002, T-003, T-004, T-006
- NFR-005 -> T-002, T-003, T-005, T-006
- NFR-006 -> T-001, T-002, T-003, T-004, T-005, T-006

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - No DB migration.
  - Factory deployment is an operational prerequisite for real issuance metadata and batch
    broadcast.
- Rollback steps:
  - Revert to the previous checked-in factory metadata and run the same sweep helper with one
    selected row.
