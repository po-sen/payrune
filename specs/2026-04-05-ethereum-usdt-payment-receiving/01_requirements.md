---
doc: 01_requirements
spec_date: 2026-04-05
slug: ethereum-usdt-payment-receiving
mode: Full
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-20-create2-eth-payment-receiving
  - 2026-03-30-eth-balance-receipt-observer
  - 2026-04-03-ethereum-ledger-batch-sweep
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Requirements

## Glossary (optional)

- Asset reference:
  - A nullable, chain-scoped identifier for one non-native payment asset. `NULL` means the native
    asset of the row's `chain + network`; a non-`NULL` value identifies the chain-specific token or
    asset reference, for example Ethereum ERC-20 token contract `0xdac17f...`.
- USDT policy:
  - An Ethereum CREATE2 address policy whose issued addresses are meant to receive USDT balances,
    not native ETH balances.
- Token-capable receiver:
  - A CREATE2 receiver contract that can recover ERC-20 balances to the configured collector, in
    addition to native ETH.
- Unified receiver:
  - The single CREATE2 receiver artifact used by both Ethereum native ETH and Ethereum USDT
    policies after the cutover.
- Row-owned recovery material:
  - The `sweep_material_json` stored on one issued payment row, including its factory address,
    init code, and init-code hash.

## Out-of-scope behaviors

- OOS1:
  - Generic support for every ERC-20 token contract in the same rollout.
- OOS2:
  - Non-Ethereum chain support.
- OOS3:
  - Mempool subscriptions or pending-transfer payment detection.
- OOS4:
  - Automatic operator sweep scheduling.

## Functional requirements

### FR-001 - Expose Ethereum USDT payment policies through the existing payment-address API

- Description:
  - The public payment-address flow must expose USDT as an explicit Ethereum payment asset without
    changing the top-level chain-scoped API shape.
- Acceptance criteria:
  - [ ] `GET /v1/chains/ethereum/address-policies` can list at least one enabled USDT policy with
        `scheme=create2`, `decimals=6`, and `assetReference` that identifies the asset as ERC-20
        USDT.
  - [ ] `POST /v1/chains/ethereum/payment-addresses` can allocate one USDT payment address by
        selecting a USDT-specific `addressPolicyId` and still uses the current request body shape.
  - [ ] `GET /v1/chains/ethereum/payment-addresses/{paymentAddressId}` includes `assetReference`
        so clients can distinguish Ethereum USDT from Ethereum native ETH.
  - [ ] Existing Bitcoin and native ETH policy flows remain available after removing `assetKind`
        , `tokenStandard`, and `minorUnit` from policy, allocation, status, and webhook payloads.
- Notes:
  - This feature must not fork the API into a separate token-only endpoint family, but it may make
    a breaking contract cleanup by removing redundant asset-shape fields.

### FR-002 - Make issued payment state asset-aware for Ethereum token policies

- Description:
  - Issued allocations and receipt tracking rows must carry enough persisted asset metadata to let
    polling, status lookup, and operator recovery behave correctly for ERC-20 payments.
- Acceptance criteria:
  - [ ] New or updated persistence fields capture one nullable `asset_reference` for issued rows,
        where `NULL` means native asset and a non-`NULL` value is sufficient to derive the
        Ethereum token contract needed for ERC-20 observation and recovery.
  - [ ] `payment_receipt_trackings` stores the same issued `asset_reference` snapshot so poller
        claim paths do not need to join `address_policy_allocations` just to identify whether the
        row is native or token-backed.
  - [ ] The domain/application model keeps asset identity explicit at the payment level instead of
        inferring it only from `chain` or address text.
  - [ ] Existing issued Bitcoin and native ETH rows remain readable after the migration.
  - [ ] Existing uniqueness and idempotency rules for payment-address issuance still hold.
- Notes:
  - The implementation must not hide asset identity only inside policy naming conventions or
    adapter-specific recovery JSON.

### FR-003 - Observe USDT payments through Ethereum ERC-20 balance snapshots

- Description:
  - The Ethereum observer must be able to observe ERC-20 USDT balances for issued payment
    addresses using the current polling lifecycle.
- Acceptance criteria:
  - [ ] The observer can read ERC-20 `balanceOf(address)` for the issued payment address at the
        latest block height and at the current confirmed block height.
  - [ ] The observer maps those balances into the existing
        `observed_total_minor / confirmed_total_minor / unconfirmed_total_minor` model.
  - [ ] The poller can process Ethereum USDT rows without affecting Bitcoin or native ETH rows.
  - [ ] Provider or ABI-call failures persist row-level polling failures and reschedule retry
        through the existing failure path.
- Notes:
  - This rollout intentionally reuses the current balance-snapshot polling model instead of
    introducing a new log-indexing subsystem.

### FR-004 - Extend CREATE2 receiver recovery for USDT balances

- Description:
  - Operators must be able to recover ERC-20 USDT balances from selected CREATE2 receiver
    addresses through the current Ledger-based recovery stance.
- Acceptance criteria:
  - [ ] The CREATE2 receiver contract used for new Ethereum CREATE2 issuance can sweep native ETH
        and ERC-20 token balances to the configured collector.
  - [ ] Recovery payload and operator tooling can identify the ERC-20 asset reference for selected
        USDT rows from row-owned material even if policy metadata changes later.
  - [ ] The existing sweep tooling validates mixed selections conservatively and rejects batches
        that combine incompatible asset references.
  - [ ] Mixed-asset recovery validation errors render native rows as `<native>` instead of an
        empty asset reference so operators can split the selection correctly.
  - [ ] The ERC-20 recovery path remains Ledger-only and dry-run-first.
  - [ ] When multiple compatible ERC-20 rows are selected, the operator can recover them with one
        Ledger-signed factory transaction, even when some receivers still need CREATE2 deployment.
  - [ ] New native ETH and USDT CREATE2 policies use the same receiver artifact family;
        address separation remains deterministic through policy-specific CREATE2 salt derivation.
  - [ ] Recovery uses the selected rows' recorded factory address and init-code material, so old
        rows remain sweepable after the network's default factory metadata changes.
- Notes:
  - Auto-sweep is out of scope; operator-triggered recovery is sufficient.

### FR-005 - Keep configuration explicit and network-scoped

- Description:
  - USDT support must be enabled only when the required Ethereum token configuration is present for
    that network.
- Acceptance criteria:
  - [ ] Runtime configuration includes one explicit non-native `assetReference` for every enabled
        Ethereum USDT policy.
  - [ ] Startup fails closed when the configured `assetReference` is malformed or missing for an
        enabled USDT policy, while native ETH policies may still omit `assetReference`.
  - [ ] Local/test deployment config exposes explicit network-scoped env vars rather than hidden
        prefix-driven defaults.
  - [ ] Checked-in OpenAPI and README examples show at least one Ethereum USDT payment flow.
  - [ ] Checked-in CREATE2 metadata points new issuance at the unified receiver artifact for
        supported Ethereum networks.
  - [ ] Checked-in CREATE2 asset sources and generated artifacts do not retain unused token-only
        receiver files once the unified receiver cutover is complete.
  - [ ] Checked-in CREATE2 artifact filenames use the current unversioned runtime artifact set.
- Notes:
  - Keep the naming explicit to Ethereum and USDT; do not add a generic token registry just to
    avoid a few explicit env vars.

### FR-006 - Provide a simple Ledger-signed USDT payment helper

- Description:
  - The repo must include one small operator helper that sends a USDT payment with a Ledger signer
    for Sepolia or mainnet testing.
- Acceptance criteria:
  - [ ] The helper lives under `scripts/`.
  - [ ] The helper supports dry-run and `--broadcast`.
  - [ ] The helper resolves Sepolia or mainnet from the configured RPC URL and uses the
        network-appropriate USDT contract unless explicitly overridden.
  - [ ] The helper remains Ledger-only and does not support raw private-key signing.
- Notes:
  - This helper is for manual testing and operator workflows, not for runtime service behavior.

## Non-functional requirements

- Performance (NFR-001):
  - A warm-path USDT payment-address allocation remains an in-process operation with no mandatory
    chain RPC dependency in the allocation request path.
- Availability/Reliability (NFR-002):
  - Polling the same USDT payment row repeatedly must be idempotent and must not create duplicate
    status transitions or duplicate notification rows.
- Security/Privacy (NFR-003):
  - Public APIs, webhooks, and default logs must not expose raw CREATE2 salts, and operator token
    recovery plus helper payments must remain Ledger-only.
- Compliance (NFR-004):
  - No change.
- Observability (NFR-005):
  - Polling and recovery failures for USDT rows must surface through the existing row-level failure
    paths and dry-run operator output.
- Maintainability (NFR-006):
  - The implementation must stay explicit to `ethereum + usdt` and must not introduce a generic
    abstraction layer that only exists for hypothetical future assets.

## Dependencies and integrations

- External systems:
  - Ethereum JSON-RPC endpoint
  - PostgreSQL
  - `cast`
  - `jq`
  - Ledger hardware wallet
- Internal services:
  - `internal/adapters/outbound/ethereum`
  - `internal/infrastructure/ethereumcreate2assets`
  - existing payment allocation, polling, status, and webhook use cases
