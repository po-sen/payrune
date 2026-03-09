---
doc: 00_problem
spec_date: 2026-03-09
slug: compose-mainnet-test-files
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-09-bitcoin-compose-defaults
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Problem & Goals

## Context

- Background:
  - Compose deployment shape should match actual operational entrypoints, with one production base file and one local/test override.
- Users or stakeholders:
  - Operators and developers who need one production-like Compose file and one local/test override file.
- Why now:
  - The desired end state is explicit: `compose.yaml` is the production-like deployment file, and `compose.test.yaml` only overrides it for local/test use.

## Constraints (optional)

- Technical constraints:
  - Keep the existing Go services and current Dockerfiles unchanged.
  - Remove separate Compose files for Bitcoin network overlays and optional tooling overlays.
  - Preserve the committed fake env file for local/test Compose rendering where base required env still apply.
- Timeline/cost constraints:
  - Prefer concrete, easy-to-run Compose entrypoints over DRY-looking file composition.
- Compliance/security constraints:
  - Production-like required env must stay explicit in `compose.yaml`.

## Problem statement

- Current pain:
  - Separate overlay files (`compose.bitcoin.*`, `compose.swagger.yaml`, `compose.webhook.yaml`) make deployment harder to reason about.
  - A minimal base stack is not useful if it excludes services required for the product to function.
  - `compose.test.yaml` should override the production-like base, not behave like a second independent deployment entrypoint.
- Evidence or examples:
  - The product needs more than `app` to be operational; receipt polling and webhook dispatch are part of the actual service behavior.
  - The desired deployment model is now one production-like file plus one test/local override file.

## Goals

- G1:
  - Make `deployments/compose/compose.yaml` the single production-like mainnet stack.
- G2:
  - Make `deployments/compose/compose.test.yaml` the single local/test override with testnet4, Swagger, and fake webhook receiver.
- G3:
  - Remove separate Compose files for Bitcoin network overlays and optional tooling overlays.
- G4:
  - Keep all relevant env keys explicitly listed in the two remaining Compose files.

## Non-goals (out of scope)

- NG1:
  - Cloudflare, ingress, or TLS deployment changes.
- NG2:
  - Business-rule changes to payment allocation or receipt tracking.
- NG3:
  - Reworking Docker images or service binaries.

## Assumptions

- A1:
  - `compose.yaml` should allow partial mainnet address-policy configuration by leaving unused xpub env keys empty.
- A2:
  - `compose.test.yaml` should override `compose.yaml`, not replace it.
- A3:
  - Local/test runs may still rely on `deployments/compose/compose.test.env` to satisfy required env declared in the production-like base file.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - Number of Compose files needed for normal production deployment.
- Target:
  - One file: `compose.yaml`.
- Metric:
  - Production-like completeness.
- Target:
  - `compose.yaml` includes the mainnet poller and webhook dispatcher.
- Metric:
  - Test override completeness.
- Target:
  - `compose.yaml + compose.test.yaml` includes testnet4 polling, Swagger, and fake webhook receiver without extra overlay files.
