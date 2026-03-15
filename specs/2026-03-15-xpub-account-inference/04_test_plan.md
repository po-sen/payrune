---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - Bitcoin outbound derivation-path calculation for account-level and branch-level xpubs.
  - Allocation-flow persistence of the deriver-provided path.
- Not covered:
  - New wallet identifier models or database schema changes.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, NFR-002, NFR-003
  - Steps:
    - Run the Bitcoin HD xpub deriver tests with a non-zero account-level xpub fixture.
  - Expected:
    - The absolute derivation path reflects the xpub child account index rather than a hard-coded
      `0'` account segment.
- TC-002:
  - Linked requirements: FR-002, NFR-003
  - Steps:
    - Run the Bitcoin HD xpub deriver tests with a branch-level xpub fixture.
  - Expected:
    - The absolute derivation path preserves the configured account prefix and uses the branch
      child index from the xpub plus the requested address index.
- TC-003:
  - Linked requirements: FR-001, FR-003, NFR-001
  - Steps:
    - Run allocation use-case tests that persist a derivation path after address derivation.
  - Expected:
    - The allocation flow stores the deriver output without regressing existing behavior.

### Integration

- TC-101:
  - Linked requirements: FR-003, NFR-001
  - Steps:
    - Run:
      - `go test ./internal/adapters/outbound/bitcoin ./internal/adapters/outbound/blockchain ./internal/application/usecases`
      - `go list ./...`
      - `go test ./...`
  - Expected:
    - The outbound derivation plumbing compiles cleanly and the repository remains green.

## Edge cases and failure modes

- Case:
  - Account-level xpub has a non-zero hardened child index.
- Expected behavior:
  - The emitted absolute derivation path contains the hardened account index derived from the xpub.
- Case:
  - Branch-level xpub is provided instead of an account-level xpub.
- Expected behavior:
  - The emitted absolute derivation path preserves the configured account prefix and appends the
    branch child index and requested address index.

## NFR verification

- Reliability:
  - Targeted and full Go test commands pass.
- Security:
  - The implementation relies on xpub metadata only and does not require private-key material.
- Maintainability:
  - Account inference is covered in the Bitcoin outbound deriver tests rather than duplicated across
    use cases.
