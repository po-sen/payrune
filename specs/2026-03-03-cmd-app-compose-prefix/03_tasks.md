---
doc: 03_tasks
spec_date: 2026-03-03
slug: cmd-app-compose-prefix
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-03-postgresql18-migration-runner-container
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
  - Change is limited to command path rename and compose naming configuration.
  - No new persistent model, external integration, or complex failure topology added.
- Upstream dependencies (`depends_on`):
  - `2026-03-03-postgresql18-migration-runner-container`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: implementation is small and structurally straightforward.
  - What would trigger switching to Full mode: broader runtime architecture or rollout risk changes.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): not applicable; `04_test_plan.md` is produced.

## Milestones

- M1: Quick spec ready and linted.
- M2: Command rename + build reference updates completed.
- M3: Compose prefix validated with local config and startup checks.

## Tasks (ordered)

1. T-001 - Finalize and lint spec package

   - Scope:
     - Write and validate quick-mode docs for this rename/prefix change.
   - Output:
     - `specs/2026-03-03-cmd-app-compose-prefix/*.md`
   - Linked requirements: FR-001, FR-002, FR-003, NFR-004
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-03-cmd-app-compose-prefix" bash scripts/spec-lint.sh`
     - [x] Expected result: lint exits with code 0.
     - [x] Logs/metrics to check (if applicable): no lint failures.

2. T-002 - Rename command directory and update build path

   - Scope:
     - Move `cmd/payrune` to `cmd/app` and update Docker build path references.
   - Output:
     - `cmd/app/main.go`
     - `build/app/Dockerfile`
   - Linked requirements: FR-001, FR-002, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./...`
     - [x] Expected result: all packages compile and tests pass.
     - [x] Logs/metrics to check (if applicable): `cmd/app` appears in package listing/build paths.

3. T-003 - Add compose project prefix `payrune`
   - Scope:
     - Add explicit compose name and validate compose stack behavior.
   - Output:
     - `deployments/compose/compose.yaml`
   - Linked requirements: FR-003, NFR-001, NFR-002, NFR-005
   - Validation:
     - [x] How to verify (manual steps or command): `docker compose -f deployments/compose/compose.yaml config` and `make up && make down`
     - [x] Expected result: compose config is valid and stack lifecycle works.
     - [x] Logs/metrics to check (if applicable): rendered config indicates project name `payrune`.

## Traceability (optional)

- FR-001 -> T-001, T-002
- FR-002 -> T-001, T-002
- FR-003 -> T-001, T-003
- NFR-001 -> T-003
- NFR-002 -> T-003
- NFR-004 -> T-001
- NFR-005 -> T-003
- NFR-006 -> T-002

## Rollout and rollback

- Feature flag:
  - Not required.
- Migration sequencing:
  - Apply path rename and compose naming updates together, then run startup verification.
- Rollback steps:
  - Revert command path and compose name changes if build/startup regressions appear.
