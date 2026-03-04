---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Test env override:
  - A compose override file that injects deterministic xpub values for local testing.

## Out-of-scope behaviors

- OOS1: Managing secrets or encrypted config stores.
- OOS2: Changing API behavior for derivation logic.

## Functional requirements

### FR-001 - Compose override with hardcoded unit-test xpub fixtures

- Description:
  - Repository MUST provide a dedicated compose override file that sets bitcoin xpub env values to the same fixtures used by current unit tests.
- Acceptance criteria:
  - [x] New file exists under `deployments/compose/`.
  - [x] File sets all supported app bitcoin xpub env vars:
    - `BITCOIN_MAINNET_LEGACY_XPUB`
    - `BITCOIN_MAINNET_SEGWIT_XPUB`
    - `BITCOIN_MAINNET_NATIVE_SEGWIT_XPUB`
    - `BITCOIN_MAINNET_TAPROOT_XPUB`
    - `BITCOIN_TESTNET4_LEGACY_XPUB`
    - `BITCOIN_TESTNET4_SEGWIT_XPUB`
    - `BITCOIN_TESTNET4_NATIVE_SEGWIT_XPUB`
    - `BITCOIN_TESTNET4_TAPROOT_XPUB`
  - [x] Values match the fixture xpubs currently hardcoded in bitcoin encoder vector tests.
- Notes:
  - Override targets local developer usage.
  - Override filename should be `compose.test-env.yaml` since there is only one test profile.

### FR-002 - Make rule for override startup

- Description:
  - `Makefile` MUST provide a simple rule to start docker compose with the test env override file.
- Acceptance criteria:
  - [x] A new make target exists for startup with the override file.
  - [x] The target builds and starts services via existing compose flow.
- Notes:
  - Target may hardcode the override path directly for simplicity.

### FR-003 - Preserve existing startup/down behavior

- Description:
  - Existing `make up` and `make down` behavior MUST remain unchanged.
- Acceptance criteria:
  - [x] `make up` still starts with base compose unless `COMPOSE_OVERRIDE` is provided.
  - [x] `make down` still stops current compose stack normally.
- Notes:
  - New rule is additive only.

## Non-functional requirements

- Performance (NFR-001): Added make rule should execute with same order of magnitude startup time as `make up`.
- Availability/Reliability (NFR-002): Override values are deterministic and do not depend on shell env.
- Security/Privacy (NFR-003): Override file contains only xpub values, no private keys.
- Compliance (NFR-004): New spec and code pass repo lint/test hooks.
- Observability (NFR-005): Startup command for test profile is explicit and discoverable in Makefile.
- Maintainability (NFR-006): Override file remains scoped to bitcoin env injection and avoids unrelated settings.

## Dependencies and integrations

- External systems:
  - Docker Compose.
- Internal services:
  - `app` service environment loading in compose stack.
