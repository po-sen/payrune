---
doc: 01_requirements
spec_date: 2026-03-04
slug: bitcoin-address-vectors
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-03-btc-xpub-address-api
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Requirements

## Glossary (optional)

- Vector fixture:
- A pair of `(xpub, expected address)` for a fixed derivation input.

## Out-of-scope behaviors

- OOS1: Runtime API changes for vector input.
- OOS2: Secret/private key handling.

## Functional requirements

### FR-001 - Hardcoded vector fixtures in test code

- Description:
  - Repository MUST keep provided xpub and expected-address vectors directly in unit test code.
- Acceptance criteria:
  - [x] Mainnet/testnet4 vectors for `legacy`, `segwit`, `nativeSegwit`, and `taproot` are embedded in scheme-specific test files.
  - [x] Vector assertions do not use env lookup.
- Notes:
  - This avoids per-environment setup when running tests.

### FR-002 - Make target for vector test run

- Description:
  - `Makefile` MUST expose a target to run vector verification tests without env preloading.
- Acceptance criteria:
  - [x] `Makefile` does not require `env.mk` include/export for vector tests.
  - [x] `make test-address-vectors` runs the vector-specific unit test.
- Notes:
  - Existing `make up/down` behavior remains unchanged.

### FR-003 - Exact-match vector unit tests

- Description:
  - Unit tests MUST derive `index=0` addresses from fixture xpub values and assert exact equality with expected addresses.
- Acceptance criteria:
  - [x] Tests cover mainnet `legacy`, `segwit`, `nativeSegwit`, `taproot`.
  - [x] Tests cover testnet4 `legacy`, `segwit`, `nativeSegwit`, `taproot`.
  - [x] Failure messages include vector key and expected/actual values.
- Notes:
  - Tests are regression guards for derivation-path and encoder behavior.

### FR-004 - Scheme-specific test file split

- Description:
  - Vector assertions MUST be split by encoder scheme into dedicated test files.
- Acceptance criteria:
  - [x] `legacy` vector assertions are in `internal/adapters/outbound/bitcoin/address_encoder_legacy_test.go`.
  - [x] `segwit` vector assertions are in `internal/adapters/outbound/bitcoin/address_encoder_segwit_test.go`.
  - [x] `nativeSegwit` vector assertions are in `internal/adapters/outbound/bitcoin/address_encoder_native_segwit_test.go`.
  - [x] `taproot` vector assertions are in `internal/adapters/outbound/bitcoin/address_encoder_taproot_test.go`.
- Notes:
  - Shared helper utilities may stay in `address_encoder_helpers_test.go`.

## Non-functional requirements

- Performance (NFR-001): Vector test suite completes in under 3 seconds on local machine.
- Availability/Reliability (NFR-002): Vector tests are deterministic and do not depend on network calls.
- Security/Privacy (NFR-003): Fixtures use xpub only; no private keys in repository.
- Compliance (NFR-004): New spec passes `scripts/spec-lint.sh`.
- Observability (NFR-005): Test output clearly identifies failing vector labels.
- Maintainability (NFR-006): New tests stay inside outbound adapter test package and avoid cross-layer coupling.

## Dependencies and integrations

- External systems:
  - None.
- Internal services:
  - Existing bitcoin outbound deriver and encoder implementations.
