---
doc: 00_problem
spec_date: 2026-03-20
slug: create2-eth-payment-receiving
mode: Full
status: READY
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

# CREATE2 ETH Payment Receiving - Problem & Goals

## Context

- Background:
  - `payrune` already allocates deterministic payment addresses, persists payment-address lifecycle
    state, polls blockchain receipts, exposes status APIs, and dispatches webhook notifications.
  - The current implementation is structurally multi-chain but materially Bitcoin-first. Public chain
    support is still limited to `bitcoin`, and the active issuance model is biased toward
    `account_public_key` plus `derivation_path`.
- Users or stakeholders:
  - Merchant backends that need one payment address per checkout or invoice.
  - Backend developers extending `payrune` to support non-Bitcoin chains.
  - Operators who must safely manage collection wallets, deployer keys, and settlement flows.
- Why now:
  - The next planned payment rail is native ETH.
  - CREATE2 allows deterministic deposit addresses without generating or storing one private key per
    payment address.
  - The repo already has chain-scoped APIs and polling/status infrastructure that can be extended if
    the address-issuance model is cleaned up first.

## Constraints (optional)

- Technical constraints:
  - Preserve the existing Go project layout and Clean Architecture boundaries under `internal/`.
  - Reuse the current `POST /v1/chains/{chain}/payment-addresses` and
    `GET /v1/chains/{chain}/payment-addresses/{paymentAddressId}` API shape where possible.
  - Avoid pushing EVM SDK types, ABI details, or RPC response structs into domain or application
    packages.
  - CREATE2 prediction must be reproducible off-chain and must match the deployed on-chain address
    exactly.
- Timeline/cost constraints:
  - Prefer a staged implementation that lands a usable ETH payment flow without rebuilding the whole
    payment system.
- Compliance/security constraints:
  - No per-payment private key material may be generated or persisted.
  - Deployer/signer credentials must stay operator-managed.
  - The collection path must prevent sweeping ETH to arbitrary destinations.

## Problem statement

- Current pain:
  - `payrune` cannot issue ETH payment addresses today.
  - The current issuance configuration and persistence naming (`account_public_key`,
    `derivation_path`) describe HD-wallet derivation, not CREATE2-based address prediction.
  - The current receipt observer stack is Bitcoin-specific and does not yet observe native ETH
    transfers or drive ETH-specific settlement actions.
- Evidence or examples:
  - `SupportedChain` only accepts `bitcoin`.
  - The current allocation flow assumes address derivation inputs that map to xpub-based issuance.
  - Receipt polling ports are chain-routed, but there is no Ethereum observer adapter wired in.

## Goals

- G1:
  - Issue one deterministic ETH payment address per allocation request using CREATE2 without
    generating one wallet keypair per payment address.
- G2:
  - Reuse the existing payment-address allocation, receipt tracking, status API, and webhook flow
    for Ethereum-native payments.
- G3:
  - Support an idempotent post-funding deployment and sweep path so ETH received at a predicted
    CREATE2 address can be forwarded to the operator collector wallet.
- G4:
  - Clean up the issuance model so Bitcoin HD derivation and Ethereum CREATE2 issuance can coexist
    without misusing Bitcoin-specific field names.

## Non-goals (out of scope)

- NG1:
  - ERC-20 token deposits in the first iteration.
- NG2:
  - Multi-EVM-chain support beyond the first selected Ethereum network(s).
- NG3:
  - Mempool-only payment detection before a transaction is mined.
- NG4:
  - Merchant payout settlement, treasury accounting, or fiat reconciliation beyond confirming and
    collecting ETH into the configured collector address.

## Assumptions

- A1:
  - The first delivery targets native ETH only and uses `wei` as the persisted minor unit with
    `18` decimals.
- A2:
  - One CREATE2 factory contract and one collector destination are configured per Ethereum network.
- A3:
  - A payment address may receive ETH before code is deployed at that address; deployment and
    sweep happen after payment detection or settlement selection.
- A4:
  - The public chain identifier should be `ethereum` in v1 to match the explicit style used by
    `bitcoin`.
- A5:
  - In v1, “unconfirmed” for Ethereum means mined but below the configured confirmation threshold;
    unmined mempool transfers are not required.

## Open questions

- Q1:
  - Should the first production scope be `ethereum/mainnet` only, or should `sepolia` also be
    supported as a first-class configured network?
- Q2:
  - Should deploy-and-sweep run in a dedicated worker, be triggered from the existing poller
    lifecycle, or remain a manual operator command in the first rollout?
- Q3:
  - Should overpayment above `expectedAmountMinor` be swept automatically in the first version or
    be left for manual review when it exceeds a threshold?
- Q4:
  - Do we want the sweeper to run after `paid_unconfirmed` or only after `paid_confirmed`?

## Success metrics

- Metric:
  - CREATE2 prediction correctness.
  - Target:
    - Go-side predicted address matches the deployed on-chain CREATE2 address in 100% of test
      vectors and local deployment smoke tests.
- Metric:
  - ETH payment issuance usability.
  - Target:
    - `POST /v1/chains/ethereum/payment-addresses` returns deterministic addresses without private
      key material and without extra client parameters beyond the existing allocation payload.
- Metric:
  - Receipt-state correctness.
  - Target:
    - A mined ETH transfer at or above the expected amount transitions the payment to the correct
      status within two poll cycles in local and staging environments.
- Metric:
  - Collection reliability.
  - Target:
    - Re-running deploy/sweep for the same funded address does not duplicate collection and
      produces a deterministic already-complete or already-in-progress result.
