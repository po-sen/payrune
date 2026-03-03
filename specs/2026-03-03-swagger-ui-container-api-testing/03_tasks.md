---
doc: 03_tasks
spec_date: 2026-03-03
slug: swagger-ui-container-api-testing
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-03-deploy-service-compose-dockerfile
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
  - Change includes cross-origin security behavior (CORS) and integration behavior across browser, Swagger UI, and API service.
  - Design clarity is needed for flow, failure modes, and security constraints.
- Upstream dependencies (`depends_on`):
  - `2026-03-03-deploy-service-compose-dockerfile`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: not applicable
  - What would trigger switching to Full mode: not applicable
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): not applicable

## Milestones

- M1: Spec updated and linted for direct-8080 + CORS behavior.
- M2: CORS middleware and Swagger contract updated.
- M3: Local compose verification completed.

## Tasks (ordered)

1. T-001 - Update and lint Full-mode spec package

   - Scope:
     - Update existing swagger spec docs to direct 8080 call model and CORS requirements.
   - Output:
     - `specs/2026-03-03-swagger-ui-container-api-testing/*.md`
   - Linked requirements: FR-001, FR-002, FR-003, NFR-004
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-03-swagger-ui-container-api-testing" bash scripts/spec-lint.sh`
     - [x] Expected result: lint exits 0 with no header/link/traceability errors.
     - [x] Logs/metrics to check (if applicable): lint output contains no failures.

2. T-002 - Implement CORS middleware at inbound HTTP layer

   - Scope:
     - Add reusable HTTP middleware and wire it in bootstrap handler chain.
   - Output:
     - `internal/adapters/inbound/http/middleware/cors.go`
     - `internal/adapters/inbound/http/middleware/cors_test.go`
     - `internal/bootstrap/app.go`
   - Linked requirements: FR-002, NFR-003, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./...`
     - [x] Expected result: tests pass including CORS middleware tests.
     - [x] Logs/metrics to check (if applicable): middleware tests validate headers and preflight.

3. T-003 - Update Swagger OpenAPI target and compose wiring

   - Scope:
     - Point OpenAPI server to 8080 and keep compose swagger service minimal for direct mode.
   - Output:
     - `deployments/swagger/openapi.yaml`
     - `deployments/compose/compose.yaml`
   - Linked requirements: FR-001, FR-003, NFR-002
   - Validation:
     - [x] How to verify (manual steps or command): `docker compose -f deployments/compose/compose.yaml config`
     - [x] Expected result: compose renders valid swagger and payrune services.
     - [x] Logs/metrics to check (if applicable): config output contains swagger service and openapi mount.

4. T-004 - Verify direct call and CORS headers end-to-end
   - Scope:
     - Run stack and confirm direct request to 8080 includes CORS headers for swagger origin.
   - Output:
     - Manual verification evidence from curl and service status.
   - Linked requirements: FR-002, FR-003, NFR-001, NFR-005
   - Validation:
     - [x] How to verify (manual steps or command): `make up && curl -i -H "Origin: http://localhost:8081" http://localhost:8080/health && make down`
     - [x] Expected result: 200 response includes allow-origin and health JSON.
     - [x] Logs/metrics to check (if applicable): `docker compose -f deployments/compose/compose.yaml ps` and logs.

## Traceability (optional)

- FR-001 -> T-001, T-003
- FR-002 -> T-001, T-002, T-004
- FR-003 -> T-001, T-003, T-004
- NFR-001 -> T-004
- NFR-002 -> T-003
- NFR-003 -> T-002
- NFR-004 -> T-001
- NFR-005 -> T-004
- NFR-006 -> T-002

## Rollout and rollback

- Feature flag:
  - Not required.
- Migration sequencing:
  - Update spec, then middleware/contract/compose, then verify locally.
- Rollback steps:
  - Revert CORS middleware wiring and OpenAPI server URL if regression is found.

## Ready-to-code checklist

- [x] Full-mode docs are present (`00` through `04`).
- [x] Frontmatter values are consistent across docs.
- [x] `depends_on` references existing DONE dependencies.
- [x] Mode decision and rationale are documented.
- [x] Requirement/task/test traceability is defined.
