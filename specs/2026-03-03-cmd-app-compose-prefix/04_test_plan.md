---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - Command directory rename and build-path update.
  - Compose project prefix configuration.
  - Local stack startup/shutdown continuity.
- Not covered:
  - Functional API behavior changes (none expected).

## Tests

### Unit

- TC-001: Command path integrity

  - Linked requirements: FR-001, FR-002
  - Steps:
    - Verify `cmd/app/main.go` exists and old path is absent.
  - Expected:
    - Build path references are consistent.

- TC-002: Go compile/test smoke
  - Linked requirements: FR-001, FR-002, NFR-006
  - Steps:
    - Run `go test ./...`.
  - Expected:
    - All packages compile and tests pass.

### Integration

- TC-101: Compose config prefix validation

  - Linked requirements: FR-003, NFR-005
  - Steps:
    - Run `docker compose -f deployments/compose/compose.yaml config`.
  - Expected:
    - Config output includes `name: payrune`.

- TC-102: Stack lifecycle smoke
  - Linked requirements: FR-003, NFR-001, NFR-002
  - Steps:
    - Run `make up` then `make down`.
  - Expected:
    - Services start/stop successfully with prefixed compose resources.

### E2E (if applicable)

- Scenario 1:
  - Developer can keep existing workflow (`make up`/`make down`) without additional manual fixes.

## Edge cases and failure modes

- Case: Stale old compose project resources exist.
- Expected behavior:
  - New stack still starts under `payrune` project name; old resources can be cleaned separately.

## NFR verification

- Performance:
  - Startup/shutdown time remains close to current local baseline.
- Reliability:
  - Service dependency order remains intact.
- Security:
  - No new sensitive config introduced.
