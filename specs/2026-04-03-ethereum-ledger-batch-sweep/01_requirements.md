---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- CREATE2 factory contract:
  - The deployed `Create2ReceiverFactory` singleton that owns CREATE2 deployment metadata for a
    network and now also exposes batch sweep recovery calls.
- Explicit selector:
  - An operator-provided list of payment address IDs or Ethereum addresses. The batch flow must not
    discover targets implicitly.

## Out-of-scope behaviors

- OOS1:
  - Hot-wallet broadcast modes, unattended cron sweepers, or private-key environment variables.
- OOS2:
  - Multi-network batches in one invocation.
- OOS3:
  - Changing DB schema or `sweep_material_json` payload format.
- OOS4:
  - Supporting batch or single recovery through superseded local-development factories after a new
    active factory is deployed for the same network.

## Functional requirements

### FR-001 - Add a Ledger-only Ethereum CREATE2 batch sweep flow

- Description:
  - The repo must provide an operator flow that builds one batch sweep transaction for multiple
    Ethereum CREATE2 receiver addresses and signs it with Ledger.
- Acceptance criteria:
  - [ ] Operators can select multiple targets in one invocation using explicit IDs or explicit
        addresses.
  - [ ] The flow emits or sends exactly one batch transaction for the selected set.
  - [ ] The same batch transaction works for both already-deployed receivers and predicted CREATE2
        addresses that still have no receiver code on-chain.
  - [ ] The flow continues to support dry-run and explicit `--broadcast`.
  - [ ] The rendered and executed `cast send` command works with the repo's current Foundry CLI
        without requiring the operator to reorder arguments manually.
  - [ ] The contract and the script expose one canonical CREATE2 batch recovery path only; there is
        no legacy single-row recovery entry point.
- Notes:
  - One selected row is just a batch of size 1; a separate contract or address should not be
    required.

### FR-002 - Keep signer security strictly Ledger-only

- Description:
  - The new batch flow must not introduce any alternative signer path.
- Acceptance criteria:
  - [ ] The batch flow validates the connected Ledger sender against `ETHEREUM_SWEEP_FROM_ADDRESS`.
  - [ ] No new private-key environment variables, hot-wallet flags, or non-Ledger broadcast options
        are introduced.
  - [ ] Documentation explicitly states that Ledger interactive signing remains mandatory.
- Notes:
  - This is the primary security constraint for the feature.

### FR-003 - Validate batch inputs conservatively

- Description:
  - The batch flow must fail closed when selected rows or operator inputs are inconsistent.
- Acceptance criteria:
  - [ ] The script rejects empty selections, duplicate matches, invalid addresses, mixed networks,
        and non-Ethereum/non-create2 rows.
  - [ ] The script rejects rows with empty or invalid `sweep_material_json`.
  - [ ] The script rejects any selected receiver whose current on-chain ETH balance is zero.
  - [ ] The script recomputes `keccak(init_code_hex)` and rejects rows whose recorded
        `init_code_hash` does not match.
  - [ ] The script recomputes the CREATE2 predicted address from `factory_address`,
        `create2_salt`, and `init_code_hash`, and rejects rows whose recorded receiver address does
        not match.
  - [ ] If a selected receiver already has deployed code, the script verifies the deployed
        contract's `collector()` equals the recorded `collector_address` and rejects mismatches.
  - [ ] The script resolves the active factory from checked-in metadata for the selected network and
        rejects rows whose recorded `factory_address` does not equal that active factory.
  - [ ] The script rejects a selected network whose active factory has no code on-chain.
- Notes:
  - Fail closed is preferred over partial best-effort cleanup.

### FR-004 - Extend the CREATE2 factory contract for recovery

- Description:
  - The repo must use the deployed `Create2ReceiverFactory` singleton as the batch recovery entry
    point, with one canonical CREATE2-aware sweep function and no legacy recovery entry points.
- Acceptance criteria:
  - [ ] `Create2ReceiverFactory` exposes one batch recovery function that takes CREATE2 salts and
        init code, derives the receiver addresses, deploys any missing receivers, and calls
        `sweep()` on each receiver in the same transaction.
  - [ ] If any receiver call fails, the whole transaction reverts.
  - [ ] No address-based sweep entry point or generic deploy-and-call recovery helper remains in
        the public contract API.
  - [ ] No separate `SweepBatchCaller` contract or artifact remains in the repo.
- Notes:
  - The factory contract should stay explicit and Ethereum-specific; do not introduce another
    singleton contract just for batch sweep.

### FR-005 - Keep initialization and sweep output reviewable

- Description:
  - Dry-run output must make both deployment and sweep operations easy to inspect before Ledger
    broadcast.
- Acceptance criteria:
  - [ ] Sweep dry-run prints selected payment address IDs, network, receiver count, receiver
        addresses, receiver balances in wei, factory address, receiver deployment states, Ledger
        sender, and final `cast send` command.
  - [ ] Sweep dry-run prints the one canonical recovery call signature.
  - [ ] Deploy dry-run prints network, metadata file, expected Ledger sender, and final deploy
        command.
  - [ ] Output order is deterministic.
  - [ ] Broadcast mode only runs after the same validations as dry-run.
- Notes:
  - Reviewability is part of the security model.

### FR-006 - Remove non-Ledger and extra-contract CREATE2 helper paths

- Description:
  - Existing CREATE2 operator scripts or tool subcommands that depend on private-key signing or a
    second batch-only singleton contract must be removed as part of this change.
- Acceptance criteria:
  - [ ] `scripts/ethereum_create2_verify_chain.sh` is removed.
  - [ ] `cmd/ethereum-create2-tool` is removed.
  - [ ] No `ETHEREUM_CREATE2_VERIFY_OPERATOR_PRIVATE_KEY`-style private-key operator env remains in
        CREATE2 operator helper code or docs.
  - [ ] No `ETHEREUM_SWEEP_BATCH_CALLER_ADDRESS` env remains in the CREATE2 operator flow.
  - [ ] `scripts/ethereum_create2_batch_sweep.sh`,
        `scripts/ethereum_create2_batch_sweep_deploy.sh`, `SweepBatchCaller.sol`, and
        `SweepBatchCallerV1.json` are removed.
- Notes:
  - The checked-in artifact build helper must live under `scripts/`.

### FR-007 - Provide one-time factory deployment that updates metadata

- Description:
  - The repo must provide one Ledger-only deployment helper that deploys `Create2ReceiverFactory`
    and writes the deployed address into checked-in metadata for the selected network.
- Acceptance criteria:
  - [ ] The deploy script supports `--dry-run` and `--broadcast`.
  - [ ] On successful broadcast, the script updates `internal/infrastructure/ethereumcreate2assets/metadata/<network>.json`
        with the deployed factory address and non-fixture mode.
  - [ ] Sweep tooling does not require a separate batch caller env var and uses the active factory
        recorded in checked-in metadata for the selected network.
- Notes:
  - This is the one-time initialization step operators asked for.

## Non-functional requirements

- Performance (NFR-001):
  - The first version must support at least 25 receiver addresses in one batch invocation without
    changing the signer model.
- Availability/Reliability (NFR-002):
  - Input validation must happen before any broadcast attempt.
- Security/Privacy (NFR-003):
  - No private key material may exist in the CREATE2 operator workflow after this change.
- Compliance (NFR-004):
  - No change.
- Observability (NFR-005):
  - Dry-run output must expose enough metadata for manual operator review.
- Maintainability (NFR-006):
  - Keep the implementation Ethereum-specific and explicit; do not introduce generic multi-chain
    sweep abstractions, a second singleton contract, or legacy compatibility branches in the
    canonical operator path.

## Dependencies and integrations

- External systems:
  - PostgreSQL via `psql`
  - `cast`
  - `jq`
  - Ledger hardware wallet
  - Ethereum JSON-RPC endpoint
- Internal services:
  - `internal/infrastructure/ethereumcreate2assets`
  - `scripts/ethereum_create2_sweep.sh`
  - `scripts/ethereum_create2_factory_deploy.sh`
  - `scripts/ethereum_create2_build_artifacts.sh`
