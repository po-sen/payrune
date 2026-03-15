---
doc: 01_requirements
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

# Requirements

## Out-of-scope behaviors

- OOS1:
  - Introducing new database columns or changing persistence keys.
- OOS2:
  - Inferring wallet ownership beyond the existing `address_policy_id` and xpub fingerprint model.

## Functional requirements

### FR-001 - Account-level xpub derivation paths reflect the real account index

- Description:
- When the configured xpub is account-level, the persisted derivation path must use the account
  child index encoded in the xpub rather than blindly preserving the configured account segment.
- Acceptance criteria:
- [ ] For an account-level xpub with a non-zero account child index, the deriver returns an
      absolute derivation path containing that account index.
- [ ] The allocation flow persists the deriver-provided absolute derivation path without
      rewriting it back to `.../0'`.
- Notes:
- This applies to Bitcoin account-level xpub inputs.

### FR-002 - Branch-level xpub derivation path behavior remains stable

- Description:
- When the configured xpub is already branch-level, the deriver must continue to emit the correct
  branch and index path under the configured account prefix.
- Acceptance criteria:
- [ ] Existing branch-level xpub address derivation tests continue to pass.
- [ ] The absolute derivation path uses the branch child index encoded in the xpub plus the
      requested address index.
- Notes:
- The configured prefix remains the source of the account segment for branch-level xpubs.

### FR-003 - Existing public behavior stays compatible

- Description:
- Address generation and allocation APIs must keep their existing external contracts while the
  internal derivation-path calculation changes.
- Acceptance criteria:
- [ ] No public DTO shape changes are required for the API layer.
- [ ] `go test ./...` passes after the derivation-path change.
- Notes:
- This is an internal correctness fix, not an API redesign.

## Non-functional requirements

- Reliability (NFR-001):
  - Targeted Bitcoin derivation tests and full `go test ./...` must pass after the change.
- Security/Privacy (NFR-002):
  - The implementation must inspect xpub metadata only and must not require private keys.
- Maintainability (NFR-003):
  - Account inference logic must live in the Bitcoin outbound derivation path rather than being
    duplicated across use cases.

## Dependencies and integrations

- External systems:
  - None.
- Internal services:
  - `internal/adapters/outbound/bitcoin`
  - `internal/adapters/outbound/blockchain`
  - `internal/application/usecases`
