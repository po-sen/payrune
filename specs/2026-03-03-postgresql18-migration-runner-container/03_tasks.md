---
doc: 03_tasks
spec_date: 2026-03-03
slug: postgresql18-migration-runner-container
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-03-deploy-service-compose-dockerfile
  - 2026-03-03-swagger-ui-container-api-testing
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Task Plan

## Mode decision

- Selected mode: Full
- Rationale:
  - Change introduces new external integration (PostgreSQL) and migration orchestration flow.
  - Requires explicit design for sequencing, failure handling, and migration contracts.
- Upstream dependencies (`depends_on`):
  - `2026-03-03-deploy-service-compose-dockerfile`
  - `2026-03-03-swagger-ui-container-api-testing`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: not applicable
  - What would trigger switching to Full mode: not applicable
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): not applicable

## Milestones

- M1: Full-mode spec prepared and linted.
- M2: Migration command and SQL assets implemented.
- M3: Postgres + migrate containers verified in compose.

## Tasks (ordered)

1. T-001 - Finalize Full-mode PostgreSQL migration spec

   - Scope:
     - Complete and lint all five spec documents for this feature.
   - Output:
     - `specs/2026-03-03-postgresql18-migration-runner-container/*.md`
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, NFR-004
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-03-postgresql18-migration-runner-container" bash scripts/spec-lint.sh`
     - [x] Expected result: lint exits with code 0.
     - [x] Logs/metrics to check (if applicable): no lint failures.

2. T-002 - Implement migration command and SQL assets

   - Scope:
     - Add `cmd/migrate` executable and baseline up/down SQL files.
   - Output:
     - `cmd/migrate/main.go`
     - migration SQL files
     - `go.mod`/`go.sum` dependency updates
   - Linked requirements: FR-002, FR-003, NFR-005, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./...`
     - [x] Expected result: build/test succeeds including migration command package.
     - [x] Logs/metrics to check (if applicable): migration command logs show no-change/apply semantics.

3. T-003 - Add migration runner image/container and postgres service

   - Scope:
     - Add postgres service, migration runner service, and migration build image in compose/deployment files.
   - Output:
     - `deployments/compose/compose.yaml`
     - `build/migrate/Dockerfile`
   - Linked requirements: FR-001, FR-004, NFR-001, NFR-002, NFR-003
   - Validation:
     - [x] How to verify (manual steps or command): `docker compose -f deployments/compose/compose.yaml config`
     - [x] Expected result: config renders valid services/healthcheck/depends_on.
     - [x] Logs/metrics to check (if applicable): config output includes `postgres` and `migrate` services.

4. T-004 - Verify end-to-end startup and schema bootstrap
   - Scope:
     - Run stack and confirm postgres health, migration completion, and table creation.
   - Output:
     - command outputs as evidence.
   - Linked requirements: FR-001, FR-002, FR-004, NFR-001, NFR-005
   - Validation:
     - [x] How to verify (manual steps or command): `make up`, inspect compose status/logs, and query DB table existence.
     - [x] Expected result: postgres is healthy, migrate exits 0, baseline table exists.
     - [x] Logs/metrics to check (if applicable): `docker compose ... logs migrate` shows successful run.

## Traceability (optional)

- FR-001 -> T-001, T-003, T-004
- FR-002 -> T-001, T-002, T-004
- FR-003 -> T-001, T-002
- FR-004 -> T-001, T-003, T-004
- NFR-001 -> T-003, T-004
- NFR-002 -> T-003
- NFR-003 -> T-003
- NFR-004 -> T-001
- NFR-005 -> T-002, T-004
- NFR-006 -> T-002

## Rollout and rollback

- Feature flag:
  - Not required.
- Migration sequencing:
  - Start postgres, run migrate job, then continue normal service startup.
- Rollback steps:
  - Run migration down (if required), remove postgres/migrate services from compose, revert related files.

## Ready-to-code checklist

- [x] Full-mode docs are present (`00` through `04`).
- [x] Frontmatter values are consistent across docs.
- [x] `depends_on` references existing DONE specs.
- [x] Mode decision and rationale are documented.
- [x] Requirement/task/test traceability IDs are present.
