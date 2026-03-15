---
doc: 03_tasks
spec_date: 2026-03-15
slug: xpub-account-inference
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on: []
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
  - This is a localized derivation-path correctness fix with no schema change, new integration, or
    API contract change.
- Upstream dependencies (`depends_on`):
  - None.
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - The change stays inside existing Bitcoin outbound derivation and allocation plumbing.
  - What would trigger switching to Full mode:
    - Any database schema change, external wallet integration change, or multi-chain design
      expansion.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Not skipped.

## Milestones

- M1:
  - Make the Bitcoin deriver compute absolute derivation paths that honor xpub account metadata.
- M2:
  - Thread the computed absolute path through the existing allocation flow without changing public
    API contracts.
- M3:
  - Add regression tests for account-level and branch-level xpub path handling.

## Tasks (ordered)

1. T-001 - Compute absolute derivation paths from xpub metadata
   - Scope:
     - Update the Bitcoin outbound deriver to emit an absolute derivation path using configured
       prefixes plus xpub depth and child-index metadata.
   - Output:
     - Account-level xpubs no longer force account `0'` in persisted paths.
   - Linked requirements: FR-001, FR-002, NFR-002, NFR-003
   - Validation:
     - [ ] How to verify (manual steps or command): run Bitcoin outbound deriver tests.
     - [ ] Expected result: account-level and branch-level path tests pass with correct absolute
           derivation paths.
     - [ ] Logs/metrics to check (if applicable): none
2. T-002 - Preserve allocation-flow compatibility
   - Scope:
     - Thread the deriver-provided derivation path through allocation use cases and chain-level
       deriver plumbing without changing public API contracts.
   - Output:
     - Allocations store the corrected derivation path while request/response DTOs stay unchanged.
   - Linked requirements: FR-001, FR-003, NFR-001, NFR-003
   - Validation:
     - [ ] How to verify (manual steps or command): run use-case and blockchain deriver tests plus
           `go list ./...`.
     - [ ] Expected result: allocation tests stay green and compile-time plumbing is intact.
     - [ ] Logs/metrics to check (if applicable): none
3. T-003 - Add regression coverage and full verification
   - Scope:
     - Add or update tests covering non-zero account-level xpubs, branch-level xpubs, and full repo
       verification.
   - Output:
     - Automated regression coverage for the fixed derivation-path behavior.
   - Linked requirements: FR-002, FR-003, NFR-001
   - Validation:
     - [ ] How to verify (manual steps or command): run targeted `go test` commands and full
           `go test ./...`.
     - [ ] Expected result: new regression cases pass and the repository remains green.
     - [ ] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001, T-002
- FR-002 -> T-001, T-003
- FR-003 -> T-002, T-003
- NFR-001 -> T-002, T-003
- NFR-002 -> T-001
- NFR-003 -> T-001, T-002

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - Update deriver output first, then wire allocation flow, then refresh regression tests.
- Rollback steps:
  - Revert the derivation-path calculation to the previous prefix-plus-relative-path behavior if
    the new xpub metadata handling causes regressions.

## Validation evidence

- `SPEC_DIR="specs/2026-03-15-xpub-account-inference" bash scripts/spec-lint.sh`
- `gofmt -w internal/application/ports/outbound/chain_address_deriver.go internal/application/usecases/generate_address_use_case.go internal/application/usecases/allocate_payment_address_use_case.go internal/application/usecases/generate_address_use_case_test.go internal/application/usecases/allocate_payment_address_use_case_test.go internal/adapters/outbound/blockchain/multi_chain_address_deriver.go internal/adapters/outbound/blockchain/multi_chain_address_deriver_test.go internal/adapters/outbound/bitcoin/chain_address_deriver.go internal/adapters/outbound/bitcoin/chain_address_deriver_test.go internal/adapters/outbound/bitcoin/hd_xpub_address_deriver.go internal/adapters/outbound/bitcoin/hd_xpub_address_deriver_test.go`
- `go test ./internal/adapters/outbound/bitcoin`
- `go test ./internal/adapters/outbound/blockchain`
- `go test ./internal/application/usecases`
- `go list ./...`
- `go test ./...`
- `bash scripts/precommit-run.sh`
