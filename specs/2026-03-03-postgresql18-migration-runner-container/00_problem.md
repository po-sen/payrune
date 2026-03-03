---
doc: 00_problem
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

# Problem & Goals

## Context

- Background: Current local stack has application and Swagger containers, but no PostgreSQL service or migration lifecycle.
- Users or stakeholders: Developers needing deterministic local DB bootstrap for feature development and testing.
- Why now: We need PostgreSQL 18 ready in compose and an automated migration runner container using `github.com/golang-migrate/migrate/v4`.

## Constraints (optional)

- Technical constraints:
  - PostgreSQL container image version must be 18.
  - Migration execution must be implemented via Go command under `cmd/`.
  - Migration runner must execute after DB is healthy.
  - Migration library must be `github.com/golang-migrate/migrate/v4`.
- Timeline/cost constraints:
  - Keep changes focused to local deployment/bootstrap path.
- Compliance/security constraints:
  - Keep spec-first workflow and preserve Clean Architecture boundaries.

## Problem statement

- Current pain:
  - No local PostgreSQL container is defined.
  - No migration structure or SQL versioning exists.
  - No automated DB-ready migration execution step exists in compose lifecycle.
- Evidence or examples:
  - `deployments/compose/compose.yaml` currently defines app/swagger only.

## Goals

- G1: Add PostgreSQL 18 service container to local compose stack.
- G2: Add migration architecture using `golang-migrate` with command entrypoint under `cmd/`.
- G3: Add migration runner container that runs after DB healthcheck passes.

## Non-goals (out of scope)

- NG1: Production database HA/replication setup.
- NG2: Integrating application runtime persistence layer in this change.

## Assumptions

- A1: Local default DB credentials can be static non-secret development values.
- A2: Initial migration can bootstrap a baseline table used for schema smoke verification.

## Open questions

- Q1: Should migration command later support configurable actions (up/down/force) via CLI args?
- Q2: Should DB credentials be moved to `.env` once more services require them?

## Success metrics

- Metric: Database availability.
- Target: `postgres:18` container reaches healthy state in compose.
- Metric: Migration automation.
- Target: migration runner container exits successfully after applying migrations.
- Metric: Schema bootstrap.
- Target: baseline migration table exists after `make up`.
