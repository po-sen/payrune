---
doc: 03_tasks
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

# Task Plan

## Mode decision

- Selected mode: Full
- Rationale:
  - This change introduces a new chain, new asset-aware policy model, a new technical persistence
    table, new outbound integrations, and a manual sweep command workflow. Quick mode would hide
    critical contract, observation, and migration decisions.
- Upstream dependencies (`depends_on`):
  - None.
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Not applicable.
- If `04_test_plan.md` is skipped:
  - Not applicable.

## Milestones

- M1:
  - Refactor policy and issuance boundaries so Bitcoin and Ethereum can coexist cleanly.
- M2:
  - Ship Ethereum allocation plus receipt observation for ETH and USDT-like ERC20 on mainnet and
    Sepolia.
- M3:
  - Ship operator-side batch sweep contracts and `cmd/evm-sweeper`.
- M4:
  - Complete migration, regression, and integration coverage for Bitcoin plus Ethereum flows.

## Current implementation status

- Completed on 2026-03-16:
  - M1 foundation slice landed: supported-chain parsing, asset-aware policy DTO/domain fields, and
    initial Ethereum wiring in the Go application.
  - M3 contract/deployment slice landed: `DepositVaultFactory`, `DepositVault`, local contract
    tests, deployment manifest generation, and operator-facing wrapper scripts under `scripts/`.
  - M3/T-006 deployment wiring slice landed: `cmd/evm-sweeper` runtime now reads network-specific
    `ETHEREUM_*` env, supports factory/collector resolution from deployment manifests, and has a
    dedicated Docker/compose entry point for operator dry runs.
  - Compose defaults were aligned with the existing Bitcoin style: public EVM RPC defaults live in
    `deployments/compose/compose.yaml`, while deployment-specific contract and token addresses stay
    in `deployments/compose/compose.test.env`.
  - M3 factory-registry slice landed: `evm_factories` now acts as the active factory registry for
    issuance and sweep, and `cmd/evm-sweeper` resolves active factory/collector addresses from the
    database instead of reading them directly from env.
  - M2/T-003 issuance slice landed: allocation now loads the active factory from `evm_factories`,
    predicts CREATE2 vault addresses with persisted salts/code hashes, writes `evm_payment_vaults`,
    and returns Ethereum payment addresses through the existing allocation API.
  - Reservation persistence is aligned with the asset-snapshot schema: fresh or reopened allocation
    rows now write `asset_code`, `asset_type`, `token_address`, `minor_unit`, `decimals`, and
    `issuance_method` at reservation time so Ethereum allocation works cleanly against the migrated
    `NOT NULL` columns before the final issued-address update runs.
  - M2/T-004 observer slice landed for the Go poller runtime: Ethereum receipt observation now
    flows through asset-aware polling inputs, a Blockscout-backed Ethereum observer, and compose
    defaults for mainnet/Sepolia local polling.
  - M3/T-005 execution slice landed: `cmd/evm-sweeper` now loads eligible `paid_confirmed` vaults
    from the database, groups them by network plus asset/factory/collector, submits
    `batchDeployAndSweepNative` or `batchDeployAndSweepToken`, waits for on-chain receipts, and
    persists submitted/succeeded/failed sweep states with tx hashes.
  - M3/T-006 docs slice is aligned: `deployments/swagger/openapi.yaml` now reflects Ethereum chain
    support and the asset-aware response fields already returned by the HTTP API.
  - Local compose keeps DB viewer plus Ethereum operator services, while actual factory deployment
    is now an explicit `cmd/evm-factory-deploy` operator action rather than an automatic
    post-migrate seed.
  - `cmd/evm-factory-sync` is retired from the design; deployment plus DB registration belongs to
    `cmd/evm-factory-deploy`, and manifest-only replay is handled by the same command for recovery.
  - Validation slice landed: repo-wide Go tests, contract tests, and spec lint are passing, while
    the pre-commit suite now only fails on the clean-diff `go mod tidy` gate because this slice
    intentionally adds the `go-ethereum` dependency.
  - Validation evidence recorded:
    - `go test ./...`
    - `go list ./...`
    - `bash scripts/ethereum-contract-test.sh`
    - `SPEC_DIR="specs/2026-03-15-eth-usdt-sweep-addresses" bash scripts/spec-lint.sh`
    - `bash scripts/precommit-run.sh` still fails only on the repository hook that requires the
      new `go.mod` / `go.sum` diff to be committed after adding `go-ethereum`

## Tasks (ordered)

1. T-001 - Generalize the policy and issuance model for asset-aware multi-chain allocation
   - Scope:
     - Add `ethereum` to supported chain parsing.
     - Extend policy DTO/domain snapshots with `assetCode`, `assetType`, `tokenAddress`, and
       `issuanceMethod`.
     - Introduce a chain-specific payment-address issuer port so allocation is no longer xpub-only.
   - Output:
     - Bitcoin and Ethereum can share the same allocation use case without forcing Ethereum through
       derivation-path inputs.
   - Linked requirements: FR-001, FR-004, FR-008, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): run domain/use-case tests covering policy
           normalization, chain parsing, and allocation snapshots.
     - [ ] Expected result: Ethereum policies are accepted, Bitcoin regressions stay green, and the
           allocation use case compiles against the new issuer abstraction.
     - [ ] Logs/metrics to check (if applicable): none
1. T-002 - Add persistence and policy configuration for Ethereum deposit vaults
   - Scope:
     - Create schema migrations for `evm_factories`, asset snapshot columns, and
       `evm_payment_vaults`.
     - Add postgres/cloudflarepostgres stores for EVM vault metadata.
     - Add a database-backed active-factory registry so factory and collector addresses are not
       managed only in env.
     - Wire mainnet and Sepolia ETH/USDT policy configs in DI, including a configurable Sepolia
       ERC20 test token.
   - Output:
     - The database can persist Ethereum allocation metadata and the app can expose ready/disabled
       policies per environment.
   - Linked requirements: FR-001, FR-004, FR-007, NFR-002, NFR-004
   - Validation:
     - [ ] How to verify (manual steps or command): run migration tests/store tests and start the
           API with sample Sepolia config.
     - [ ] Expected result: migrations apply cleanly and policy listing shows enabled/disabled
           Ethereum policies correctly.
     - [ ] Logs/metrics to check (if applicable): policy configuration logs show explicit network
           and token-address selections
1. T-003 - Implement Ethereum counterfactual address issuance
   - Scope:
     - Implement Ethereum issuer adapter that generates secure random salts, predicts vault
       addresses, and persists EVM vault metadata linked to the generic allocation row.
     - Load the active factory generation from `evm_factories` during allocation rather than from
       env defaults.
     - Keep idempotency replay behavior identical to the current API contract.
   - Output:
     - `POST /v1/chains/ethereum/payment-addresses` returns unique Ethereum payment addresses
       without waiting for an on-chain deployment.
   - Linked requirements: FR-002, FR-003, FR-008, NFR-001, NFR-002, NFR-007
   - Validation:
     - [ ] How to verify (manual steps or command): run allocation use-case tests with fake RNG,
           idempotency replay tests, and store tests for unique salt/address constraints.
     - [ ] Expected result: unique addresses are issued, replay returns the same address, and no
           on-chain dependency is required in the issuance path.
     - [ ] Logs/metrics to check (if applicable): allocation logs include
           `issuanceMethod=create2_forwarder`
1. T-004 - Implement Ethereum receipt observation for ETH and ERC20
   - Scope:
     - Extend observer ports/input DTOs with asset metadata.
     - Add Ethereum observer adapter for native and ERC20 transfer observation through a configurable
       provider.
     - Integrate Ethereum observation into the existing poller lifecycle.
   - Output:
     - Ethereum payment addresses transition through the existing receipt lifecycle based on ETH or
       ERC20 activity.
   - Linked requirements: FR-004, FR-005, FR-007, FR-008, NFR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): run observer integration tests with mocked
           provider responses and poller use-case tests for ETH and USDT.
     - [ ] Expected result: confirmed/unconfirmed totals are computed correctly and poller status
           transitions remain stable across reorg-safe confirmation depths.
     - [ ] Logs/metrics to check (if applicable): receipt polling logs and error counters for the
           configured provider
1. T-005 - Implement deposit-vault contracts and batch sweep execution
   - Scope:
     - Add the deposit-vault factory/implementation contracts under the repo's existing architecture
       conventions.
     - Add build/deploy/test automation under `scripts/`.
     - Implement Ethereum sweep executor, a command-oriented sweep use case, and `cmd/evm-sweeper`.
   - Output:
     - Operators can batch collect confirmed ETH and USDT-like deposits into one collector account.
   - Linked requirements: FR-006, FR-007, NFR-003, NFR-005, NFR-007
   - Validation:
     - [ ] How to verify (manual steps or command): run contract tests plus `cmd/evm-sweeper`
           integration tests against a local EVM dev chain.
     - [ ] Expected result: predicted addresses match the contract implementation and batch sweep
           moves balances to the collector with persisted tx hashes.
     - [ ] Logs/metrics to check (if applicable): command sweep logs, tx hash logs, and failure
           counters
1. T-006 - Finish API, docs, and deployment wiring

- Scope:
  - Extend HTTP DTOs/status output with asset metadata.
  - Add mainnet/Sepolia env loading, `cmd/evm-factory-deploy`, `cmd/evm-sweeper` config,
    compose defaults, and operator docs.
  - Document the difference between deterministic Bitcoin derivation and Ethereum random vault
    allocation.
- Output:
  - The API contract and local/dev deployment docs fully describe Ethereum usage and operator
    config.
- Linked requirements: FR-001, FR-004, FR-006, FR-007, NFR-004, NFR-006
- Validation:
  - [ ] How to verify (manual steps or command): run API controller tests, OpenAPI/doc checks,
        and local bootstrap smoke tests.
  - [ ] Expected result: Ethereum responses expose asset fields consistently and deployment docs
        match the new env vars and `cmd/evm-sweeper`.
  - [ ] Logs/metrics to check (if applicable): bootstrap logs show configured Ethereum networks

1. T-007 - Run full regression and pre-commit verification
   - Scope:
     - Run targeted and repo-wide tests for Bitcoin plus Ethereum, including migrations, stores,
       use cases, HTTP, poller, and command-driven sweep flows.
   - Output:
     - The repository is validated end-to-end with the new chain support.
   - Linked requirements: FR-008, NFR-001, NFR-002, NFR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): run spec lint, targeted Go tests, contract
           tests, `go list ./...`, and `bash scripts/precommit-run.sh`.
     - [ ] Expected result: all required validation passes and Bitcoin remains green.
     - [ ] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001, T-002, T-006
- FR-002 -> T-003
- FR-003 -> T-003
- FR-004 -> T-001, T-002, T-004, T-006
- FR-005 -> T-004
- FR-006 -> T-005, T-006
- FR-007 -> T-002, T-004, T-005, T-006
- FR-008 -> T-001, T-003, T-004, T-007
- NFR-001 -> T-003, T-007
- NFR-002 -> T-002, T-003, T-007
- NFR-003 -> T-005
- NFR-004 -> T-002, T-006
- NFR-005 -> T-004, T-005, T-007
- NFR-006 -> T-001, T-004, T-006, T-007
- NFR-007 -> T-003, T-005

## Rollout and rollback

- Feature flag:
  - Not required initially; policies may remain disabled until env, provider, and contracts are in
    place.
- Migration sequencing:
  - Apply schema changes first.
  - Deploy factory/implementation contracts per network.
  - Configure policies and observers.
  - Enable Sepolia first, then enable mainnet after end-to-end validation.
- Rollback steps:
  - Disable Ethereum policies in config.
  - Stop running `cmd/evm-sweeper` or revoke relayer credentials if sweep execution is the source
    of failure.
  - Revert application code and migrations only if schema rollback is explicitly planned and safe.
