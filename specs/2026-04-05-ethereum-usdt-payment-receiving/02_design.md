---
doc: 02_design
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

# Technical Design

## High-level approach

- Summary:
  - Extend the current Ethereum CREATE2 flow with one explicit ERC-20 payment asset: USDT.
  - Keep issuance under the existing `ethereum` chain and `create2` scheme, but make the issued
    payment model asset-aware.
  - Reuse the current balance-snapshot observer pattern by adding ERC-20 `balanceOf(address)`
    observation for token-backed rows.
  - Upgrade the CREATE2 receiver and sweep tooling so operators can recover ERC-20 balances with
    the same Ledger-first workflow already used for ETH.
- Key decisions:
  - The first rollout is explicit to `ethereum + usdt`; no generic token registry layer.
  - Remove redundant asset-shape/display fields from public contracts and use `assetReference`
    plus `decimals` as the only explicit asset metadata needed by clients.
  - Persist one nullable `asset_reference` on issued payment rows. `NULL` means the chain/network
    native asset; non-`NULL` means a chain-scoped token identifier such as an Ethereum token
    contract. Polling and recovery should not depend on mutable policy configuration alone.
  - For ERC-20 receipt tracking, use `eth_call balanceOf(address)` snapshots at latest and
    confirmed block heights, matching the existing ETH observer shape.
  - Keep recovery Ledger-only and cut over to one unified CREATE2 receiver artifact for both
    native ETH and USDT issuance instead of maintaining two nearly identical receiver families.
  - ERC-20 recovery must batch compatible selected receivers into one factory call so operators
    sign once instead of once per receiver.
  - Address separation between native ETH and USDT CREATE2 policies comes from policy-specific salt
    derivation, not from separate receiver bytecode.
  - The repo should not keep an extra token-only receiver source/artifact once the unified receiver is the only
    active receiver family for new issuance.
  - Recovery must trust the selected rows' recorded factory address and init-code material instead
    of forcing every sweep through the current metadata factory.
  - Solidity interfaces should define the stable receiver/factory call surface; shell scripts
    should only depend on those call signatures and row-owned recovery material, not on contract
    implementation details.
  - Checked-in current artifacts should use stable unversioned names, with historical tracking
    coming from git commits, deployed addresses, and row-owned recovery material rather than
    suffix-based artifact versioning.

## System context

- Components:
  - Ethereum address policy bootstrap and HTTP policy/status controllers
  - allocation and receipt-tracking application use cases
  - Ethereum observer adapter
  - PostgreSQL / Cloudflare PostgreSQL persistence adapters
  - CREATE2 receiver contract artifact and Ledger sweep scripts
- Interfaces:
  - `POST /v1/chains/ethereum/payment-addresses`
  - `GET /v1/chains/ethereum/address-policies`
  - `GET /v1/chains/ethereum/payment-addresses/{paymentAddressId}`
  - Ethereum JSON-RPC `eth_call`, `eth_blockNumber`, and related block-scoped calls
  - `scripts/ethereum_create2_sweep.sh`

## Key flows

- Flow 1:
  - Bootstrap builds explicit Ethereum USDT policies from network-scoped env vars, including one
    Ethereum `asset_reference` per enabled USDT policy.
- Flow 2:
  - Allocation issues one CREATE2 payment address from a USDT policy, persists one non-`NULL`
    `asset_reference` plus row-owned recovery material, and returns `assetReference` without
    redundant `assetKind` or `tokenStandard` fields.
- Flow 2a:
  - Allocation issues one CREATE2 payment address from a native ETH policy through the same
    unified receiver artifact; the address still differs from the USDT policy because the
    CREATE2 salt derivation includes `AddressPolicyID`.
- Flow 3:
  - Poller claims due rows directly from `payment_receipt_trackings`, treats `NULL`
    `asset_reference` as the native-asset path, treats Ethereum non-`NULL` `asset_reference` as an
    ERC-20 asset reference, then applies the existing receipt status lifecycle.
  - Core runtime contracts carry `assetReference` as a plain nullable string value; no standalone
    `PaymentAssetReference` value object is retained unless cross-chain invariants later justify it.
  - Call sites read normalized policy `assetReference` directly; no separate
    `ConfiguredAssetReference()` wrapper is retained.
  - Startup validation remains in `internal/bootstrap/api.go`, with one loop and direct
    chain-specific helper calls instead of an extra dispatch helper layer.
  - Shared Ethereum address/hex normalization helpers live in a dedicated helper file within the
    same adapter package instead of being embedded inside `chain_address_deriver.go`.
- Flow 4:
  - Operator selects one or more issued Ethereum USDT rows, the sweep helper validates the shared
    ERC-20 asset reference plus shared row-owned factory address, and the CREATE2 factory/receiver
    path deploys missing receivers and sweeps token balances to the configured collector through
    one Ledger-signed transaction.
- Flow 5:
  - Operator can also send one USDT payment manually with a small Ledger-only helper that resolves
    network and `asset_reference` explicitly from runtime input.

## Diagrams (optional)

- Mermaid sequence / flow:

  ```mermaid
  flowchart TD
    A[Bootstrap USDT policy] --> B[Allocate CREATE2 address]
    B --> C[Persist asset-aware allocation + tracking]
    C --> D[Poller claims due tracking row]
    D --> E{asset_reference null?}
    E -- yes --> F[eth_getBalance snapshots]
    E -- no --> G[eth_call balanceOf snapshots]
    F --> H[Apply observation]
    G --> H
    H --> I[Persist status + webhook outbox]
    I --> J{paid_confirmed?}
    J -- yes --> K[Stop polling]
    J -- no --> L[Reschedule]
    K --> M[Operator dry-run token sweep]
    M --> N[Ledger signed factory batch ERC-20 recovery]
  ```

## Data model

- Entities:
  - `PaymentAddressAllocation` and `PaymentReceiptTracking` persist one nullable `asset_reference`
    for the issued payment flow.
  - `PaymentReceiptTracking` keeps the immutable issuance-time asset snapshot needed by the poller;
    `NULL` means native asset for the row's `chain + network`, and a non-`NULL` value is the
    chain-specific token or asset identifier.
- Schema changes or migrations:
  - Rewrite migration `000017` to add nullable `asset_reference` to both
    `address_policy_allocations` and `payment_receipt_trackings`, assuming the schema state
    immediately after `000016`.
  - Do not make `000017` absorb intermediate draft schemas; it only needs to consume the clean
    pre-feature baseline from `000016`.
- Consistency and idempotency:
  - Allocation still reserves and completes one payment address per idempotent request.
  - Polling stays row-idempotent: one observation update per claimed row, no duplicate terminal
    status transitions.

## API or contracts

- Endpoints or events:
  - Remove `assetKind`, `tokenStandard`, and `minorUnit` from policy, allocation-response,
    status-response, and webhook DTOs; keep `assetReference` as the canonical asset identifier and
    `decimals` as the numeric display contract.
  - Extend the CREATE2 receiver contract and operator sweep contract call path for ERC-20 recovery.
  - Point checked-in CREATE2 metadata at one unversioned unified receiver artifact for new
    issuance.
  - Define stable Solidity interfaces for the factory and receiver sweep surface.
  - Add one Ledger-only operator helper for manual USDT payments.
- Request/response examples:
  - Policy example additions:
    - `assetReference: 0x...`
  - Allocation example additions:
    - one explicit USDT receive example uses concise `ethereumUSDT` naming, `customerReference=null`,
      no `minorUnit`, and `expectedAmountMinor=1000000` to represent `1 USDT`
  - Status/webhook example additions:
    - the same `assetReference` carried alongside existing amount totals
  - Recovery additions:
    - sweep material for ERC-20 rows includes one canonical `asset_reference`
    - factory/receiver call path can deploy and sweep ERC-20 balances for selected rows in one
      transaction

## Backward compatibility (optional)

- API compatibility:
  - Existing endpoints remain in place, but `assetKind`, `tokenStandard`, and `minorUnit` are
    removed from public responses in favor of `assetReference` plus `decimals`.
- Data migration compatibility:
  - Existing Bitcoin and native ETH rows remain valid after migration.
  - Historical Ethereum native rows must continue to read status without requiring manual data
    repair.
  - Historical polling rows must continue to work after the refactor because they carry their own
    `asset_reference` snapshot and can still represent native assets with `NULL`.
  - Factory redeploys change the CREATE2 namespace for new rows on that network.
  - Historical rows remain recoverable because recovery uses each row's recorded factory address
    and init-code material instead of only the current metadata factory.

## Failure modes and resiliency

- Retries/timeouts:
  - Reuse current RPC timeout handling for ERC-20 `eth_call` balance snapshots.
- Backpressure/limits:
  - Sweep helper continues to enforce explicit selection and conservative batch-size limits.
- Degradation strategy:
  - A malformed or unavailable asset reference causes row-level polling failure and retry; it must
    not break unrelated rows or chains.
  - A row batch that mixes factories cannot be sent as one transaction; the sweep helper must fail
    closed before broadcast and ask operators to split the selection.
  - A row batch that mixes asset references must fail closed before broadcast.

## Observability

- Logs:
  - Keep CREATE2 salt material out of default logs.
  - Include `asset_reference`, payment address id, chain, and network in operator/poller
    diagnostics where useful.
- Metrics:
  - Reuse existing poll-cycle counts and failure tracking.
- Traces:
  - None added.
- Alerts:
  - Existing row-level failure reasons remain the primary signal.

## Security

- Authentication/authorization:
  - No public auth changes.
  - Operator recovery remains Ledger-only.
  - Secrets:
  - Network-scoped CREATE2 derivation keys and operator RPC credentials remain existing secret
    surfaces; asset references are configuration, not secrets.
- Abuse cases:
  - Mixed asset-reference recovery selections must fail closed before broadcast.
  - Malformed asset references must fail closed at bootstrap or row-validation time.
  - Receiver token transfer logic must support non-standard ERC-20 return behavior conservatively
    so USDT-like contracts do not silently fail.
  - Manual payment helper usage must validate recipient, amount, network, and Ledger sender before
    broadcast.

## Alternatives considered

- Option A:
  - Introduce a generic asset registry and token framework across all chains.
- Option B:
  - Build a dedicated ERC-20 log indexer and token-transfer event model.
- Why chosen:
  - Option A is broader than the requested scope and would add abstractions this repo does not yet
    need.
  - Option B adds significant complexity when the current poller model can be extended with
    `balanceOf(address)` snapshots for this rollout.
  - The chosen design keeps the new capability explicit, reviewable, and aligned with the current
    ETH implementation style.

## Risks

- Risk:
  - Asset metadata changes touch multiple layers: persistence, DTOs, observer input, and operator
    tooling.
- Mitigation:
  - Land the feature in explicit slices with migration coverage and targeted regression tests for
    Bitcoin, native ETH, and USDT rows.
- Risk:
  - USDT transfer semantics may differ from idealized ERC-20 behavior.
- Mitigation:
  - Use low-level token calls in the receiver and validate empty-return or `true` return behavior
    explicitly in tests.
