---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - Hardcoded fixture vectors in scheme-specific tests + `Makefile` target.
  - Exact-match address verification for provided index-0 vectors.
- Not covered:
  - Runtime API behavior changes.
  - Non-zero derivation index vector matrix.

## Tests

### Unit

- TC-001: Zero-setup make target execution

  - Linked requirements: FR-001, FR-002, NFR-002
  - Steps:
    - Run `make test-address-vectors` without setting any env.
  - Expected:
    - Test command executes and vectors are loaded from hardcoded test fixtures.

- TC-002: Mainnet vector exact-match

  - Linked requirements: FR-003, NFR-005, NFR-006
  - Steps:
    - Derive `index=0` for mainnet `legacy`, `segwit`, `nativeSegwit`, `taproot` using fixture xpub.
    - Compare against fixture expected addresses.
  - Expected:
    - All mainnet vectors match exactly.

- TC-003: Testnet4 vector exact-match

  - Linked requirements: FR-003, NFR-005, NFR-006
  - Steps:
    - Derive `index=0` for testnet4 `legacy`, `segwit`, `nativeSegwit`, `taproot` using fixture xpub.
    - Compare against fixture expected addresses.
  - Expected:
    - All testnet4 vectors match exactly.

- TC-004: Scheme-specific test-file ownership
  - Linked requirements: FR-004, NFR-006
  - Steps:
    - Confirm vector assertions are split into:
      - `address_encoder_legacy_test.go`
      - `address_encoder_segwit_test.go`
      - `address_encoder_native_segwit_test.go`
      - `address_encoder_taproot_test.go`
  - Expected:
    - Each address scheme has dedicated vector assertions in its own test file.

### Integration

- TC-101: Targeted package run
  - Linked requirements: FR-002, NFR-001
  - Steps:
    - Run `go test ./internal/adapters/outbound/bitcoin -run AddressEncoderProvidedVectors -count=1`.
  - Expected:
    - Execution completes quickly and runs all four scheme-specific vector tests.

### E2E (if applicable)

- Scenario 1:
  - Not applicable.
- Scenario 2:
  - Not applicable.

## Edge cases and failure modes

- Case:
  - Accidental change of hardcoded vector value.
- Expected behavior:
  - Exact-match assertion fails with clear expected/actual output.

## NFR verification

- Performance:
  - Verify vector test completes under 3 seconds.
- Reliability:
  - Run same command multiple times; results remain stable.
- Security:
  - Confirm hardcoded fixtures contain xpub only and no private key material.
