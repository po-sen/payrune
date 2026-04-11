---
doc: 00_problem
spec_date: 2026-04-07
slug: ethereum-contract-readiness
mode: Full
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-05-ethereum-usdt-payment-receiving
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
  - `payrune` now supports Ethereum CREATE2 payment-address issuance for native ETH and USDT.
  - The API process currently validates Ethereum policy config only at the string-format level.
  - The current `enabled` flag is derived from partially overlapping conditions, which makes it easy to confuse operator intent, static config completeness, and startup-time on-chain readiness.
  - Address generation and allocation still derive Ethereum addresses without proving that the on-chain factory or token contract is actually deployed and compatible with the checked-in runtime expectations.
  - The poller already knows how to build Ethereum RPC clients, but the API process does not currently load Ethereum RPC config or perform any on-chain readiness checks before issuing addresses.
- Users or stakeholders:
  - Operators who want issuance to fail closed when Ethereum contract configuration drifts or a deployment is broken.
  - Merchant backends that would rather receive a temporary API failure than issue a payment address that can never be observed or recovered safely.
- Developers maintaining the current Ethereum CREATE2 flow who need an explicit, reviewable readiness contract instead of implicit trust in env values.
  - Developers maintaining the policy catalog who need `enabled` to mean static config completeness instead of a half-step before readiness.
- Why now:
  - The repo just added Ethereum USDT support, which made the gap more obvious: one wrong factory or token contract can now break observation and recovery after an address is already issued.
  - The user explicitly wants Ethereum address generation blocked unless the relevant contracts are confirmed healthy first.
  - This is easier to add now while Ethereum issuance logic is still fresh and the new spec can depend directly on the completed USDT rollout.

## Constraints (optional)

- Technical constraints:
  - Preserve the current Go layout and Clean Architecture boundaries.
  - Keep the first readiness feature explicit to Ethereum issuance; do not invent a generic multi-chain contract health framework.
  - Reuse the existing Ethereum RPC env contract already used by the poller where possible.
  - Avoid adding new persistent state or background caches unless they solve a concrete issue in this feature.
- Timeline/cost constraints:
  - Prefer a fail-closed read path at issuance time over a broader long-lived health subsystem.
- Compliance/security constraints:
  - No issuance path may bypass readiness checks for enabled Ethereum policies.
  - Validation failures must not expose raw CREATE2 salts or other sensitive recovery material.

## Problem statement

- Current pain:
  - `generate-address` and `allocate-payment-address` can return Ethereum addresses even when the configured factory or token contract is missing, undeployed, or incompatible with the runtime contract assumptions.
  - The current policy-enable path makes it hard to tell whether a policy is intentionally disabled, accidentally incomplete, or blocked by startup readiness.
  - Bootstrap catches malformed Ethereum asset-reference strings, but it does not confirm on-chain code or ABI compatibility.
  - This means the system can hand out a payment address first and only discover the broken contract later during polling or recovery.
- Evidence or examples:
  - API bootstrap only calls `validateConfiguredEnabledEthereumAddressIssuancePolicy`, which currently checks `assetReference` shape but performs no chain RPC validation.
  - The API process does not currently construct Ethereum RPC clients, while the poller does.
  - The CREATE2 sweep helper already validates factory code and token balance calls at operator time, showing the repo already considers on-chain contract readiness operationally important.

## Goals

- G1:
  - Block Ethereum address preview and address allocation when the relevant on-chain contracts are not ready.
- G2:
  - Keep the readiness check explicit and deterministic: factory code must match the checked-in factory artifact, and Ethereum token policies must confirm token-contract availability plus required read calls.
- G3:
  - Reuse the existing Ethereum RPC configuration model so operators do not need a second set of API-only RPC env names.
- G4:
  - Fail API startup loudly and with operator-meaningful diagnostics when enabled Ethereum issuance readiness cannot be established.
- G5:
  - Make policy state explicit and simple: one operator-controlled `ENABLED` flag decides intent, static config validation confirms completeness, and startup readiness remains a separate gate.

## Non-goals (out of scope)

- NG1:
  - Generic multi-chain contract validation across Bitcoin or future chains.
- NG2:
  - Continuous background health monitoring or caching dashboards for Ethereum contracts.
- NG3:
  - Automatic contract remediation or redeployment.
- NG4:
  - Deep semantic verification of every ERC-20 write behavior beyond the read-side compatibility this runtime actually depends on.

## Assumptions

- A1:
  - For Ethereum CREATE2 issuance, the minimum safe readiness contract is the checked-in factory runtime code plus the policy-specific ERC-20 read contract when `assetReference` is non-empty.
- A2:
  - Native ETH policies do not need a token-contract check, but they still require the configured factory to be present and compatible.
- A3:
  - USDT-like token policies can be treated as ready when the asset-reference contract has code, `balanceOf(address)` is callable, and `decimals()` matches the policy decimals.
- A4:
  - Failing API startup for issuance-readiness failures is acceptable because the problem is operational and should be corrected before the API serves Ethereum issuance traffic.
- A5:
  - Disabled policies should be ignored completely by static-config validation and startup readiness.
  - If a policy is explicitly enabled but required static config is missing or malformed, startup should fail instead of silently disabling it.

## Open questions

- None. This spec fixes the first rollout to explicit Ethereum issuance readiness checks in the API process.

## Success metrics

- Metric:
  - Fail-closed Ethereum issuance.
- Target:
  - `POST /v1/chains/ethereum/payment-addresses` rejects enabled Ethereum policies when on-chain readiness checks fail, and `GET /v1/chains/ethereum/generate-address` keeps the readiness hook for any future Ethereum preview-capable policy while current CREATE2 preview stays unsupported.
- Metric:
  - Operator clarity.
- Target:
  - Readiness failures produce one clear startup error that distinguishes contract-readiness problems from generic bootstrap failures.
- Metric:
  - No regression to non-Ethereum issuance.
- Target:
  - Bitcoin issuance and status flows continue to work without requiring Ethereum RPC configuration.
