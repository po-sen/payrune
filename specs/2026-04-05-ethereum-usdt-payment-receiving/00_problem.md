---
doc: 00_problem
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

# Problem & Goals

## Context

- Background:
  - `payrune` already supports deterministic Bitcoin payment addresses and native ETH payment
    addresses backed by Ethereum CREATE2.
  - The current Ethereum implementation is explicitly native-asset-only: receipt observation reads
    ETH balance snapshots and the CREATE2 receiver contract can only sweep native ETH.
  - The current CREATE2 asset bundle still mixes version-suffixed artifact naming with
    implementation changes, which makes it harder to tell whether `V1` means a frozen historical
    deployment or simply the current checked-in build output.
  - The current sweep helper also treats the active metadata factory as the only valid recovery
    target, which makes old CREATE2 rows look stale after a factory redeploy even though each row
    already stores the factory address and init-code material it was issued with.
- Users or stakeholders:
  - Merchant backends that want to accept USDT on Ethereum with one payment address per checkout.
  - Operators who need an explicit, reviewable recovery path for ERC-20 balances held at predicted
    CREATE2 receiver addresses.
  - Backend developers extending the current multi-chain payment flow without collapsing the repo
    into speculative generic abstractions.
- Why now:
  - The next requested payment rail is USDT on Ethereum, not another Bitcoin variant.
  - The repo already has Ethereum issuance, polling, status, and sweep foundations; extending that
    path directly is cheaper and easier to review than introducing a second unrelated deposit flow.
  - Operator recovery for multiple USDT receivers currently requires multiple Ledger signatures,
    which is too error-prone for the intended workflow.
  - The repo still has no stable release tag or artifact-versioning system, so naming and recovery
    behavior should be simplified before mainnet rollout hardens accidental conventions.

## Constraints (optional)

- Technical constraints:
  - Preserve the existing Go layout and Clean Architecture boundaries under `cmd/` and `internal/`.
  - Keep the public chain identifier as `ethereum`; do not introduce a parallel `erc20` chain.
  - Keep the first rollout explicit to one ERC-20 asset family: USDT on Ethereum-backed policies.
  - Avoid generic multi-chain token registries or provider frameworks unless they solve a concrete
    need in this feature.
  - Keep operator recovery assets under `scripts/` and checked-in CREATE2 artifacts under
    `internal/infrastructure/ethereumcreate2assets`.
- Timeline/cost constraints:
  - Prefer extending the existing ETH CREATE2 flow over inventing a second issuance model.
- Compliance/security constraints:
  - No per-payment private keys may be generated or stored.
  - Ledger-only operator signing remains the required recovery stance.
  - Client-facing APIs and logs must not expose internal CREATE2 salt material.

## Problem statement

- Current pain:
  - `payrune` can issue `ethereum/*/create2` payment addresses for native ETH, but it cannot model,
    observe, or recover ERC-20 token payments such as USDT.
  - The current payment model treats `chain + network + address` as enough to identify what is
    being paid; that is insufficient once one Ethereum address can receive multiple assets.
  - The current receiver contract and sweep flow have no token-aware recovery path.
- Evidence or examples:
  - The current Ethereum observer only calls `eth_getBalance`.
  - The current `FixedCollectorReceiver` contract only exposes native ETH `receive()` and `sweep()`
    behavior.
  - Current persistence and API responses still expose redundant asset-shape fields rather than
    one canonical asset identifier.

## Goals

- G1:
  - Issue deterministic Ethereum payment addresses for USDT using the existing CREATE2 issuance
    model.
- G2:
  - Track USDT payment status through the current allocation, polling, status API, and webhook
    lifecycle.
- G3:
  - Make the payment model asset-aware enough that Ethereum native ETH and Ethereum USDT can
    coexist honestly without guessing asset identity from chain metadata alone.
- G4:
  - Extend the CREATE2 receiver and operator recovery flow so confirmed USDT balances can be swept
    safely to the configured collector through one unified receiver family and a stable factory
    interface.
- G6:
  - Remove unshipped legacy token-only receiver source/artifact files once the unified receiver
    cutover is complete so the checked-in asset bundle matches real runtime usage.
- G5:
  - Keep manual operator testing practical by supporting one-signature ERC-20 recovery batches and
    one small Ledger-signed USDT payment helper.
- G7:
  - Make recovery depend on row-owned CREATE2 material so old rows can still be swept after a
    factory redeploy on the same network.

## Non-goals (out of scope)

- NG1:
  - Generalized support for arbitrary ERC-20 tokens in the same rollout.
- NG2:
  - Support for non-Ethereum EVM chains such as Base, Arbitrum, or BSC.
- NG3:
  - Pending-transaction or mempool-only detection for ERC-20 transfers.
- NG4:
  - Automatic treasury accounting, exchange-rate handling, or payout settlement beyond payment
    receipt tracking and token recovery.
- NG5:
  - A generic asset abstraction that rewrites every chain path in one migration.
- NG6:
  - A generalized treasury console or wallet orchestration framework.

## Assumptions

- A1:
  - The first asset is USDT with `decimals=6`; runtime and public contracts do not need a separate
    token-specific `minorUnit` label.
- A2:
  - Runtime configuration provides the ERC-20 token contract address per enabled Ethereum network.
- A3:
  - Receipt polling can remain balance-snapshot-based for this asset family by reading
    ERC-20 `balanceOf(address)` at latest and confirmed blocks, consistent with the existing ETH
    polling model and current stop-polling behavior after `paid_confirmed`.
- A4:
  - Recovery remains operator-triggered and Ledger-signed; auto-sweep is not part of this change.
- A5:
  - Local/staging verification may use Tether's published Sepolia USD₮ test-token contract rather
    than a repo-maintained mock-token deployment flow.
- A6:
  - Address separation between Ethereum native ETH and Ethereum USDT can rely on
    `AddressPolicyID`-driven CREATE2 salt derivation instead of two different receiver bytecodes.
- A7:
  - The checked-in CREATE2 artifact file names can stay unversioned as long as deployed addresses
    and row-level recovery payloads remain the actual stable references.
- A8:
  - Because migration `000017` is still unshipped, it is acceptable to rewrite that migration in
    place instead of stacking a second corrective migration on top of draft asset metadata work,
    and the rewritten migration only needs to consume the schema baseline after `000016`.

## Open questions

- None. This spec fixes the first rollout to explicit Ethereum USDT support with one canonical
  `assetReference` contract and persistence changes.

## Success metrics

- Metric:
  - USDT issuance usability.
- Target:
  - `POST /v1/chains/ethereum/payment-addresses` can issue a deterministic USDT payment address
    through a USDT policy id without requiring new client-side CREATE2 parameters.
- Metric:
  - USDT receipt-state correctness.
- Target:
  - A funded USDT payment address reaches `paid_confirmed` through the existing poller and status
    API without affecting native ETH payment flows.
- Metric:
  - Asset identity correctness.
- Target:
  - Payment-status responses and webhook payloads expose `assetReference` so clients can
    distinguish Ethereum USDT from Ethereum native ETH without redundant shape fields.
- Metric:
  - Token recovery operability.
- Target:
  - Operators can recover confirmed USDT from selected CREATE2 payment addresses through the
    existing Ledger-based recovery workflow extended for ERC-20 balances.
