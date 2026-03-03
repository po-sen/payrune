---
doc: 00_problem
spec_date: 2026-03-03
slug: btc-xpub-address-api
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-03-swagger-ui-container-api-testing
  - 2026-03-03-cmd-app-compose-prefix
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background: We need chain-extensible address APIs that derive bitcoin payment addresses from xpub while keeping deployment opt-in by compose overrides.
- Users or stakeholders: Backend developers and integrators issuing payment addresses for testing and local development.
- Why now: Endpoint design should scale beyond legacy-only route handling and support multiple address schemes in one chain-scoped API contract.

## Constraints (optional)

- Technical constraints:
  - Use `compose.bitcoin.mainnet.yaml` and `compose.bitcoin.testnet4.yaml` to inject xpub for bitcoin policy variants.
  - Base `compose.yaml` must keep bitcoin policy disabled by default.
  - Address API should use extensible path prefix `/v1/chains/{chain}`.
  - Supported bitcoin scheme values are `legacy`, `segwit`, `nativeSegwit`, and `taproot`.
  - API security controls can be deferred.
- Timeline/cost constraints:
  - Keep first iteration simple and deterministic for local usage.
- Compliance/security constraints:
  - Keep spec-first workflow and preserve Clean Architecture + Hexagonal boundaries.

## Problem statement

- Current pain:
  - Bitcoin address route and request shape were legacy-only and not chain-extensible.
  - Caller needed low-level network selection instead of stable policy identifiers.
  - No API was provided to list available/active address policies.
- Evidence or examples:
  - Existing API only had health endpoint plus bitcoin network/index derivation path.

## Goals

- G1: Introduce chain-scoped endpoints: `/v1/chains/{chain}/address-policies` and `/v1/chains/{chain}/addresses`.
- G2: Derive addresses by `addressPolicyId` to decouple callers from derivation internals.
- G3: Keep compose override based enablement for mainnet/testnet4 bitcoin policies.
- G4: Support bitcoin address schemes: `legacy`, `segwit`, `nativeSegwit`, and `taproot`.
- G5: Include `minorUnit` and `decimals` metadata in policy and address responses.
- G6: Support multiple compose override files through one `COMPOSE_OVERRIDE` input.

## Non-goals (out of scope)

- NG1: Authentication/authorization or rate limiting for address endpoints.
- NG2: Wallet state management, gap-limit scanning, or address reservation workflows.
- NG3: Full multi-chain derivation implementation beyond bitcoin in this iteration.

## Assumptions

- A1: xpub values supplied through compose overrides are non-hardened derivable keys.
- A2: Policy catalog is environment-driven; callers use `addressPolicyId` and do not provide derivation internals.

## Open questions

- Q1: Should future versions expose richer policy metadata (purpose/coin_type/script variants) in list response?
- Q2: Should unsupported chain responses move to a standard machine-readable error code model?

## Success metrics

- Metric: Chain-scoped API availability.
- Target: `GET /v1/chains/bitcoin/address-policies` and `GET /v1/chains/bitcoin/addresses` return expected payloads.
- Metric: Policy-driven derivation.
- Target: Address generation works with `addressPolicyId` + `index` and rejects disabled policies with explicit status.
- Metric: Address scheme coverage.
- Target: `legacy`, `segwit`, `nativeSegwit`, and `taproot` policies are listed and derivable when corresponding xpub is configured.
- Metric: Override enablement.
- Target: `compose.bitcoin.mainnet.yaml` and `compose.bitcoin.testnet4.yaml` independently enable matching bitcoin policies.
- Metric: Amount metadata availability.
- Target: Policy list and address generation payloads include `minorUnit` and `decimals`.
- Metric: Compose ergonomics.
- Target: `COMPOSE_OVERRIDE` supports multiple files (space or comma separated).
