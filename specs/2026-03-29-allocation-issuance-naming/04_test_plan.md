---
doc: 04_test_plan
spec_date: 2026-03-29
slug: allocation-issuance-naming
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-20-create2-eth-payment-receiving
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Allocation Issuance Naming - Test Plan

## Scope

- Covered:
  - Allocation migration and persistence SQL changes.
  - Core issuance model renames in domain/application code.
  - Bitcoin and Ethereum issued-address reference persistence behavior.
- Not covered:
  - End-to-end payment receipt polling across a live chain.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-002, FR-003
  - Steps:
    - Run Bitcoin deriver and allocation domain tests after the rename.
  - Expected:
    - Bitcoin outputs still persist absolute HD-path issuance references with
      `issuance_ref_kind=hd_path_absolute`.
- TC-002:
  - Linked requirements: FR-002, FR-003, NFR-003
  - Steps:
    - Run Ethereum deriver tests after the rename.
  - Expected:
    - Ethereum CREATE2 outputs persist salt-only issuance references with
      `issuance_ref_kind=create2_salt`.

### Integration

- TC-101:
  - Linked requirements: FR-001, FR-002, FR-003, NFR-002
  - Steps:
    - Run allocation store tests for both SQL adapters.
  - Expected:
    - Queries, inserts, updates, and issued-row reads use the renamed columns and still preserve
      allocation replay semantics.

### E2E (if applicable)

- Scenario 1:
  - Not applicable in this change.
- Scenario 2:
  - Not applicable in this change.

## Edge cases and failure modes

- Case:
  - Existing Ethereum issued rows still contain prefixed CREATE2 reference strings before migration.
- Expected behavior:
  - The migration rewrites them so only the salt payload remains in `issuance_ref` and sets
    `issuance_ref_kind=create2_salt`.

## NFR verification

- Performance:
  - No new performance target; focused tests must pass without adding extra full-table query paths.
- Reliability:
  - Ensure reservation, issued-allocation lookup, and idempotent replay tests still pass.
- Security:
  - Ensure no code path reintroduces policy-ID-prefixed CREATE2 issuance references after the
    refactor.
