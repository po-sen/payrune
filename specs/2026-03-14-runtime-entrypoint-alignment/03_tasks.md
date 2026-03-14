---
doc: 03_tasks
spec_date: 2026-03-14
slug: runtime-entrypoint-alignment
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-10-cloudflare-workers-postgres
  - 2026-03-11-cloudflare-poller-workers
  - 2026-03-12-api-worker-naming-unification
  - 2026-03-13-cloudflare-webhook-dispatcher-worker
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
  - This refactor crosses multiple runtime entrypoints, transport boundaries, DI call paths, and
    bootstrap naming conventions.
- Upstream dependencies (`depends_on`):
  - `2026-03-10-cloudflare-workers-postgres`
  - `2026-03-11-cloudflare-poller-workers`
  - `2026-03-12-api-worker-naming-unification`
  - `2026-03-13-cloudflare-webhook-dispatcher-worker`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`

## Milestones

- M1:
  - Establish shared HTTP and scheduler inbound boundaries.
- M2:
  - Route standalone and worker runtimes through the shared bootstrap/inbound structure.
- M3:
  - Normalize bootstrap naming and validate the final runtime entrypoint shape.

## Tasks (ordered)

1. T-001 - Establish shared inbound runtime boundaries

   - Scope:
     - Add a shared public HTTP handler assembly point.
     - Move the Cloudflare API bridge under HTTP inbound.
     - Move scheduler cycle mapping into a dedicated scheduler inbound package.
   - Output:
     - Shared HTTP handler.
     - HTTP-scoped Cloudflare bridge.
     - Scheduler package and updated runtime DI/imports.
   - Linked requirements: FR-001, FR-002, FR-005, NFR-002, NFR-003
   - Validation:
     - [ ] How to verify (manual steps or command): run targeted Go tests for inbound HTTP,
           scheduler, API worker, and DI packages.
     - [ ] Expected result: HTTP and scheduler runtimes compile and preserve contracts.
     - [ ] Logs/metrics to check (if applicable): none

2. T-002 - Reuse scheduler handlers in standalone runtimes

   - Scope:
     - Update standalone poller and receipt webhook dispatcher bootstrap flows to execute each cycle
       through scheduler handlers exposed by DI.
   - Output:
     - Updated standalone bootstrap files and DI containers.
   - Linked requirements: FR-002, FR-005, NFR-001, NFR-003
   - Validation:
     - [ ] How to verify (manual steps or command): run targeted Go tests for scheduler, bootstrap,
           DI, `cmd/poller`, and `cmd/webhook-dispatcher`.
     - [ ] Expected result: bootstrap keeps lifecycle ownership while scheduler mapping is reused.
     - [ ] Logs/metrics to check (if applicable): none

3. T-003 - Thin worker command entrypoints through bootstrap

   - Scope:
     - Move worker payload orchestration into bootstrap so worker `cmd/*` packages remain thin.
   - Output:
     - Bootstrap worker helper files/functions and updated worker command `main.go` files.
   - Linked requirements: FR-003, FR-005, NFR-001, NFR-003
   - Validation:
     - [ ] How to verify (manual steps or command): run targeted Go tests for worker commands and
           bootstrap, plus import/search checks.
     - [ ] Expected result: worker commands no longer import inbound adapter or DI packages
           directly.
     - [ ] Logs/metrics to check (if applicable): none

4. T-004 - Normalize bootstrap naming
   - Scope:
     - Rename bootstrap files, exported functions, and config types so `api`, `poller`, and
       `receipt webhook dispatcher` nouns are consistent.
   - Output:
     - Updated bootstrap names and updated direct call sites/tests.
   - Linked requirements: FR-004, FR-005, NFR-001, NFR-003
   - Validation:
     - [ ] How to verify (manual steps or command): run targeted tests for renamed bootstrap/cmd
           packages, `rg` checks for old names, and full `go test ./...`.
     - [ ] Expected result: old mixed names are removed and the repo remains green.
     - [ ] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-001, T-002
- FR-003 -> T-003
- FR-004 -> T-004
- FR-005 -> T-001, T-002, T-003, T-004
- NFR-001 -> T-002, T-003, T-004
- NFR-002 -> T-001
- NFR-003 -> T-001, T-002, T-003, T-004

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - Shared inbound boundaries first, then standalone reuse, then worker bootstrap unification, then
    naming cleanup.
- Rollback steps:
  - Restore the prior package paths and symbol names if compile/test verification fails.

## Validation evidence

- `SPEC_DIR="specs/2026-03-14-runtime-entrypoint-alignment" bash scripts/spec-lint.sh`
- `go fmt ./...`
- `go list ./...`
- `go test ./internal/adapters/inbound/http/... ./internal/adapters/inbound/scheduler ./internal/infrastructure/di ./cmd/api ./cmd/api-worker ./cmd/poller-worker ./cmd/webhook-dispatcher-worker`
- `go test ./internal/adapters/inbound/scheduler ./internal/infrastructure/di ./internal/bootstrap ./cmd/poller ./cmd/webhook-dispatcher`
- `go test ./internal/bootstrap ./cmd/api ./cmd/webhook-dispatcher ./cmd/webhook-dispatcher-worker`
- `go test ./...`
- `rg -n "internal/adapters/inbound/cloudflareworker|package cloudflareworker" cmd internal`
- `rg -n "RunReceiptPollingCycleInput|RunReceiptWebhookDispatchCycleInput" internal/bootstrap`
- `rg -n "internal/adapters/inbound|internal/infrastructure/di" cmd/api-worker/main.go cmd/poller-worker/main.go cmd/webhook-dispatcher-worker/main.go`
- `rg -n "bootstrap\\.Run\\(|\\bReceiptWebhookDispatchConfig\\b|\\bloadReceiptWebhookDispatchConfigFromEnv\\b|\\bHandleCloudflareWebhookDispatcherRequestJSON\\b|internal/bootstrap/app.go|internal/bootstrap/webhook_dispatcher_worker.go" cmd internal -g '*.go'`
