---
doc: 03_tasks
spec_date: 2026-03-09
slug: compose-mainnet-test-files
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-09-bitcoin-compose-defaults
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
  - This change restructures deployment files only and preserves runtime behavior.
- Upstream dependencies (`depends_on`):
  - `2026-03-09-bitcoin-compose-defaults`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - No schema, transport contract, or runtime orchestration changes are introduced.
  - What would trigger switching to Full mode:
    - Introducing a new deployment platform or runtime configuration subsystem.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Validation steps are listed under each task below.

## Milestones

- M1:
  - Define the new two-file Compose shape.
- M2:
  - Collapse the Compose files into a production-like base and a test override, then validate them.

## Tasks (ordered)

1. T-001 - Specify the two-file Compose deployment shape
   - Scope:
     - Document the shift from multiple overlays to a production-like base plus a test override.
   - Output:
     - `specs/2026-03-09-compose-mainnet-test-files/*.md`
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, NFR-002, NFR-003, NFR-004, NFR-008
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-09-compose-mainnet-test-files" bash scripts/spec-lint.sh`
     - [x] Expected result: spec lint passes and the target two-file deployment shape is explicit.
     - [x] Logs/metrics to check (if applicable): none
2. T-002 - Collapse Compose files into `compose.yaml` and `compose.test.yaml`
   - Scope:
     - Move mainnet settings and required core services into `compose.yaml`, make `compose.test.yaml` a local/test override, keep inherited service overrides minimal, remove obsolete overlay files, and simplify local Makefile wiring to one explicit override path.
   - Output:
     - `deployments/compose/compose.yaml`
     - `deployments/compose/compose.test.yaml`
     - `Makefile`
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, NFR-001, NFR-002, NFR-003, NFR-004, NFR-005, NFR-007, NFR-008
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-09-compose-mainnet-test-files" bash scripts/spec-lint.sh` and `docker compose --env-file deployments/compose/compose.test.env -f deployments/compose/compose.yaml config` and `docker compose --env-file deployments/compose/compose.test.env -f deployments/compose/compose.yaml -f deployments/compose/compose.test.yaml config`
     - [x] Expected result: production-like compose renders with provided env, the test override renders cleanly on top of the base file, and removed overlay files are no longer needed.
     - [x] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001, T-002
- FR-002 -> T-001, T-002
- FR-003 -> T-001, T-002
- FR-004 -> T-001, T-002
- FR-005 -> T-001, T-002
- NFR-001 -> T-002
- NFR-002 -> T-001, T-002
- NFR-003 -> T-001, T-002
- NFR-004 -> T-001, T-002
- NFR-005 -> T-002
- NFR-007 -> T-002
- NFR-008 -> T-001, T-002

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Restore the previous standalone-test shape if the base-plus-override model proves less clear in practice.
