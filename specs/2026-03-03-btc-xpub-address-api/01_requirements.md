---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Address policy: A stable derivation profile identified by `addressPolicyId`.
- Enabled policy: A policy with non-empty xpub configured at runtime.
- Scheme: Bitcoin address encoding strategy (`legacy`, `segwit`, `nativeSegwit`, `taproot`).

## Out-of-scope behaviors

- OOS1: API authentication, authorization, and abuse protection.
- OOS2: Persistent tracking of issued addresses.
- OOS3: Non-bitcoin derivation implementation.

## Functional requirements

### FR-001 - Bitcoin compose override files

- Description:
  - Deployment configuration MUST provide optional compose override files for bitcoin mainnet and testnet4 scheme-specific xpub injection.
- Acceptance criteria:
  - [x] `deployments/compose/compose.bitcoin.mainnet.yaml` exists and configures `BITCOIN_MAINNET_LEGACY_XPUB`, `BITCOIN_MAINNET_SEGWIT_XPUB`, `BITCOIN_MAINNET_NATIVE_SEGWIT_XPUB`, `BITCOIN_MAINNET_TAPROOT_XPUB` for `app` service.
  - [x] `deployments/compose/compose.bitcoin.testnet4.yaml` exists and configures `BITCOIN_TESTNET4_LEGACY_XPUB`, `BITCOIN_TESTNET4_SEGWIT_XPUB`, `BITCOIN_TESTNET4_NATIVE_SEGWIT_XPUB`, `BITCOIN_TESTNET4_TAPROOT_XPUB` for `app` service.
  - [x] Base `deployments/compose/compose.yaml` does not require bitcoin xpub variables.
- Notes:
  - Enabling each bitcoin policy is controlled by corresponding xpub env var presence.

### FR-002 - Multi-file COMPOSE_OVERRIDE support

- Description:
  - `Makefile` MUST support multiple compose override files via one `COMPOSE_OVERRIDE` input.
- Acceptance criteria:
  - [x] `COMPOSE_OVERRIDE` accepts space-separated file list.
  - [x] `COMPOSE_OVERRIDE` accepts comma-separated file list.
  - [x] Rendered compose command includes all override files in order.
- Notes:
  - This is a developer ergonomics requirement.

### FR-003 - Chain-scoped address policy listing endpoint

- Description:
  - Service MUST expose `GET /v1/chains/{chain}/address-policies` to list policies for the chain.
- Acceptance criteria:
  - [x] Endpoint returns `200` with `chain` and `addressPolicies[]` payload.
  - [x] Each policy includes `addressPolicyId`, `chain`, `network`, `scheme`, `minorUnit`, `decimals`, and `enabled`.
  - [x] `scheme` is one of `legacy`, `segwit`, `nativeSegwit`, `taproot`.
  - [x] Unsupported chain path returns `404`.
- Notes:
  - For this iteration, `chain=bitcoin` is implemented.

### FR-004 - Chain-scoped address generation endpoint

- Description:
  - Service MUST expose `GET /v1/chains/{chain}/addresses` and derive by `addressPolicyId` and `index`.
- Acceptance criteria:
  - [x] Endpoint accepts query parameters `addressPolicyId` and `index`.
  - [x] Successful response returns `addressPolicyId`, `chain`, `network`, `scheme`, `minorUnit`, `decimals`, `index`, and `address`.
  - [x] Address derivation supports `legacy`, `segwit`, `nativeSegwit`, and `taproot` policies.
  - [x] For account-level xpub (depth <= 3), derivation uses external chain branch `/0/index`.
  - [x] For change-level-or-deeper xpub (depth >= 4), derivation uses direct `index` from provided xpub.
  - [x] Endpoint rejects unsupported HTTP methods with `405` and `Allow: GET`.
- Notes:
  - API keeps derivation internals out of caller contract while enforcing deterministic path semantics by xpub depth.

### FR-005 - Policy and error semantics

- Description:
  - Address generation behavior MUST be policy-aware and return explicit status classes.
- Acceptance criteria:
  - [x] Unknown `addressPolicyId` returns `400`.
  - [x] Disabled (missing xpub) policy returns `501`.
  - [x] Invalid `index` or missing `addressPolicyId` returns `400`.
  - [x] Unsupported chain returns `404`.
- Notes:
  - Existing chain support is explicit instead of implicit fallback.

### FR-006 - OpenAPI contract update

- Description:
  - OpenAPI spec MUST describe chain-scoped policy listing and generation endpoints.
- Acceptance criteria:
  - [x] `deployments/swagger/openapi.yaml` contains both `/v1/chains/{chain}/address-policies` and `/v1/chains/{chain}/addresses` operations.
  - [x] Request parameters and response schemas align with runtime JSON.
  - [x] Server URL remains `http://localhost:8080`.
  - [x] Swagger compose service mounts `deployments/swagger/` as a directory and serves `/specs/openapi.yaml` to avoid stale single-file bind-mount inode issues.
- Notes:
  - Swagger UI remains on `http://localhost:8081`.

### FR-007 - Bitcoin outbound adapter test coverage

- Description:
  - Bitcoin outbound derivation adapter MUST provide unit tests that verify scheme behavior and deterministic derivation.
- Acceptance criteria:
  - [x] Tests cover `legacy`, `segwit`, `nativeSegwit`, and `taproot` for mainnet and testnet4.
  - [x] Tests assert scheme-correct address type (P2PKH/P2SH-P2WPKH/P2WPKH/P2TR).
  - [x] Tests verify deterministic output for the same `(network, scheme, xpub, index)` input.
  - [x] Tests include unsupported network and unsupported scheme failure paths.

## Non-functional requirements

- Performance (NFR-001): Single derivation request should complete within 300ms on local development environment.
- Availability/Reliability (NFR-002): Missing xpub should not crash service startup; policy should be listed as disabled.
- Security/Privacy (NFR-003): No private keys are introduced; only xpub values are consumed.
- Compliance (NFR-004): `SPEC_DIR="specs/2026-03-03-btc-xpub-address-api" bash scripts/spec-lint.sh` passes.
- Observability (NFR-005): Error responses remain explicit (`400`, `404`, `405`, `501`, `500`) for local debugging.
- Maintainability (NFR-006): Implementation follows existing Clean Architecture layering (ports/use_case/controller/adapter separation).

## Dependencies and integrations

- External systems:
  - Bitcoin derivation library for BIP32 xpub parsing and child derivation.
- Internal services:
  - Existing app HTTP server bootstrap and DI container.
  - Existing Swagger UI compose service and OpenAPI file.
