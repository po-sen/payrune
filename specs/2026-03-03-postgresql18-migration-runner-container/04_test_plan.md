---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - Postgres 18 service definition and healthcheck behavior.
  - Migration command and baseline SQL migration execution.
  - Migration runner sequencing after DB readiness.
- Not covered:
  - Production-scale DB tuning and failover.

## Tests

### Unit

- TC-001: Migration command action handling

  - Linked requirements: FR-003, NFR-006
  - Steps:
    - Build/run migration command with supported action inputs in isolated tests where feasible.
  - Expected:
    - Command handles valid actions and error paths deterministically.

- TC-002: Migration SQL pair integrity
  - Linked requirements: FR-002
  - Steps:
    - Validate up/down files exist with matching version prefix.
  - Expected:
    - File set is complete for baseline migration.

### Integration

- TC-101: Compose config validation

  - Linked requirements: FR-001, FR-004, NFR-002
  - Steps:
    - Run `docker compose -f deployments/compose/compose.yaml config`.
  - Expected:
    - Services render correctly including `postgres` healthcheck and `migrate` depends_on.

- TC-102: End-to-end migration run
  - Linked requirements: FR-001, FR-002, FR-004, NFR-001, NFR-005
  - Steps:
    - Run `make up`.
    - Verify `postgres` healthy and `migrate` exited successfully.
    - Query DB for baseline table existence.
  - Expected:
    - Migration is applied and schema object is present.

### E2E (if applicable)

- Scenario 1:
  - Full local stack startup performs DB bootstrap without manual migration commands.

## Edge cases and failure modes

- Case: DB credentials mismatch.
- Expected behavior:

  - Migration runner exits non-zero with clear error log.

- Case: Migration already applied.
- Expected behavior:
  - Runner treats `no change` as successful completion.

## NFR verification

- Performance:
  - Ensure migration completion within 30 seconds post-DB readiness.
- Reliability:
  - Ensure postgres uses restart policy and healthcheck gating.
- Security:
  - Ensure credentials remain local-dev scoped and not embedded into application code.
