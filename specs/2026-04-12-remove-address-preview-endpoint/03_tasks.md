---
doc: 03_tasks
spec_date: 2026-04-12
slug: remove-address-preview-endpoint
mode: Quick
status: DONE
owners:
  - codex
depends_on: []
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale:
  - This is a focused public contract removal with no new persistence, integration, or rollout design.
- Upstream dependencies (`depends_on`): []
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - The change is deletion-only and the affected code paths are small and explicit.
  - What would trigger switching to Full mode:
    - A replacement preview mechanism or any new external integration.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - In each task below.

## Milestones

- M1:
  - Remove the preview API surface.
- M2:
  - Remove preview-only implementation artifacts and validate the remaining API surface.

## Tasks (ordered)

1. T-001 - Remove preview route and API contract
   - Scope:
     - Remove router registration, bootstrap wiring, and Swagger documentation for `/v1/chains/{chain}/addresses`.
   - Output:
     - Public API no longer exposes the preview route.
   - Linked requirements: FR-001, FR-003, NFR-001
   - Validation:
     - [x] How to verify (manual steps or command): `rg -n "/v1/chains/\\{chain\\}/addresses|GenerateAddress" internal deployments/swagger README.md -S`
     - [x] Expected result:
       - No runtime route or Swagger path remains for the preview endpoint.
     - [x] Logs/metrics to check (if applicable):
       - None.
2. T-002 - Remove preview-only runtime code
   - Scope:
     - Delete preview-only controller, use case, DTO, error, and policy helpers, plus tests that exist only for preview.
   - Output:
     - No preview-only runtime abstractions remain.
   - Linked requirements: FR-002, NFR-001
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./internal/...`
     - [x] Expected result:
       - Build and tests pass after removing preview-specific code.
     - [x] Logs/metrics to check (if applicable):
       - None.
3. T-003 - Final validation and spec closeout
   - Scope:
     - Lint the spec, run repo validation, and update spec status to final state.
   - Output:
     - Implementation and spec are both complete.
   - Linked requirements: NFR-002, NFR-003
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-04-12-remove-address-preview-endpoint" bash scripts/spec-lint.sh`
     - [x] Expected result:
       - Spec lint passes.
     - [x] Logs/metrics to check (if applicable):
       - `bash scripts/precommit-run.sh` passes.

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-002
- FR-003 -> T-001
- NFR-001 -> T-001, T-002
- NFR-002 -> T-003
- NFR-003 -> T-003

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Restore the removed route and preview code from git if any client still depends on it.

## Validation evidence

- `go test ./internal/...`
- `SPEC_DIR="specs/2026-04-12-remove-address-preview-endpoint" bash scripts/spec-lint.sh`
- `bash scripts/precommit-run.sh`
