---
doc: 03_tasks
spec_date: 2026-03-15
slug: infrastructure-driver-extraction
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

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale:
  - The change is a small infrastructure-boundary refactor with no new runtime contract or
    integration.
- Upstream dependencies (`depends_on`):
  - None.
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - The only targeted move is low-level PostgreSQL connection setup.
  - What would trigger switching to Full mode:
    - Any expansion into Cloudflare runtime drivers or broader persistence adapter restructuring.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Not skipped.

## Milestones

- M1:
  - Add a concrete PostgreSQL driver helper under infrastructure.
- M2:
  - Repoint standalone DI containers to the new helper and verify behavior.
- M3:
  - Move the Cloudflare Postgres raw bridge implementation under infrastructure drivers and update
    Cloudflare DI wiring.
- M4:
  - Move the Cloudflare webhook raw bridge implementation under infrastructure drivers and update
    Cloudflare DI wiring.

## Tasks (ordered)

1. T-001 - Add PostgreSQL driver helper
   - Scope:
     - Create a concrete helper package under `internal/infrastructure/drivers` that opens and
       pings a PostgreSQL connection.
   - Output:
     - New PostgreSQL driver package and unit tests.
   - Linked requirements: FR-001, FR-002, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): run targeted Go tests for
           `./internal/infrastructure/drivers/...`.
     - [ ] Expected result: the helper returns clear errors for missing DSN or ping failure and
           returns a usable `*sql.DB` on success.
     - [ ] Logs/metrics to check (if applicable): none
2. T-002 - Rewire standalone DI containers to use the driver helper

   - Scope:
     - Update standalone API, poller, and receipt webhook dispatcher containers to obtain their
       database handle through the driver helper.
   - Output:
     - Leaner DI containers without repeated open-and-ping boilerplate.
   - Linked requirements: FR-001, FR-002, FR-003, NFR-001, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): run targeted DI/bootstrap command tests,
           `go list ./...`, and full `go test ./...`.
     - [ ] Expected result: standalone runtimes behave the same and the repo remains green.
     - [ ] Logs/metrics to check (if applicable): none

3. T-003 - Move the Cloudflare Postgres raw bridge implementation into infrastructure drivers

   - Scope:
     - Relocate the raw Cloudflare Postgres `NewJSBridge` implementation into an infrastructure
       driver package and update Cloudflare DI builders to inject it into the adapter package.
   - Output:
     - Infrastructure driver package for the raw Cloudflare Postgres bridge and updated Cloudflare
       runtime builders.
   - Linked requirements: FR-004, FR-005, NFR-001, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): run targeted tests for Cloudflare DI and
           Cloudflare Postgres packages, `go list ./...`, and full `go test ./...`.
     - [ ] Expected result: Cloudflare runtimes compile and tests pass with the raw bridge no
           longer living in the adapter package.
     - [ ] Logs/metrics to check (if applicable): none

4. T-004 - Move the Cloudflare webhook raw bridge implementation into infrastructure drivers
   - Scope:
     - Relocate the raw Cloudflare webhook `NewCloudflarePaymentReceiptStatusWebhookBridge`
       implementation into an infrastructure driver package and update Cloudflare webhook DI
       wiring to inject it into the notifier adapter.
   - Output:
     - Infrastructure driver package for the raw Cloudflare webhook bridge and updated Cloudflare
       webhook runtime builder.
   - Linked requirements: FR-006, FR-007, NFR-001, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): run targeted tests for Cloudflare webhook
           adapter, Cloudflare DI, infrastructure drivers, `go list ./...`, and full `go test ./...`.
     - [ ] Expected result: the raw webhook bridge no longer lives in the adapter package and the
           webhook dispatcher runtime still compiles and tests cleanly.
     - [ ] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001, T-002
- FR-002 -> T-001, T-002
- FR-003 -> T-002
- FR-004 -> T-003
- FR-005 -> T-003
- FR-006 -> T-004
- FR-007 -> T-004
- NFR-001 -> T-002, T-003, T-004
- NFR-002 -> T-001, T-002, T-003, T-004

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - Add the standalone PostgreSQL helper first, then switch DI containers, then move the Cloudflare
    raw Postgres bridge, then move the raw webhook bridge and rewire Cloudflare DI.
- Rollback steps:
  - Restore direct connection setup in DI containers or move the raw Cloudflare bridge back into
    the adapter package if the extraction adds confusion or breaks runtime startup.

## Validation evidence

- `SPEC_DIR="specs/2026-03-15-infrastructure-driver-extraction" bash scripts/spec-lint.sh`
- `gofmt -w internal/infrastructure/drivers/postgres/connection.go internal/infrastructure/drivers/postgres/connection_test.go internal/infrastructure/drivers/cloudflarepostgres/js_bridge_js_wasm.go internal/infrastructure/drivers/cloudflarepostgres/js_bridge_unsupported.go internal/infrastructure/drivers/cloudflarepostgres/js_bridge_unsupported_test.go internal/infrastructure/drivers/cloudflarewebhook/bridge_js_wasm.go internal/infrastructure/drivers/cloudflarewebhook/bridge_unsupported.go internal/infrastructure/drivers/cloudflarewebhook/bridge_unsupported_test.go internal/infrastructure/di/container.go internal/infrastructure/di/poller_container.go internal/infrastructure/di/receipt_webhook_dispatcher_container.go internal/infrastructure/di/cloudflare_api_worker.go internal/infrastructure/di/cloudflare_poller_worker.go internal/infrastructure/di/cloudflare_webhook_dispatcher_worker.go`
- `go test ./internal/infrastructure/drivers/... ./internal/adapters/outbound/persistence/cloudflarepostgres ./internal/adapters/outbound/webhook ./internal/infrastructure/di ./internal/bootstrap ./cmd/api ./cmd/api-worker ./cmd/poller ./cmd/poller-worker ./cmd/webhook-dispatcher ./cmd/webhook-dispatcher-worker`
- `go list ./...`
- `go test ./...`
- `rg -n "sql\\.Open\\(|DATABASE_URL is required" internal/infrastructure/di internal/infrastructure/drivers -g '*.go'`
- `rg -n "func NewJSBridge|jsFnBeginTx" internal/adapters/outbound/persistence/cloudflarepostgres internal/infrastructure/drivers/cloudflarepostgres -g '*.go'`
- `rg -n "NewCloudflarePaymentReceiptStatusWebhookBridge\\(|jsFnWebhookPost" internal cmd -g '*.go'`
