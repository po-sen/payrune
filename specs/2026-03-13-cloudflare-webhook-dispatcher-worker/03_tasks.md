---
doc: 03_tasks
spec_date: 2026-03-13
slug: cloudflare-webhook-dispatcher-worker
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-06-receipt-webhook-delivery
  - 2026-03-10-cloudflare-workers-postgres
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
  - This adds a new Cloudflare runtime, deployment shell, Go/Wasm entrypoint, and outbound delivery
    bridge.
- Upstream dependencies (`depends_on`):
  - `2026-03-06-receipt-webhook-delivery`
  - `2026-03-10-cloudflare-workers-postgres`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - Not safe; this change adds a new runtime path.
  - What would trigger switching to Full mode:
    - Already Full.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Not skipped.

## Milestones

- M1:
  - Add Cloudflare dispatcher worker runtime and binding-only notifier path.
- M2:
  - Add mock worker plus deployment shell and automation.
- M3:
  - Validate worker runtime, binding path, and deployment config.

## Tasks (ordered)

1. T-001 - Add Go/Wasm dispatcher worker runtime

   - Scope:
     - Introduce a Cloudflare-compatible webhook dispatcher worker entrypoint and DI wiring that
       reuse the existing Go dispatch use case.
   - Output:
     - `cmd/webhook-dispatcher-worker/`
     - `internal/infrastructure/di/cloudflare_webhook_dispatcher_worker.go`
     - inbound handler additions under `internal/adapters/inbound/cloudflareworker/`
   - Linked requirements: FR-001, FR-002, NFR-004
   - Validation:
     - [ ] How to verify (manual steps or command): `go test` for worker runtime packages.
     - [ ] Expected result: scheduled handler can execute one dispatch cycle via Go use case.
     - [ ] Logs/metrics to check (if applicable): none

1. T-002 - Add Cloudflare-compatible binding-only webhook notifier

   - Scope:
     - Introduce the outbound adapter or bridge needed for webhook delivery inside the Cloudflare
       worker runtime, fixed to the internal mock service binding path, and keep the Cloudflare
       PostgreSQL outbox scanner compatible with webhook time columns.
   - Output:
     - Cloudflare notifier bridge/adapter implementation plus tests.
   - Linked requirements: FR-003, NFR-001, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): unit/integration tests for sent / retry /
           failed paths plus Cloudflare PostgreSQL scan tests.
     - [ ] Expected result: delivery result semantics match existing use case behavior and the path
           does not rely on a fake internal URL sentinel, and outbox rows with time columns scan
           successfully in the worker runtime.
     - [ ] Logs/metrics to check (if applicable): none

1. T-003 - Add receipt-webhook-mock worker and deploy automation

   - Scope:
     - Create the internal mock worker and wire deploy/delete scripts plus top-level orchestration
       order for mock then dispatcher.
   - Output:
     - `deployments/cloudflare/receipt-webhook-mock/`
     - `scripts/cf-receipt-webhook-mock-worker-deploy.sh`
     - `scripts/cf-receipt-webhook-mock-worker-delete.sh`
     - updated `Makefile`
   - Linked requirements: FR-004, FR-005, NFR-003
   - Validation:
     - [ ] How to verify (manual steps or command): `npm test` and `wrangler deploy --dry-run`
           for the mock worker plus `make -n`.
     - [ ] Expected result: mock worker deploys cleanly and orchestration order is correct.
     - [ ] Logs/metrics to check (if applicable): invocation logs enabled by config.

1. T-004 - Add dispatcher deployment shell and automation

   - Scope:
     - Create dispatcher Cloudflare deployment shell and wire deploy/delete scripts.
   - Output:
     - `deployments/cloudflare/payrune-webhook-dispatcher/`
     - `scripts/cf-webhook-dispatcher-worker-deploy.sh`
     - `scripts/cf-webhook-dispatcher-worker-delete.sh`
   - Linked requirements: FR-001, FR-005, FR-006, NFR-003, NFR-004
   - Validation:
     - [ ] How to verify (manual steps or command): `wrangler deploy --dry-run` and `make -n`.
     - [ ] Expected result: dispatcher worker can be deployed and deleted with repo automation.
     - [ ] Logs/metrics to check (if applicable): invocation logs enabled by config.

1. T-005 - Document runtime boundaries and fixed binding usage

   - Scope:
     - Document that dispatcher depends only on the internal mock worker binding and does not
       expose or require a public webhook URL in the Cloudflare runtime.
   - Output:
     - Updated dispatcher README and spec finalization.
   - Linked requirements: FR-006, NFR-004, NFR-005
   - Validation:
     - [ ] How to verify (manual steps or command): doc review plus spec lint.
     - [ ] Expected result: runtime responsibilities are explicit and deployment shell remains
           thin.
     - [ ] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001, T-004
- FR-002 -> T-001
- FR-003 -> T-002
- FR-004 -> T-003
- FR-005 -> T-003, T-004
- FR-006 -> T-004, T-005
- NFR-001 -> T-002
- NFR-002 -> T-002
- NFR-003 -> T-003, T-004
- NFR-004 -> T-001, T-004, T-005
- NFR-005 -> T-005

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - No schema change; existing `cf-migrate` flow remains unchanged.
- Rollback steps:
  - Delete the dispatcher worker deployment and continue using the existing process runtime.
