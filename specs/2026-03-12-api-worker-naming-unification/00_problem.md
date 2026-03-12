---
doc: 00_problem
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

# Problem

## Summary

- The repository currently mixes `app`, `api`, hyphenated tracked paths, underscored tracked
  paths, and internal package directories whose names no longer match the preferred package naming
  style.
- The intended naming baseline is:
  - the main HTTP runtime uses `api`
  - tracked underscore paths use `-`, not `_`
  - selected internal package-oriented directories use compact names without separators
- The previous underscore-based worker directory rename was incorrect and must be reversed.

## Why now

- Runtime and deployment paths are currently inconsistent and harder to predict.
- The recent rename moved some directories to `_`, but the desired directory convention is `-`.
- This should be corrected before more API or worker features are added.

## In scope

- Keep the API runtime rename from `app` to `api`.
- Rename tracked directories and script filenames that currently use `_` so they use `-` instead.
- Rename selected internal package-oriented directories to compact names:
  - `internal/application/usecases`
  - `internal/domain/valueobjects`
  - `internal/application/ports/inbound`
  - `internal/application/ports/outbound`
- Update code, import paths, package names, docs, `AGENTS.md`, build paths, and current spec
  references that point to the renamed tracked paths.

## Out of scope

- Renaming generic words like `application` in architecture docs.
- Changing domain/application package names unless required by the directory rename.
- Editing historical specs beyond what is needed for this new spec.
- Behavioral changes to API, poller, webhook, or migrations.

## Goals

- G1: Repository runtime naming is consistent around `api`, hyphenated tracked paths, and compact
  internal package-oriented directory names.
- G2: Build, deploy, import, and compose paths remain functional after the rename.
- G3: Current docs/specs and `AGENTS.md` no longer point to stale tracked paths.

## Success metrics

- `rg` finds no stale tracked-path references for:
  - `cmd/app`
  - `build/app`
  - `cmd/api_worker`
  - `cmd/poller_worker`
  - `cmd/webhook_dispatcher`
  - `cmd/fake_webhook_receiver`
  - `internal/application/use_cases`
  - `internal/domain/value_objects`
  - `internal/application/use-cases`
  - `internal/domain/value-objects`
  - `internal/application/ports/in`
  - `internal/application/ports/out`
  - `scripts/build-cf-api_worker-wasm.sh`
  - `scripts/build-cf-poller_worker-wasm.sh`
  - `scripts/cf-api_worker-deploy.sh`
  - `scripts/cf-api_worker-delete.sh`
  - `scripts/cf-poller_worker-deploy.sh`
  - `scripts/cf-poller_worker-delete.sh`
  - `use_cases` in `AGENTS.md`
  - `value_objects` in `AGENTS.md`
  - `ports/in` in `AGENTS.md`
  - `ports/out` in `AGENTS.md`
- `go test` for touched Go packages passes.
- This spec and the related Cloudflare specs pass `spec-lint`.
