---
doc: 00_problem
spec_date: 2026-03-04
slug: compose-bitcoin-test-env
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-03-deploy-service-compose-dockerfile
  - 2026-03-04-bitcoin-address-vectors
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background: We already have compose overrides that accept bitcoin xpub values from environment variables.
- Users or stakeholders: Developers who want a ready-to-run local environment for manual API testing.
- Why now: We need a deterministic local startup profile that uses the same xpub fixtures as unit tests.

## Constraints (optional)

- Technical constraints:
  - Keep override as a separate compose file under `deployments/compose/`.
  - Reuse existing compose stacking with `COMPOSE_OVERRIDE`.
- Timeline/cost constraints:
  - Implement as a small change with minimal Makefile additions.
- Compliance/security constraints:
  - Include xpub values only (no private keys).

## Problem statement

- Current pain:
  - Local test environments require manual env setup for bitcoin xpubs.
- Evidence or examples:
  - Unit tests already include stable xpub fixtures, but compose startup does not provide a one-command profile with those values.

## Goals

- G1: Add one compose override (`compose.test-env.yaml`) with hardcoded bitcoin xpub fixtures aligned with unit test vectors.
- G2: Add one Makefile rule to start stack with this override.
- G3: Keep existing `make up` / `make down` behavior unchanged.

## Non-goals (out of scope)

- NG1: Add private key handling or secret management.
- NG2: Add new runtime feature flags unrelated to bitcoin xpub configuration.

## Assumptions

- A1: Hardcoded xpub fixtures are acceptable for local testing environments.
- A2: The fixture set should include both mainnet and testnet4 xpub variables currently supported by the app.

## Open questions

- Q1: Should we later split this test profile into separate mainnet/testnet4 preset files?
- Q2: Should we add a dedicated `down` helper for this profile or keep using existing `make down`?

## Success metrics

- Metric: Local startup ergonomics.
- Target: One Makefile command starts app with fixture-based bitcoin xpub env values.
- Metric: Fixture parity.
- Target: Override file xpub values match the current unit test fixture set.
