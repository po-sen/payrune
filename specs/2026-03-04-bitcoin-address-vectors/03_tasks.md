---
doc: 03_tasks
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

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale:
  - Change is focused on test fixtures + unit-test coverage without new API/DB/integration design changes.
- Upstream dependencies (`depends_on`):
  - `2026-03-03-btc-xpub-address-api`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - No new architecture component or external integration is introduced.
  - What would trigger switching to Full mode:
    - Adding runtime derivation-policy metadata or persistence for vectors.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Not skipped.

## Milestones

- M1: Spec package for vector-based test enhancement is complete.
- M2: Hardcoded vectors + make wiring implemented.
- M3: Vector tests implemented and validated.
- M4: Vector assertions are split by scheme-specific encoder test files.

## Tasks (ordered)

1. T-001 - Update and lint Quick-mode spec

   - Scope:
     - Define fixture format, make target behavior, and vector test acceptance.
   - Output:
     - `specs/2026-03-04-bitcoin-address-vectors/*.md`
   - Linked requirements: FR-001, FR-002, FR-003, NFR-004
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-04-bitcoin-address-vectors" bash scripts/spec-lint.sh`
     - [x] Expected result: lint exits with code 0.
     - [x] Logs/metrics to check (if applicable): no lint errors.

2. T-002 - Remove env injection and keep simple make target

   - Scope:
     - Remove env-based fixture injection and keep `Makefile` target for vector tests.
   - Output:
     - `Makefile`
     - `internal/adapters/outbound/bitcoin/address_encoder_helpers_test.go`
   - Linked requirements: FR-001, FR-002, NFR-002
   - Validation:
     - [x] How to verify (manual steps or command): `make test-address-vectors`
     - [x] Expected result: command runs vector tests with no env setup.
     - [x] Logs/metrics to check (if applicable): test command uses expected package/run filter.

3. T-003 - Implement vector exact-match unit tests

   - Scope:
     - Add unit tests for 8 hardcoded vectors (mainnet/testnet4 x 4 schemes) at `index=0`.
   - Output:
     - `internal/adapters/outbound/bitcoin/address_encoder_legacy_test.go`
     - `internal/adapters/outbound/bitcoin/address_encoder_segwit_test.go`
     - `internal/adapters/outbound/bitcoin/address_encoder_native_segwit_test.go`
     - `internal/adapters/outbound/bitcoin/address_encoder_taproot_test.go`
     - `internal/adapters/outbound/bitcoin/address_encoder_helpers_test.go`
   - Linked requirements: FR-003, NFR-001, NFR-005, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./internal/adapters/outbound/bitcoin -run AddressEncoderProvidedVectors -count=1`
     - [x] Expected result: all vectors pass exact-match assertion.
     - [x] Logs/metrics to check (if applicable): mismatch output includes vector label, expected address, actual address.

4. T-004 - Keep vector tests split by scheme file
   - Scope:
     - Remove single aggregate vector test file and ensure each scheme file owns its own vectors.
   - Output:
     - `internal/adapters/outbound/bitcoin/address_encoder_legacy_test.go`
     - `internal/adapters/outbound/bitcoin/address_encoder_segwit_test.go`
     - `internal/adapters/outbound/bitcoin/address_encoder_native_segwit_test.go`
     - `internal/adapters/outbound/bitcoin/address_encoder_taproot_test.go`
   - Linked requirements: FR-004, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./internal/adapters/outbound/bitcoin -run AddressEncoderProvidedVectors -count=1`
     - [x] Expected result: each scheme test file contains and executes its own vector assertions.
     - [x] Logs/metrics to check (if applicable): test names map directly to scheme-specific files.

## Traceability (optional)

- FR-001 -> T-001, T-002
- FR-002 -> T-001, T-002
- FR-003 -> T-001, T-003
- FR-004 -> T-004
- NFR-001 -> T-003
- NFR-002 -> T-002
- NFR-004 -> T-001
- NFR-005 -> T-003
- NFR-006 -> T-003, T-004

## Rollout and rollback

- Feature flag:
  - Not applicable.
- Migration sequencing:
  - Not applicable.
- Rollback steps:
  - Revert `Makefile` target and scheme-specific vector test changes.
