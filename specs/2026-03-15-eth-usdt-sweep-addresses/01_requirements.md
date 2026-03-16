---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Deposit vault:
  - A per-payment Ethereum address created from a factory/implementation design, intended to hold
    ETH or ERC20 balances and later sweep them to one collector account.
- Collector:
  - The operator-controlled Ethereum account that ultimately receives swept funds.
- Asset policy:
  - A caller-visible policy describing chain, network, asset, decimals, issuance method, and any
    token contract address needed to issue and observe a payment address.
- Counterfactual address:
  - A deterministic address predicted from deployment parameters before the contract code is
    deployed on-chain.
- Sweep batch:
  - One relayer/factory operation triggered by `cmd/evm-sweeper` that deploys and/or sweeps
    multiple deposit vaults into the collector account.

## Out-of-scope behaviors

- OOS1:
  - Supporting arbitrary ERC20 tokens beyond the configured USDT-like policies in this slice.
- OOS2:
  - Exposing sweep execution to public unauthenticated HTTP endpoints.
- OOS3:
  - Running an always-on automatic Ethereum sweeper worker.
- OOS4:
  - Automatically topping up per-address gas balances, because the selected vault pattern avoids
    per-address EOAs and gas management.

## Functional requirements

### FR-001 - Ethereum policies must be first-class alongside Bitcoin

- Description:
  - The system must support `ethereum` as a valid chain and expose policy metadata for both
    mainnet and Sepolia.
- Acceptance criteria:
  - [ ] `supported_chain` parsing and chain-routed HTTP endpoints accept `ethereum`.
  - [ ] Address-policy listing returns at least four ready-to-configure policies:
        `ethereum-mainnet-eth`, `ethereum-mainnet-usdt`, `ethereum-sepolia-eth`,
        `ethereum-sepolia-usdt`.
  - [ ] Each Ethereum policy exposes asset metadata required by clients:
        `assetCode`, `assetType`, `decimals`, and `tokenAddress` when applicable.
- Notes:
  - Bitcoin policies remain available and unchanged unless explicitly extended by the same refactor.

### FR-002 - Each Ethereum allocation must issue a unique random deposit address

- Description:
  - Each call to allocate an Ethereum payment address must create a unique deposit-vault allocation
    using cryptographically secure random material rather than a reused sequential derivation path.
- Acceptance criteria:
  - [ ] A new Ethereum allocation persists a unique 32-byte salt or equivalent unique issuance
        reference linked to the `payment_address_id`.
  - [ ] The returned address is deterministic from the persisted issuance metadata and remains
        stable on idempotent replay.
  - [ ] Two successful non-replayed allocations for the same policy do not return the same address.
- Notes:
  - The implementation may still keep an internal primary key or sequence, but the caller-visible
    address issuance model must not depend on Bitcoin-style derivation paths.

### FR-003 - Address issuance must not require one on-chain deployment per payment

- Description:
  - The selected Ethereum issuance flow must keep the API fast and avoid synchronous per-payment
    contract deployment as a requirement for returning an address.
- Acceptance criteria:
  - [ ] The issuance use case can return an Ethereum payment address after local persistence without
        waiting for a deployment transaction confirmation.
  - [ ] The persisted metadata is sufficient for a later `cmd/evm-sweeper` run or relayer call to
        deploy and sweep the same vault address.
  - [ ] The design documents the contract invariant that allows balances sent to the predicted
        address before deployment to be recovered after deployment.
- Notes:
  - This requirement is the main reason for choosing a counterfactual CREATE2/clone-based vault
    pattern instead of per-payment deployment.

### FR-004 - Payment tracking must be asset-aware for Ethereum

- Description:
  - Receipt tracking and payment-status lookup must distinguish ETH from ERC20 deposits on the same
    chain/network.
- Acceptance criteria:
  - [ ] A persisted payment allocation and receipt tracking record include an immutable asset
        snapshot containing at least `assetCode`, `assetType`, and `tokenAddress` when applicable.
  - [ ] Polling logic observes ETH balances/transfers for `assetType=native` and ERC20 `Transfer`
        activity for `assetType=erc20`.
  - [ ] `GET /v1/chains/{chain}/payment-addresses/{paymentAddressId}` returns the asset snapshot
        together with existing payment status fields.
- Notes:
  - Policy lookups alone are not sufficient because tracking must remain stable even if operator
    config changes later.

### FR-005 - Ethereum deposits must follow the existing receipt lifecycle

- Description:
  - Once an Ethereum payment address is issued, the system must track observed, unconfirmed, and
    confirmed amounts using the same domain lifecycle as existing chains.
- Acceptance criteria:
  - [ ] ETH and ERC20 observations update `observedTotalMinor`, `unconfirmedTotalMinor`, and
        `confirmedTotalMinor`.
  - [ ] Confirmation thresholds are configurable per Ethereum network and applied to both ETH and
        USDT policies on that network.
  - [ ] Existing terminal and expiration behaviors remain compatible with the current
        `payment_receipt_trackings` lifecycle policy.
- Notes:
  - This requirement keeps the product-facing status model stable across chains.

### FR-006 - Sweep execution must support ETH and USDT batch collection

- Description:
  - The system must provide an operator-side manual sweep flow through `cmd/evm-sweeper` that can
    collect funds from multiple issued Ethereum deposit vaults into one collector account.
- Acceptance criteria:
  - [ ] `cmd/evm-sweeper` can select eligible Ethereum allocations and group them by network plus
        asset.
  - [ ] Factory and collector addresses used for sweep execution are loaded from a database-backed
        factory registry so multiple factory generations can coexist per network history.
  - [ ] Factory addresses are not treated as long-lived env defaults; operator deployment uses
        `cmd/evm-factory-deploy` to deploy the contract and register the active row in one flow.
  - [ ] `cmd/evm-factory-deploy` can also recover from a partial failure by re-registering an
        existing deployment manifest without requiring a separate sync/import command.
  - [ ] The Ethereum outbound adapter can execute a batch sweep for native ETH vaults and a batch
        sweep for ERC20 vaults.
  - [ ] `cmd/evm-sweeper` supports operator-controlled filters including at least `paymentAddressId`,
        `network`, `assetCode`, and `beforeIssuedAt` or equivalent time bound.
  - [ ] Sweep persistence records the sweep tx hash and terminal success/failure state per vault.
- Notes:
  - The public API does not expose sweep execution. This slice uses an internal command, not a
    background worker.

### FR-007 - Testnet USDT must be configurable rather than hard-coded to an unofficial contract

- Description:
  - The Sepolia USDT policy must support a configurable ERC20 test token so the system can be
    tested without assuming an official Tether Sepolia deployment.
- Acceptance criteria:
  - [ ] The Sepolia USDT-like policy reads its `tokenAddress` from config.
  - [ ] The policy remains disabled when no test token address is configured.
  - [ ] Test coverage uses a 6-decimal ERC20 token to match USDT amount semantics.
- Notes:
  - Ethereum mainnet USDT remains a fixed operator-visible policy with the known mainnet contract.

### FR-008 - Bitcoin support must remain intact during the issuance-model refactor

- Description:
  - Refactoring issuance to support Ethereum must not regress existing Bitcoin address generation,
    allocation, or receipt polling behavior.
- Acceptance criteria:
  - [ ] Bitcoin policy listing and address allocation tests remain green.
  - [ ] Bitcoin continues to persist derivation-path metadata as before.
  - [ ] The new abstraction boundary does not force Bitcoin-specific fields into Ethereum flows or
        Ethereum-specific fields into Bitcoin flows beyond shared allocation snapshots.
- Notes:
  - This is a compatibility requirement for the refactor itself.

## Non-functional requirements

- Performance (NFR-001):
  - Ethereum address allocation must remain local-CPU/DB work with target API latency p95
    `<= 300 ms` excluding database network latency variance in local/dev environments.
- Availability/Reliability (NFR-002):
  - Idempotent replay must return the original allocation when the same `(chain, idempotency key)`
    is reused, and salt/address uniqueness must be enforced by persistence constraints.
- Security/Privacy (NFR-003):
  - The application must not store per-deposit private keys. Sweep signing credentials must be used
    only in `cmd/evm-sweeper` and its dedicated relayer code paths, and token transfers must
    handle USDT's ERC20 compatibility quirks safely.
- Compliance (NFR-004):
  - Mainnet and testnet config must be explicit and readable; no hidden defaults should select a
    production token or collector address.
- Observability (NFR-005):
  - The implementation must emit structured logs and metrics for allocation issuance, receipt
    polling lag, `cmd/evm-sweeper` invocations, sweep attempts, sweep failures, and explorer/indexer
    dependency errors.
- Maintainability (NFR-006):
  - Ethereum-specific issuance, observation, and sweep logic must live in dedicated outbound
    adapters and ports rather than leaking chain-specific branching into HTTP controllers or
    bootstrap code.
- Cost efficiency (NFR-007):
  - Per-payment issuance must not require a gas-spending deployment transaction; on-chain cost
    should be paid only when sweeping or when an operator explicitly deploys a vault.

## Dependencies and integrations

- External systems:
  - Ethereum mainnet RPC and/or explorer/indexer provider.
  - Ethereum Sepolia RPC and/or explorer/indexer provider.
  - Deployed deposit-vault factory and implementation contracts per supported network.
  - Ethereum mainnet USDT ERC20 contract.
  - Operator-configured Sepolia ERC20 test token with 6 decimals.
- Internal services:
  - `internal/application/usecases/allocate_payment_address_use_case.go`
  - `internal/application/usecases/run_receipt_polling_cycle_use_case.go`
  - `internal/adapters/outbound/blockchain`
  - `internal/adapters/outbound/persistence/postgres`
  - new `internal/adapters/outbound/ethereum`
