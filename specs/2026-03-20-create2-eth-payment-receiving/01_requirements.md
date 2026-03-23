---
doc: 01_requirements
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

# CREATE2 ETH Payment Receiving - Requirements

## Glossary (optional)

- CREATE2 payment address:
  - A deterministic Ethereum address predicted off-chain from a known factory address, salt, and
    receiver init code hash before the receiver contract is deployed.
- Factory contract:
  - The deployed contract that executes CREATE2 for receiver deployment. Its address is part of the
    CREATE2 preimage and therefore part of the deterministic payment-address space.
- Receiver contract:
  - The payable contract that will eventually be deployed at the predicted payment address and can
    sweep ETH to the configured collector destination.
- Collector address:
  - The operator-controlled destination wallet encoded into the receiver init code for this address
    space. It does not participate in first-rollout receipt polling.
- Operator signer:
  - A future runtime-controlled EOA, KMS-backed account, or equivalent sender that would pay gas to
    deploy the factory and later call factory or receiver collection transactions. It is not part
    of the CREATE2 address formula.
- Address source reference:
  - A canonical internal value that identifies the issuance source configuration used to allocate
    deterministic address slots. For Bitcoin this is expected to remain xpub-like material; for
    Ethereum CREATE2 this should reflect the active factory/init-code configuration rather than a
    fake public key and should exclude operator-signer identity.
- Address reference:
  - The canonical internal reference needed to reconstruct or reconcile one issued address. For
    Bitcoin this is an HD derivation path; for Ethereum CREATE2 this is expected to carry the
    CREATE2 salt or equivalent deterministic reference.

## Out-of-scope behaviors

- OOS1:
  - ERC-20 token receipt observation and collection.
- OOS2:
  - Generalized support for every EVM-compatible chain in the same iteration.
- OOS3:
  - Mempool subscriptions or pending-transaction payment detection.
- OOS4:
  - End-user or merchant-facing on-chain withdrawal controls from the receiver contract.

## Functional requirements

### FR-001 - Support Ethereum address policies and allocation through existing chain-scoped APIs

- Add first-class payment-address issuance support for `ethereum` using a CREATE2-backed address
  policy.
- Acceptance criteria:
  - [ ] `SupportedChain` accepts `ethereum`, while existing `bitcoin` behavior remains unchanged.
  - [ ] `GET /v1/chains/ethereum/address-policies` returns configured Ethereum policies for
        `mainnet` and `sepolia` when enabled, each with `scheme=create2`, `minorUnit=wei`, and
        `decimals=18`.
  - [ ] `POST /v1/chains/ethereum/payment-addresses` allocates one ETH payment address using the
        existing request body shape and returns the existing success payload shape with Ethereum
        values.
  - [ ] Public index-based address preview via `GET /v1/chains/ethereum/addresses` is rejected,
        disabled, or otherwise unavailable for privacy-preserving Ethereum CREATE2 policies, while
        the existing Bitcoin behavior remains unchanged.
  - [ ] Disabled or incomplete Ethereum policy configuration remains discoverable as disabled and
        is not issuable for either network.
- Notes:
  - The client should not need to know CREATE2 internals to request a payment address.

### FR-002 - Predict CREATE2 addresses deterministically and persist reconstructible metadata

- The system must derive the same ETH payment address off-chain that the on-chain factory will
  produce later.
- Acceptance criteria:
  - [ ] One allocation-specific CREATE2 salt or equivalent internal-only reference yields one
        deterministic CREATE2 payment address.
  - [ ] Go-side prediction matches Solidity or ABI-backed `computeAddress` vectors for the same
        factory, salt, collector behavior, and init code.
  - [ ] Ethereum CREATE2 salt derivation must not rely only on public metadata plus a sequential
        public index; the derivation input must include a runtime-managed non-public secret plus
        stable allocation identity so one issued address can be reconstructed without persisting a
        random one-off salt blob as the sole source of truth.
  - [ ] Allocation persistence stores a chain-agnostic `address_source_ref` equivalent and
        `address_reference` equivalent, rather than overloading Bitcoin-specific naming for
        Ethereum-issued rows.
  - [ ] The persisted metadata is sufficient to reconcile one issued ETH address and verify the
        expected CREATE2 preimage inputs later.
- Notes:
  - The exact salt strategy is a design decision, but it must stay stable for one allocation,
    testable, recoverable from runtime-managed secret material plus allocation metadata, and not
    make future addresses enumerable from public inputs alone.

### FR-003 - Observe native ETH receipts through the existing polling lifecycle

- Description:
  - Receipt polling must observe native ETH transfers to issued CREATE2 payment addresses and map
    them into the current payment receipt lifecycle.
- Acceptance criteria:
  - [ ] The poller can claim and process `ethereum` receipt-tracking rows without affecting the
        existing Bitcoin flows.
  - [ ] Deployment wiring can run one poller scope per `(chain, network)` so `ethereum/mainnet`
        and `ethereum/sepolia` do not rely on one mixed scheduled scope.
  - [ ] The Ethereum observer scans a bounded block range based on `issued_at`,
        `last_observed_block_height`, or both, rather than scanning the full chain every cycle.
  - [ ] The observer aggregates inbound native ETH value to the payment address in `wei`.
  - [ ] The observer distinguishes `confirmed_total_minor` from recently mined but not-yet-final
        value using the configured confirmation threshold.
  - [ ] Provider or scan failures persist row-level polling error state and schedule retry without
        breaking the rest of the cycle.
- Notes:
  - In this iteration, “unconfirmed” may be limited to mined transfers below the confirmation
    threshold rather than mempool observations.
  - In this iteration, the observer may limit itself to canonical block transactions whose
    destination address matches the issued payment address; trace-based internal ETH transfers are
    out of scope unless the configured provider and implementation explicitly support them.

### FR-004 - Keep payment status retrieval and webhook behavior chain-consistent for Ethereum

- Description:
  - The existing payment-status and webhook flow must remain the client-facing source of truth for
    ETH payment progress.
- Acceptance criteria:
  - [ ] `GET /v1/chains/ethereum/payment-addresses/{paymentAddressId}` returns the latest
        persisted payment state using the existing response shape.
  - [ ] Existing payment-receipt status transitions remain valid for Ethereum rows.
  - [ ] Existing webhook notification payload shape remains unchanged, except for Ethereum-specific
        data values such as `chain`, `network`, `address`, and amount totals.
  - [ ] No new mandatory public API endpoint is required for a client to issue, poll, or receive
        webhook updates for an ETH payment.
- Notes:
  - No additional public collection control surface is required in this iteration.

### FR-005 - Validate and bootstrap Ethereum CREATE2 runtime configuration explicitly

- Description:
  - Ethereum issuance and observation must start only when all required network configuration is
    present and internally consistent.
- Acceptance criteria:
  - [ ] Runtime configuration includes Ethereum RPC endpoint, collector address, derivation-key
        configuration, receipt confirmation threshold, and receipt expiry settings.
  - [ ] Deployment-facing Compose and Cloudflare defaults may provide overrideable public
        Ethereum JSON-RPC endpoints per network for local or bootstrap convenience, while still
        allowing operators to replace them explicitly in production.
  - [ ] Checked-in deployment metadata provides the factory address per network, and checked-in
        receiver contract artifacts provide bytecode or a derivable init code hash for prediction.
  - [ ] Before T-003 lands real deployment artifacts, local API testing may rely on checked-in
        deterministic fixture metadata so configured Ethereum policies remain issuable in non-prod
        workflows.
  - [ ] Startup fails fast when Ethereum addresses, hashes, or config combinations are invalid.
  - [ ] Deployment-facing config examples expose separate `ETHEREUM_MAINNET_CREATE2_COLLECTOR_ADDRESS`
        and `ETHEREUM_SEPOLIA_CREATE2_COLLECTOR_ADDRESS` settings instead of hand-entered env vars
        for factory addresses or init code hashes.
  - [ ] Contract verification tooling can run against an explicitly configured Ethereum RPC
        network using operator-provided signer credentials, without requiring repo-managed devnet
        infrastructure.
  - [ ] Important CREATE2 contract tooling is maintained as Go CLI entry points under `cmd/`,
        while any helper shell wrappers used for setup or orchestration live under `scripts/`.
- Notes:
  - Configuration should be explicit and network-scoped rather than hidden behind indirect prefix
    logic.

### FR-006 - Keep receiver behavior safe by construction

- Description:
  - The CREATE2 receiver design must minimize custody and routing risk.
- Acceptance criteria:
  - [ ] No per-payment private key material is generated or stored.
  - [ ] The receiver contract forwards ETH only to the configured collector destination and does
        not expose a generic arbitrary-call surface.
  - [ ] The runtime verifies that Go-side prediction inputs and deployed contract bytecode
        expectations match before issuing payment addresses or running explicit contract
        verification on a network.
- Notes:
  - The goal is operational safety, not a generic wallet contract platform.

### FR-007 - Preserve privacy of future Ethereum payment-address issuance

- Description:
  - Ethereum CREATE2 issuance must avoid exposing enough public information for third parties to
    precompute or enumerate future payment addresses for one active address space. Checked-in
    factory metadata may remain public; the privacy requirement applies to the full combination of
    public metadata, public API behavior, and salt derivation rules.
- Acceptance criteria:
  - [ ] Checked-in factory metadata plus public API inputs are insufficient to derive future
        payment addresses without internal allocation-only salt material.
  - [ ] Public customer-facing APIs, webhooks, and OpenAPI examples do not expose raw CREATE2
        salts, full `address_source_ref`, or collector-derived preimage inputs.
  - [ ] Default operational logs avoid emitting raw CREATE2 salts or full source references; use
        `paymentAddressId`, `chain`, `network`, and payment address instead.
  - [ ] The design explicitly documents that v1 privacy protects against address-space enumeration
        before settlement, not guaranteed anonymity after final on-chain collection to a known
        treasury address.
- Notes:
  - Privacy here is about preventing easy precomputation and mass linkage of future addresses, not
    about hiding publicly visible blockchain transactions after settlement. Public metadata alone is
    acceptable if it is not sufficient to enumerate future addresses.

## Non-functional requirements

- Performance (NFR-001):
  - Payment-address allocation for an enabled Ethereum policy must remain an in-process operation
    with p95 latency <= 250 ms in a warm local environment because CREATE2 address prediction does
    not require chain IO.
- Availability/Reliability (NFR-002):
  - Re-running issuance or polling for the same Ethereum payment address must be safe and must not
    create duplicate receipt rows or inconsistent payment status transitions.
- Security/Privacy (NFR-003):
  - No per-payment secret keys may be persisted. For privacy-preserving Ethereum issuance, each
    allocation must use at least 128 bits of non-public salt entropy or an equivalent non-public
    derivation secret so future addresses are not enumerable from public inputs alone.
- Compliance (NFR-004):
  - No additional compliance controls are introduced in this iteration beyond existing payment
    auditability and deterministic state persistence.
- Observability (NFR-005):
  - Logs and persisted state must let an operator diagnose prediction mismatch and observation
    failure using `paymentAddressId`, chain, network, address, and block-height context, without
    requiring raw CREATE2 salt or full source-ref material in default logs.
- Maintainability (NFR-006):
  - EVM-specific RPC, ABI, and contract details must stay confined to adapters or infrastructure,
    and the existing Bitcoin tests and flows must remain green after the Ethereum changes.

## Dependencies and integrations

- External systems:
  - Ethereum JSON-RPC provider.
  - CREATE2 factory and receiver contract artifacts plus deployment flow.
- Internal services:
  - Existing payment-address allocation flow.
  - Existing payment receipt polling and status API flow.
  - Existing payment-receipt webhook dispatcher flow.
