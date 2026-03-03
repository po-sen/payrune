---
doc: 03_tasks
spec_date: 2026-03-03
slug: deploy-service-compose-dockerfile
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
  - Scope is limited to deployment/config files and simple developer commands.
  - No new persistent data model, external runtime integration, or complex failure workflow is introduced.
- Upstream dependencies (`depends_on`): []
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: architecture and domain behavior are unchanged; only packaging/deployment wiring is added.
  - What would trigger switching to Full mode: adding environment matrix, external infra dependencies, or rollout/risk controls.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): not applicable; `04_test_plan.md` is produced.

## Milestones

- M1: Quick-mode spec package prepared and linted.
- M2: Dockerfile, compose manifest, and Makefile targets implemented and verified.

## Tasks (ordered)

1. T-001 - Finalize spec package for deployment scaffold

   - Scope:
     - Fill problem, requirements, tasks, and test plan documents with concrete details for Docker/Compose/Make changes.
   - Output:
     - `specs/2026-03-03-deploy-service-compose-dockerfile/00_problem.md`
     - `specs/2026-03-03-deploy-service-compose-dockerfile/01_requirements.md`
     - `specs/2026-03-03-deploy-service-compose-dockerfile/03_tasks.md`
     - `specs/2026-03-03-deploy-service-compose-dockerfile/04_test_plan.md`
   - Linked requirements: FR-001, FR-002, FR-003, NFR-004
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-03-deploy-service-compose-dockerfile" bash scripts/spec-lint.sh`
     - [x] Expected result: spec lint passes with no missing keys/placeholders.
     - [x] Logs/metrics to check (if applicable): lint output exits with code 0.

2. T-002 - Add Dockerfile and compose deployment manifest

   - Scope:
     - Add runtime packaging and compose service definition for local deployment.
   - Output:
     - `build/app/Dockerfile`
     - `deployments/compose/compose.yaml`
   - Linked requirements: FR-001, FR-002, NFR-002, NFR-003, NFR-005
   - Validation:
     - [x] How to verify (manual steps or command): `docker compose -f deployments/compose/compose.yaml up -d --build`
     - [x] Expected result: one running `payrune` container and build success.
     - [x] Logs/metrics to check (if applicable): `docker compose -f deployments/compose/compose.yaml ps` shows service `Up`.

3. T-003 - Add minimal Makefile up/down commands
   - Scope:
     - Add concise lifecycle wrappers for compose up/down.
   - Output:
     - `Makefile`
   - Linked requirements: FR-003, NFR-001, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `make up && curl -sf http://localhost:8080/health && make down`
     - [x] Expected result: health endpoint returns HTTP 200 JSON while service is up, then containers are removed.
     - [x] Logs/metrics to check (if applicable): `docker compose -f deployments/compose/compose.yaml logs payrune` shows service logs.

## Traceability (optional)

- FR-001 -> T-001, T-002
- FR-002 -> T-001, T-002
- FR-003 -> T-001, T-003
- NFR-001 -> T-003
- NFR-002 -> T-002
- NFR-003 -> T-002
- NFR-004 -> T-001
- NFR-005 -> T-002, T-003
- NFR-006 -> T-003

## Rollout and rollback

- Feature flag:
  - Not required.
- Migration sequencing:
  - Apply spec, then add Dockerfile/compose, then add Makefile wrappers.
- Rollback steps:
  - Revert `build/app/Dockerfile`, `deployments/compose/compose.yaml`, and `Makefile` to remove deployment scaffold.

## Ready-to-code checklist

- [x] Quick-mode docs exist (`00_problem.md`, `01_requirements.md`, `03_tasks.md`).
- [x] Optional `04_test_plan.md` is produced and linked.
- [x] Frontmatter (`spec_date`, `slug`, `mode`, `status`, `owners`, `depends_on`) is consistent across docs.
- [x] Links key set is complete and valid.
- [x] Mode decision and rationale are documented.
- [x] All requirements and tasks have traceable IDs.
