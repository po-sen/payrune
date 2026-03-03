---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Migration command: Go executable at `cmd/migrate/main.go`.
- Migration runner: Compose service that runs the migration command then exits.

## Out-of-scope behaviors

- OOS1: Application repositories/ORM integration.
- OOS2: DB backup/restore workflows.

## Functional requirements

### FR-001 - PostgreSQL 18 compose service

- Description:
  - Compose stack MUST provide a PostgreSQL 18 container with persistent storage and healthcheck.
- Acceptance criteria:
  - [ ] `deployments/compose/compose.yaml` defines `postgres` with image `postgres:18`.
  - [ ] `postgres` has data volume mounted and exposes port `5432`.
  - [ ] `postgres` healthcheck verifies readiness via `pg_isready`.
- Notes:
  - Service restart behavior should match local reliability expectations.

### FR-002 - Migration SQL structure

- Description:
  - Repository MUST include versioned SQL migration files with up/down pairs.
- Acceptance criteria:
  - [ ] Migration files exist under a dedicated folder in deployment tree.
  - [ ] At least one baseline migration pair (`*.up.sql`, `*.down.sql`) exists.
  - [ ] Baseline migration creates and rollback drops the same schema object(s).
- Notes:
  - File naming must be compatible with golang-migrate conventions.

### FR-003 - golang-migrate command under cmd/

- Description:
  - Migration execution MUST be implemented in a decoupled command located under `cmd/` using `github.com/golang-migrate/migrate/v4`.
- Acceptance criteria:
  - [ ] `cmd/migrate/main.go` exists and compiles.
  - [ ] Command reads DB connection and migration source configuration from environment.
  - [ ] Command can execute `up` migration action and returns non-zero on unrecoverable error.
  - [ ] `go.mod` includes golang-migrate dependency.
- Notes:
  - Keep command independent from application runtime bootstrap (`cmd/payrune`).

### FR-004 - Migration runner container sequencing

- Description:
  - Compose MUST include a migration runner container that starts only after DB is healthy and executes migration command.
- Acceptance criteria:
  - [ ] Compose defines `migrate` service building/running migration command image.
  - [ ] `migrate` depends on `postgres` healthy condition.
  - [ ] `migrate` exits successfully when migration is applied or already up-to-date.
- Notes:
  - Runner should not restart endlessly after successful completion.

## Non-functional requirements

- Performance (NFR-001): Migration runner should complete baseline migration within 30 seconds after DB is healthy.
- Availability/Reliability (NFR-002): Postgres service should use restart policy `unless-stopped`.
- Security/Privacy (NFR-003): DB credentials for local dev are explicit and scoped to local compose network only.
- Compliance (NFR-004): `SPEC_DIR="specs/2026-03-03-postgresql18-migration-runner-container" bash scripts/spec-lint.sh` passes.
- Observability (NFR-005): Migration runner logs clearly indicate migration status (applied/no-change/error).
- Maintainability (NFR-006): Migration command and SQL assets are organized for future incremental schema versions.

## Dependencies and integrations

- External systems:
  - Docker image `postgres:18`
  - Go module `github.com/golang-migrate/migrate/v4`
- Internal services:
  - compose-managed `postgres` service
