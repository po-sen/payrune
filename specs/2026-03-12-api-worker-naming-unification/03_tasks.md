---
doc: 03_tasks
spec_date: 2026-03-12
slug: api-worker-naming-unification
mode: Full
status: DONE
owners:
  - payrune-team
depends_on: []
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Tasks

## Mode decision

- Selected mode: Full
- Rationale:
  - The rename is broad and cross-cuts `cmd/`, `build/`, `internal/`, `deployments/`, `scripts/`,
    and `specs/`.
  - It needs explicit rename mapping and verification scope.

## Implementation tasks

- [x] T-001 Rename API runtime paths from `app` to `api`.
      Links to: FR-001, NFR-001

  - Update command path, build path, and compose service-key references.
  - Validate with targeted search and build/test commands.

- [x] T-002 Rename tracked underscore directories to hyphenated names.
      Links to: FR-002, NFR-001

  - Rename worker command directories plus existing underscored `cmd/` and `internal/`
    directories.
  - Update Go imports and build paths to the new hyphenated directories.

- [x] T-003 Rename compact internal package-oriented directories and align package names.
      Links to: FR-003, NFR-001

  - Rename `use-cases`, `value-objects`, `ports/in`, and `ports/out`.
  - Update package declarations and imports to `usecases`, `valueobjects`, `inbound`, and
    `outbound`.

- [x] T-004 Update references to the renamed tracked paths.
      Links to: FR-004, NFR-001

  - Align build scripts, helper script filenames, deploy docs, and runtime references with the
    renamed tracked paths.

- [x] T-005 Update affected current docs/specs to the new names.
      Links to: FR-005, NFR-002, NFR-003

  - Update current Cloudflare specs, this spec, and `AGENTS.md`.
  - Do not modify unrelated historical specs.

- [x] T-006 Run verification and mark spec complete.
      Links to: NFR-001, NFR-002, NFR-003
  - Run spec-lint for this spec and related current Cloudflare specs.
  - Run targeted Go tests and repo-wide `rg` checks for stale names.

## Validation notes

- `rg` checks should be run against the exact stale names being removed.
- Validation should focus on tracked runtime/deployment directory names, not generic architecture
  words like `application`.

## Validation evidence

- `SPEC_DIR="specs/2026-03-12-api-worker-naming-unification" bash scripts/spec-lint.sh`
- `GOCACHE=/tmp/go-build go test ./internal/infrastructure/di ./internal/adapters/inbound/cloudflareworker ./cmd/api ./cmd/api-worker ./cmd/poller ./cmd/poller-worker ./cmd/webhook-dispatcher ./cmd/fake-webhook-receiver -count=1`
- `cd deployments/cloudflare/payrune-api && npm test`
- `cd deployments/cloudflare/payrune-poller && npm test`
- `find internal/application -maxdepth 3 -type d | sort`
- `find internal/domain -maxdepth 2 -type d | sort`
- `rg -n 'internal/application/use-cases|internal/domain/value-objects|internal/application/ports/in\\b|internal/application/ports/out\\b|package use_cases|package value_objects|\\bvalue_objects\\b|\\buse_cases\\b|^package in$|^package out$' cmd internal deployments scripts Makefile specs/2026-03-10-cloudflare-workers-postgres specs/2026-03-11-cloudflare-poller-workers specs/2026-03-12-api-worker-naming-unification`
- `rg -n 'use_cases|value_objects|ports/in\\b|ports/out\\b' AGENTS.md`
