---
doc: 02_design
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

# Design

## Summary

- This is a repository-wide mechanical rename with no intended behavior change.
- The rename is constrained to runtime/deployment naming, helper script naming, selected internal
  package-oriented directories, `AGENTS.md` path alignment, and current spec/doc references.
- Architecture package names such as `application` remain unchanged.

## Rename map

### API service

- `cmd/app` -> `cmd/api`
- `build/app` -> `build/api`
- compose service key `app` -> `api`

### Tracked underscore directories

- `cmd/api_worker` -> `cmd/api-worker`
- `cmd/poller_worker` -> `cmd/poller-worker`
- `cmd/webhook_dispatcher` -> `cmd/webhook-dispatcher`
- `cmd/fake_webhook_receiver` -> `cmd/fake-webhook-receiver`
- `internal/application/use_cases` -> `internal/application/use-cases`
- `internal/domain/value_objects` -> `internal/domain/value-objects`

### Compact internal package-oriented directories

- `internal/application/use-cases` -> `internal/application/usecases`
- `internal/domain/value-objects` -> `internal/domain/valueobjects`
- `internal/application/ports/in` -> `internal/application/ports/inbound`
- `internal/application/ports/out` -> `internal/application/ports/outbound`

### Package-name alignment

- `package use_cases` -> `package usecases`
- `package value_objects` -> `package valueobjects`
- `package in` -> `package inbound`
- `package out` -> `package outbound`

### Tracked underscore script filenames

- `scripts/build-cf-api_worker-wasm.sh` -> `scripts/build-cf-api-worker-wasm.sh`
- `scripts/build-cf-poller_worker-wasm.sh` -> `scripts/build-cf-poller-worker-wasm.sh`
- `scripts/cf-api_worker-deploy.sh` -> `scripts/cf-api-worker-deploy.sh`
- `scripts/cf-api_worker-delete.sh` -> `scripts/cf-api-worker-delete.sh`
- `scripts/cf-poller_worker-deploy.sh` -> `scripts/cf-poller-worker-deploy.sh`
- `scripts/cf-poller_worker-delete.sh` -> `scripts/cf-poller-worker-delete.sh`

## Code placement rules

- The rename preserves existing architecture boundaries.
- No domain/application logic moves across layers.
- Only runtime/deployment naming and references are updated.

## Documentation alignment

- `AGENTS.md` is treated as repo-local source of truth and must point at current internal paths.
- Generic examples such as `cmd/<app>/main.go` remain unchanged where they are still structurally
  correct.

## Failure modes

- Missing one side of a rename causes broken build, deploy, or import paths.
- Missing compose service-key updates breaks service references and `depends_on`.
- Missing current spec/doc updates leaves stale operational instructions.
- Missing package-name alignment after directory rename leaves inconsistent imports and harder-to-
  follow code.

## Observability

- No observability behavior change.
- Validation relies on targeted `go test`, `spec-lint`, and repo-wide `rg` checks.
