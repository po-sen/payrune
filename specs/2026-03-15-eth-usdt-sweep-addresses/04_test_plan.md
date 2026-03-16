---
doc: 04_test_plan
spec_date: 2026-03-15
slug: eth-usdt-sweep-addresses
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
  - Policy/config parsing for Ethereum mainnet and Sepolia.
  - Asset-aware allocation snapshots and idempotent replay.
  - Ethereum counterfactual vault address prediction and persistence.
  - ETH and ERC20 receipt observation with confirmation handling.
  - Deposit-vault contract prediction, deployment, and `cmd/evm-sweeper` behavior.
  - Bitcoin regression coverage for the issuance-model refactor.
- Not covered:
  - Production RPC/provider SLAs.
  - Real mainnet fund movement during automated CI.
  - Non-Ethereum EVM chains.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, FR-004, NFR-006
  - Steps:
    - Validate supported-chain parsing and policy normalization for `ethereum`, `mainnet`,
      `sepolia`, `eth`, and `usdt`.
  - Expected:
    - Ethereum policies normalize successfully and expose the expected asset snapshot fields.
- TC-002:
  - Linked requirements: FR-002, FR-003, NFR-002, NFR-007
  - Steps:
    - Use a fake RNG to drive the Ethereum issuer and allocate multiple addresses under the same
      policy.
  - Expected:
    - Each non-replayed allocation gets a unique salt/address and idempotent replay returns the
      original address.
- TC-003:
  - Linked requirements: FR-004, FR-005
  - Steps:
    - Apply mocked ETH and ERC20 observation totals to receipt-tracking lifecycle logic.
  - Expected:
    - Status moves through `watching`, `paid_unconfirmed`, and `paid_confirmed` correctly.
- TC-004:
  - Linked requirements: FR-008, NFR-006
  - Steps:
    - Run Bitcoin allocation and generation unit tests against the refactored issuer boundary.
  - Expected:
    - Bitcoin behavior remains unchanged.

### Integration

- TC-101:
  - Linked requirements: FR-001, FR-004, FR-007, NFR-002
  - Steps:
    - Apply postgres migrations to a test database and execute store tests for `evm_factories`,
      asset snapshots, and `evm_payment_vaults`.
  - Expected:
    - Schema applies cleanly, only one active factory exists per network, and EVM vault metadata
      persists with uniqueness guarantees.
- TC-105:
  - Linked requirements: FR-006, FR-007, NFR-004
  - Steps:
    - Load `cmd/evm-sweeper` and `cmd/evm-factory-deploy` runtime config from database-backed
      factory registrations plus direct env and deployment-manifest JSON for both mainnet and
      Sepolia.
  - Expected:
    - The runtime accepts complete config, rejects partial config, and resolves factory plus
      collector addresses from manifest files when direct env is omitted.
  - Note:
    - Dry-run bootstrap accepts compose-provided non-secret defaults without requiring a sweeper
      private key, while execute mode still treats the private key as mandatory.
- TC-106:
  - Linked requirements: FR-006, NFR-004
  - Steps:
    - Run `cmd/evm-factory-deploy --deployment-manifest=...` against a test database and then
      replace the active factory for the same network with a second manifest plus
      `--replace-active`.
  - Expected:
    - The newer factory becomes `active`, the older one becomes `retired`, and both records remain
      queryable for later sweep/version history.
- TC-107:
  - Linked requirements: FR-006, NFR-004
  - Steps:
    - Run local `make up`, then execute `cmd/evm-factory-deploy` or the compose ops equivalent for
      one Ethereum network, and inspect `evm_factories` through the bundled DB viewer or a direct
      SQL query.
  - Expected:
    - Local compose starts without hidden mainnet/sepolia seeding, and the chosen operator deploy
      command writes exactly one active row for the selected network into `evm_factories`.
- TC-102:
  - Linked requirements: FR-004, FR-005, NFR-005
  - Steps:
    - Mock the explorer/indexer provider for ETH and ERC20 transfer responses and run the poller
      use case end-to-end.
  - Expected:
    - Poller stores the right observed and confirmed totals and reschedules correctly on errors.
- TC-103:
  - Linked requirements: FR-006, NFR-003, NFR-007
  - Steps:
    - Deploy factory and vault contracts to a local EVM dev chain, predict addresses, send ETH and
      6-decimal ERC20 balances to those addresses, and execute `cmd/evm-sweeper`.
  - Expected:
    - Funds arrive in the collector account and persisted sweep tx hashes/statuses are updated.
  - Note:
    - The current Ganache-backed contract suite verifies ERC20 counterfactual deposits directly.
      Native ETH sweep is validated after vault deployment because Ganache does not preserve ETH
      balance sent to an undeployed CREATE2 address in this local setup.
- TC-104:
  - Linked requirements: FR-001, FR-004, FR-006
  - Steps:
    - Exercise HTTP controller tests for Ethereum allocation and status endpoints.
  - Expected:
    - Ethereum responses include asset metadata and remain compatible with existing route patterns.

### E2E (if applicable)

- Scenario 1:
  - Start the API and poller against a local database and local EVM chain.
  - Allocate one Sepolia-style ETH address and one Sepolia-style USDT address.
  - Simulate deposits, poll to `paid_confirmed`, then run `cmd/evm-sweeper` to sweep both to the
    collector.
- Scenario 2:
  - Run a mixed regression flow with Bitcoin plus Ethereum allocations in the same environment and
    verify status endpoints remain chain-correct.

## Edge cases and failure modes

- Case:
  - Explorer/indexer provider timeout during Ethereum polling.
- Expected behavior:
  - `last_error` is updated, the tracking is rescheduled, and no false payment transition occurs.
- Case:
  - Duplicate salt or predicted address insertion attempt.
- Expected behavior:
  - Persistence rejects the write and allocation returns an internal error rather than reusing an
    address silently.
- Case:
  - Sepolia USDT policy configured without a token address.
- Expected behavior:
  - The policy is listed as disabled and allocation returns the same disabled-policy error path used
    elsewhere.
- Case:
  - Sweep batch partially fails due to gas limit or contract revert.
- Expected behavior:
  - Failed vault rows remain retryable with error details recorded; successful vaults remain marked
    succeeded.
- Case:
  - Operator runs `cmd/evm-sweeper` with filters that match no eligible vaults.
- Expected behavior:
  - The command exits cleanly, reports zero selected rows, and does not mutate sweep state.
- Case:
  - Bitcoin allocation after the issuer refactor.
- Expected behavior:
  - Bitcoin tests continue to pass and no Ethereum-specific fields break the flow.

## NFR verification

- Performance:
  - Measure Ethereum allocation latency in tests that avoid live chain writes and confirm the p95
    target stays within `<= 300 ms` in local/dev conditions.
- Reliability:
  - Run repeated idempotency, migration, and poller retry tests to confirm address uniqueness and
    stable replay behavior.
- Security:
  - Verify no public API or poller path loads relayer keys and confirm contract tests enforce
    authorized sweep callers only in `cmd/evm-sweeper`.
