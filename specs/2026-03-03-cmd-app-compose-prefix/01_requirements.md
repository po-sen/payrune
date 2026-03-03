---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Compose prefix: project name used by Docker Compose for generated resource names.
- App command: entrypoint command source under `cmd/app`.

## Out-of-scope behaviors

- OOS1: Renaming Go module path `payrune`.
- OOS2: Refactoring runtime architecture beyond command path updates.

## Functional requirements

### FR-001 - App command directory rename

- Description:
  - Repository MUST move the main service command path from `cmd/payrune` to `cmd/app`.
- Acceptance criteria:
  - [ ] `cmd/app/main.go` exists.
  - [ ] `cmd/payrune/main.go` no longer exists.
  - [ ] `go test ./...` succeeds after rename.
- Notes:
  - Imports/package semantics should remain unchanged.

### FR-002 - Build reference updates for renamed command

- Description:
  - Build/deploy assets MUST reference the new command path.
- Acceptance criteria:
  - [ ] `build/app/Dockerfile` builds from `./cmd/app`.
  - [ ] Local compose app image still builds successfully.
- Notes:
  - Runtime binary output name can remain unchanged.

### FR-003 - Compose project prefix naming

- Description:
  - Compose file MUST explicitly set project prefix to `payrune`.
- Acceptance criteria:
  - [ ] `deployments/compose/compose.yaml` defines compose name `payrune`.
  - [ ] `docker compose -f deployments/compose/compose.yaml config` renders successfully.
- Notes:
  - Prefix applies to generated container/network/volume names.

## Non-functional requirements

- Performance (NFR-001): Service startup time remains within existing local baseline after rename.
- Availability/Reliability (NFR-002): Existing compose service dependencies and startup sequencing remain intact.
- Security/Privacy (NFR-003): No credential/config exposure changes are introduced by path/name updates.
- Compliance (NFR-004): `SPEC_DIR="specs/2026-03-03-cmd-app-compose-prefix" bash scripts/spec-lint.sh` passes.
- Observability (NFR-005): Existing compose logs remain available under new prefixed resource names.
- Maintainability (NFR-006): Command naming aligns with current compose service key `app`.

## Dependencies and integrations

- External systems:
  - Docker Compose v2
- Internal services:
  - existing compose services (`app`, `postgres`, `migrate`, `swagger`)
