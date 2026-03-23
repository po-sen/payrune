---
doc: 02_design
spec_date: 2026-03-20
slug: create2-eth-payment-receiving
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-16-remove-xpub-fingerprint
  - 2026-03-05-blockchain-receipt-polling-service
  - 2026-03-08-payment-address-status-api
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# CREATE2 ETH Payment Receiving - Technical Design

## High-level approach

- Summary:
  - Keep the existing chain-scoped payment-address API and receipt lifecycle.
  - Add a first-class Ethereum issuance path that predicts native ETH payment addresses with
    CREATE2 instead of generating EOAs.
  - Extend the current allocation and receipt-tracking flow with explicit Ethereum adapters for
    CREATE2 prediction and native ETH observation.
  - Clean up issuance naming so Bitcoin HD derivation and Ethereum CREATE2 no longer share
    misleading field names.
- Key decisions:
  - Public chain id should be `ethereum`, not `eth`, to stay consistent with the explicit style of
    `bitcoin`.
  - The first configured Ethereum networks should be `mainnet` and `sepolia`, each with explicit
    CREATE2 policy ids plus per-network collector and derivation-key runtime config rather than one
    implicit shared EVM config.
  - Factory addresses should come from checked-in deployment metadata, and receiver init code
    hashes should be derived from checked-in contract artifacts instead of hand-entered runtime
    env vars.
  - Checked-in deployment metadata may stay public and reviewable. The privacy boundary is that
    public metadata, public API behavior, and public sequential guesses together still must not be
    enough to enumerate future payment addresses.
  - Until T-003 replaces them with deployed values, checked-in deterministic fixture metadata may
    be used to keep local API and Swagger testing paths enabled for configured Ethereum policies.
  - Public default Ethereum RPC endpoints, when provided for local bootstrap convenience, should
    live in deployment config (`compose`, `wrangler`) rather than inside Go loader code so
    operators can override them without changing runtime wiring semantics.
  - Poller deployment should stay scope-specific: one runtime or scheduled job per
    `(chain, network)` pair, even if one binary can construct multiple chain observers.
  - Compose-level poller service names should also stay scope-explicit, for example
    `poller-bitcoin-mainnet`, `poller-bitcoin-testnet4`, `poller-ethereum-mainnet`, and
    `poller-ethereum-sepolia`, so local operations do not mix chain and network identity.
  - The production-like base Compose file should keep mainnet-scoped pollers, while
    `compose.test.yaml` may add test or verification scopes such as Bitcoin `testnet4` and
    Ethereum `sepolia` for local development.
  - Policy `scheme` for this flow should be `create2`.
  - Allocation cursoring may continue to use the existing deterministic sequence model for internal
    reservation and uniqueness, but Ethereum CREATE2 salt must not be derived only from public
    policy metadata plus `derivationIndex`. Each allocation must derive non-public salt material
    from a runtime-managed secret plus stable allocation identity, so operators can reconstruct one
    issued address without relying on a persisted random salt as the only recovery path.
  - Public `GET /v1/chains/{chain}/addresses?addressPolicyId=...&index=...` preview semantics do
    not fit privacy-preserving Ethereum issuance and should remain Bitcoin-only or reject Ethereum
    CREATE2 policies.
  - Replace xpub-biased persistence and config naming with neutral equivalents such as
    `address_source_ref` and `address_reference`.
  - Treat privacy scope honestly: v1 aims to prevent public enumeration of future Ethereum payment
    addresses before settlement, not to provide complete anonymity after funds are swept on-chain.
  - Keep Ethereum technical process state explicit and chain-specific rather than hiding it inside
    a generic untyped blob.

## System context

- Components:
  - Domain:
    - Existing `PaymentAddressAllocation`, `PaymentReceiptTracking`, and payment receipt status
      transitions.
    - Extended issuance configuration that can represent Bitcoin HD and Ethereum CREATE2 without
      fake placeholder fields.
  - Application:
    - Existing `AllocatePaymentAddressUseCase`
    - Existing `GetPaymentAddressStatusUseCase`
    - Existing `RunReceiptPollingCycleUseCase`
  - Outbound adapters:
    - Existing Bitcoin address deriver and receipt observer
    - New `ethereum` CREATE2 address deriver
    - New `ethereum` native ETH receipt observer
  - Infrastructure:
    - Ethereum JSON-RPC client
    - Thin verification shell wrappers under `scripts/`
    - Go contract tooling under `cmd/` for build, prediction, and explicit chain verification
    - Deployment config under `deployments/compose/` and `deployments/cloudflare/payrune/` with
      separate `mainnet` and `sepolia` collector and derivation-key envs plus checked-in CREATE2
      deployment metadata and receiver artifacts
- Interfaces:
  - Existing public API:
    - `GET /v1/chains/ethereum/address-policies`
    - `POST /v1/chains/ethereum/payment-addresses`
    - `GET /v1/chains/ethereum/payment-addresses/{paymentAddressId}`
  - Internal runtime:
    - Receipt poller cycle for Ethereum rows
  - Contract-side:
    - CREATE2 factory deployment method

## Key flows

- Flow 1: allocate one ETH payment address

  - Client calls `POST /v1/chains/ethereum/payment-addresses`.
  - Controller validates the existing request payload and chain path.
  - Allocation use case loads the Ethereum CREATE2 policy, reserves the next deterministic slot,
    derives one allocation-specific non-public salt from runtime-managed secret material, and
    derives a predicted ETH address in-process using the configured factory and init code
    semantics.
  - Allocation persistence stores:
    - public payment address
    - neutral source-reference metadata
    - neutral address-reference metadata that includes the derived CREATE2 salt or equivalent
      deterministic locator
  - Receipt tracking row is created using the existing lifecycle with `chain=ethereum`,
    `network=<configured network>`, `address=<predicted address>`, and ETH amount metadata.
  - Public preview-by-index endpoints do not expose the same issuance path for Ethereum because the
    address is no longer derivable from public index inputs alone.

- Flow 2: observe an ETH payment

  - Poller claims due Ethereum receipt rows.
  - Ethereum observer resolves a bounded scan range from `issued_at`, `since_block_height`, and
    `latest_block_height`.
  - Observer loads blocks or transactions from the configured RPC source and sums native ETH value
    transferred to the tracked payment address.
  - For v1, the observer may scan canonical block transactions with `to == payment address`
    semantics rather than provider-specific execution traces; this keeps the implementation
    portable across standard JSON-RPC providers at the cost of excluding trace-only internal ETH
    transfers from the initial scope.
  - Observer returns `observed_total_minor`, `confirmed_total_minor`,
    `unconfirmed_total_minor`, and `latest_block_height`.
  - Existing receipt-tracking domain logic updates status and drives webhook-outbox behavior.

- Flow 3: startup preflight
  - Bootstrap validates Ethereum addresses, init-code expectations, confirmation settings, fixed
    factory metadata, collector configuration, and derivation-key configuration.
  - Checked-in CREATE2 contract artifacts and deployment metadata should be loaded through a
    dedicated infrastructure asset package instead of living inside the DI wiring package.
  - If configured, the runtime compares Go-side prediction vectors against the active contract
    metadata and fails fast on mismatch before issuing payment addresses.

## Diagrams (optional)

- Mermaid sequence / flow:

```mermaid
sequenceDiagram
    participant U as Client
    participant C as HTTP Controller
    participant A as AllocatePaymentAddressUseCase
    participant D as EthereumCreate2Deriver
    participant P as Receipt Poller
    participant O as EthereumReceiptObserver
    participant F as CREATE2 Factory

    U->>C: POST /v1/chains/ethereum/payment-addresses
    C->>A: allocate(chain=ethereum, policy, amount)
    A->>D: derive(factory, collector behavior, stored salt)
    D-->>A: predicted payment address + salt
    A-->>C: issued payment address
    C-->>U: 201 Created

    P->>O: observe(address, issuedAt, sinceBlockHeight)
    O-->>P: observed/confirmed totals
    P-->>P: update receipt status
```

## Data model

- Entities:
  - `PaymentAddressAllocation` remains the business record for one issued payment address.
  - `PaymentReceiptTracking` remains the business record for payment observation and lifecycle
    status.
  - No new Ethereum-specific domain aggregate is required for deployment or sweep; that state is a
    technical process record.
- Technical records:
  - No Ethereum-specific deploy-and-sweep technical table is introduced in the first rollout.
- Schema changes or migrations:
  - Add a migration that renames or replaces generic allocation columns so the active schema uses
    neutral naming instead of `account_public_key` and `derivation_path`.
  - Keep existing Bitcoin rows readable and migratable.
- Consistency and idempotency:
  - Allocation uniqueness remains enforced by `(address_policy_id, address_source_ref,
derivation_index)`.
  - For Ethereum, `derivation_index` remains an internal allocation cursor; the actual CREATE2
    address prediction additionally depends on the persisted non-public salt.
  - Public address uniqueness remains enforced by `(chain, address)`.
  - No separate Ethereum collection task or sweeper-claim state exists in the first rollout.

## API or contracts

- Endpoints or events:
  - Existing public endpoints remain the primary contract:
    - `GET /v1/chains/ethereum/address-policies`
    - `POST /v1/chains/ethereum/payment-addresses`
    - `GET /v1/chains/ethereum/payment-addresses/{paymentAddressId}`
  - Public preview-by-index endpoint behavior differs by chain:
    - Bitcoin may keep `GET /v1/chains/{chain}/addresses`
    - Ethereum CREATE2 policies should reject or disable this route to avoid making future payment
      addresses enumerable
  - Existing payment receipt status webhook events remain unchanged in shape.
  - Contract caller model should satisfy:
    - caller cannot override the collector destination
  - Internal contract methods should expose at least:
    - address computation semantics that match Go-side prediction
    - deterministic deployment
- Request/response examples:

```http
POST /v1/chains/ethereum/payment-addresses
Content-Type: application/json

{
  "addressPolicyId": "ethereum-mainnet-create2",
  "expectedAmountMinor": 15000000000000000,
  "customerReference": "order-1234"
}
```

```json
{
  "paymentAddressId": "501",
  "addressPolicyId": "ethereum-mainnet-create2",
  "chain": "ethereum",
  "network": "mainnet",
  "scheme": "create2",
  "minorUnit": "wei",
  "decimals": 18,
  "expectedAmountMinor": "15000000000000000",
  "customerReference": "order-1234",
  "address": "0x1234567890abcdef1234567890abcdef12345678"
}
```

## Backward compatibility (optional)

- API compatibility:
  - Existing payment-address request and status response shapes remain unchanged.
  - Existing Bitcoin endpoints remain unchanged.
- Data migration compatibility:
  - Existing Bitcoin rows and current payment status lookups must remain readable after the
    allocation-column cleanup.
  - No Ethereum backfill is required before rollout because Ethereum payment addresses do not exist
    yet.

## Failure modes and resiliency

- Retries/timeouts:
  - Receipt observation must be retriable without duplicate side effects.
  - JSON-RPC timeouts or transient scan failures should update row-level error state and retry
    later.
- Backpressure/limits:
  - Ethereum observation must use bounded block ranges and chunking rather than unbounded full-chain
    scans.
- Degradation strategy:
  - If Ethereum config is disabled or invalid, Ethereum policies remain non-issuable while Bitcoin
    behavior stays available.
  - If Go-side prediction cannot be trusted for the active contract metadata, Ethereum issuance
    should fail closed rather than issuing unverifiable addresses.
  - If privacy-preserving salt generation or persistence is unavailable, Ethereum issuance should
    fail closed rather than falling back to a publicly enumerable index-based derivation path.

## Observability

- Logs:
  - Allocation logs for Ethereum should include `paymentAddressId`, `addressPolicyId`, `chain`,
    `network`, and `address`, but should avoid raw salt or full source-ref material outside
    explicit debug-only workflows.
  - Receipt polling logs should include chain, network, address, scan range, and totals.
- Metrics:
  - Count issued Ethereum addresses and payment observation successes and failures.
- Traces:
  - Keep public API request traces separate from poller internal traces or cycle logs.
- Alerts:
  - Alert on repeated prediction mismatch, repeated observation failure, or ETH receipts older than
    the expected polling window.

## Security

- Authentication/authorization:
  - No public auth changes are required for customer-facing issuance or status APIs in this scope.
- Secrets:
  - Allocation-specific CREATE2 salt material is internal-only operational data and must not be
    exposed through customer-facing APIs, webhooks, or default logs.
- Abuse cases:
  - Anyone can send dust ETH to a predicted address; the system must tolerate unexpected inbound
    value.
  - A misconfigured collector address would route funds incorrectly, so startup validation and
    operator review are mandatory.
  - The receiver contract must not expose arbitrary target selection for sweeps.
  - A public preview or index-derived issuance path would leak future payment-address space and is
    therefore not acceptable for Ethereum privacy mode.

## Alternatives considered

- Option A:
  - Generate one EOA private key per payment address.
- Option B:
  - Use an HD-wallet-like Ethereum derivation scheme and keep sweeping from EOAs.
- Option C:
  - Use CREATE2-predicted payable receiver addresses even though post-funding deployment and sweep
    are deferred from the first rollout.
- Why chosen:
  - Option C preserves one-address-per-payment semantics without per-payment private-key custody and
    fits the current deterministic-address issuance model better than EOA management.

## Risks

- Risk:
  - Current issuance and persistence naming is still Bitcoin-biased, so a naive implementation can
    easily hide Ethereum behavior behind misleading field semantics.
  - Mitigation:
    - Clean up naming first and keep Ethereum issuance metadata explicit.
- Risk:
  - A publicly enumerable salt rule or public preview endpoint would let third parties derive
    future payment addresses for one Ethereum policy.
  - Mitigation:
    - Persist non-public allocation salts, keep preview-by-index disabled for Ethereum, and avoid
      exposing raw source-ref material in public surfaces.
- Risk:
  - Native ETH observation is harder than token log scanning because plain ETH transfers are not
    emitted as ERC-20 logs.
  - Mitigation:
    - Use block or transaction scanning with bounded ranges and contract tests against an explicit
      verification network.
- Risk:
  - A contract or prediction mismatch could strand funds at an address the runtime cannot deploy or
    sweep correctly.
  - Mitigation:
    - Add preflight checks, vector tests, local deployment smoke tests, and fail-closed startup
      validation.
