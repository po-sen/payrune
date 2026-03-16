---
doc: 00_problem
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

# Problem & Goals

## Context

- Background:
  - The current payment-address flow is built around Bitcoin xpub derivation. Allocation reserves a
    deterministic derivation index, derives an address from xpub metadata, persists the derivation
    path, and then polls the chain by `(chain, address)`.
  - The next product step is to support Ethereum deposits on both production and test networks while
    keeping the same high-level API shape: issue one payment address per order, track payment
    status, and later let the operator sweep funds into one collector account on demand.
  - For Ethereum, the operator does not want to manage one private key per deposit address. The
    desired model is one operator account plus a contract factory that can generate many deposit
    addresses and later collect funds in batches when the operator chooses to run a sweep command.
- Users or stakeholders:
  - payrune maintainers implementing chain support.
  - Wallet operators who want manual control over treasury collection timing.
  - Product/API consumers that expect one unique deposit address per payment.
- Why now:
  - The existing architecture is chain-scoped already, but its issuance and observation ports are
    still Bitcoin-shaped.
  - Adding Ethereum now requires a deliberate design before code, otherwise the codebase will end up
    with xpub-specific abstractions leaking into an EVM flow that has different constraints.

## Constraints (optional)

- Technical constraints:
  - Preserve the repo's `cmd/` + `internal/` layout and Clean Architecture boundaries.
  - Do not introduce a new top-level architecture folder for EVM logic; keep chain-specific code
    under existing `internal/` structure.
  - Support both mainnet and an application-oriented Ethereum testnet. As of 2026-03-15, the
    Ethereum docs list Sepolia as the recommended default testnet for application development and
    mark Holesky as deprecated.
  - The design must support ETH and USDT on Ethereum mainnet, and an operator-configured USDT-like
    ERC20 on Sepolia for testing.
- Timeline/cost constraints:
  - Payment address issuance should stay cheap and fast. A per-payment on-chain deployment step is
    undesirable because it adds latency, gas cost, and external-IO failure modes to the address
    issuance API.
- Compliance/security constraints:
  - No per-deposit private keys should be stored in application config or database.
  - Sweep signing material must stay isolated to dedicated relayer/operator credentials.

## Problem statement

- Current pain:
  - The current core model assumes address issuance means `xpub + derivation_path_prefix + index ->
address`. Ethereum deposit addresses do not need that model if we use a contract factory and
    CREATE2-style deterministic vault addresses.
  - The current receipt observer model is also insufficient for Ethereum assets because it only
    knows `(chain, network, address)` and has no way to distinguish native ETH from an ERC20 such as
    USDT.
  - The current persistence is xpub-centric (`xpub_fingerprint`, `derivation_index`,
    `derivation_path`) and does not store EVM-specific issuance metadata such as factory address,
    collector address, or CREATE2 salt.
- Evidence or examples:
  - `internal/application/ports/outbound/chain_address_deriver.go` requires
    `AccountPublicKey`, `DerivationPathPrefix`, and `Index`.
  - `internal/domain/valueobjects/supported_chain.go` only recognizes `bitcoin`.
  - `internal/application/ports/outbound/blockchain_receipt_observer.go` tracks by chain/network
    and address only, which is not enough for token-specific observation on Ethereum.

## Goals

- G1:
  - Support `ethereum` as a first-class chain with `mainnet` and `sepolia` payment policies.
- G2:
  - Issue one unique deposit address per payment for both ETH and USDT without managing one private
    key per address.
- G3:
  - Enable on-demand batch sweep of many issued deposit addresses into one collector account through
    a smart contract factory plus deposit-vault pattern.
- G4:
  - Extend payment tracking so ETH and ERC20 deposits can be observed, confirmed, and reported
    through the existing payment-status lifecycle.
- G5:
  - Keep Bitcoin support intact while refactoring xpub-centric issuance into a more general
    chain-specific issuance model.

## Non-goals (out of scope)

- NG1:
  - Adding non-Ethereum EVM chains such as Arbitrum, Optimism, BSC, or Polygon in this iteration.
- NG2:
  - Automatic conversion, swap, or treasury routing after a manual sweep.
- NG3:
  - An always-on background Ethereum sweeper worker.
- NG4:
  - Supporting every ERC20 token generically in the first slice; this scope is ETH plus USDT-like
    ERC20 only.
- NG5:
  - Replacing the current HTTP contract family with a wholly new API surface.

## Assumptions

- A1:
  - Sepolia is the correct Ethereum application testnet to target for this feature on
    2026-03-15.
- A2:
  - The Ethereum mainnet USDT policy will use Tether's published ERC20 contract address
    `0xdAC17F958D2ee523a2206206994597C13D831ec7`.
- A3:
  - A public, official Tether Sepolia USDT contract is not relied on. The Sepolia USDT policy will
    therefore use an operator-configured ERC20 test token with 6 decimals.
- A4:
  - Address issuance should be counterfactual: the application predicts a unique CREATE2/clone
    address and persists its salt immediately, while on-chain deployment can happen later when the
    operator runs a sweep command.
- A5:
  - Because Ethereum JSON-RPC does not efficiently enumerate native incoming transfers by address,
    the EVM receipt observer will need an address-indexed explorer/indexer provider rather than raw
    JSON-RPC alone.

## Open questions

- Q1:
  - None. The design assumes Sepolia plus an operator-configured USDT-like test token and is ready
    for implementation on that basis.

## Success metrics

- Metric:
  - Payment-address issuance for Ethereum returns a unique address and stable idempotent replay
    behavior.
- Target:
  - Repeated requests with the same idempotency key return the same `paymentAddressId` and address,
    while distinct requests return unique Ethereum addresses.
- Metric:
  - Operator can sweep multiple confirmed ETH/USDT deposits into one collector account.
- Target:
  - Contract and command tests demonstrate successful batch sweep for both native ETH and an
    ERC20-compatible USDT-like token.
- Metric:
  - Receipt tracking correctly distinguishes ETH and ERC20 deposits.
- Target:
  - Poller integration tests pass for Ethereum mainnet/Sepolia policy snapshots and confirm
    `watching -> paid_unconfirmed -> paid_confirmed` transitions with asset-aware observation.
