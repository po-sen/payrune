---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - Docker image build and runtime behavior for `payrune`.
  - Compose lifecycle (`up`/`down`) and service port mapping.
  - Makefile wrappers for compose lifecycle.
- Not covered:
  - Remote image publishing.
  - Production deployment environments.

## Tests

### Unit

- TC-001: Dockerfile static contract

  - Linked requirements: FR-001, NFR-003
  - Steps:
    - Verify `build/app/Dockerfile` includes build stage and runtime stage.
    - Verify runtime user is non-root and entrypoint/cmd runs service binary.
  - Expected:
    - Dockerfile structure matches security/runtime requirements.

- TC-002: Makefile static contract
  - Linked requirements: FR-003, NFR-006
  - Steps:
    - Verify `Makefile` defines `up` and `down` targets and shared compose variables.
  - Expected:
    - Minimal targets are present with no unrelated complexity.

### Integration

- TC-101: Compose build and boot

  - Linked requirements: FR-001, FR-002, NFR-002, NFR-005
  - Steps:
    - Run `docker compose -f deployments/compose/compose.yaml up -d --build`.
    - Run `docker compose -f deployments/compose/compose.yaml ps`.
  - Expected:
    - Image builds successfully and service status is `Up`.

- TC-102: Health endpoint check via make wrappers
  - Linked requirements: FR-003, NFR-001
  - Steps:
    - Run `make up`.
    - Run `curl -sf http://localhost:8080/health`.
    - Run `make down`.
  - Expected:
    - Health endpoint returns JSON and make lifecycle commands succeed.

### E2E (if applicable)

- Scenario 1:
  - Developer starts local container stack with `make up` and confirms endpoint accessibility.
- Scenario 2:
  - Developer stops stack cleanly with `make down` and confirms no running compose services.

## Edge cases and failure modes

- Case: Port 8080 already in use.
- Expected behavior: `make up` fails with bind error and no partial successful startup is reported.

- Case: Docker daemon unavailable.
- Expected behavior: compose command fails fast with actionable error output.

## NFR verification

- Performance:
  - Verify `/health` is reachable within 15 seconds after `make up`.
- Reliability:
  - Verify compose service includes `restart: unless-stopped`.
- Security:
  - Verify runtime image user is non-root.
