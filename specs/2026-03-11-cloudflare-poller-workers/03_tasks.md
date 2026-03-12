---
doc: 03_tasks
spec_date: 2026-03-11
slug: cloudflare-poller-workers
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-05-blockchain-receipt-polling-service
  - 2026-03-09-shared-tip-height-polling
  - 2026-03-09-poller-interval-separation
  - 2026-03-09-receipt-expire-final-check
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
  - This change affects runtime model, Cloudflare scheduled execution, PostgreSQL access, Bitcoin observer integration, and future poller development boundaries.
- Upstream dependencies (`depends_on`):
  - `2026-03-05-blockchain-receipt-polling-service`
  - `2026-03-09-shared-tip-height-polling`
  - `2026-03-09-poller-interval-separation`
  - `2026-03-09-receipt-expire-final-check`
  - `2026-03-10-cloudflare-workers-postgres`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`

## Milestones

- M1:
  - Finalize Cloudflare standalone poller Worker architecture.
- M2:
  - Implement Go/Wasm scheduled runtime and Worker-compatible observer bridges.
- M3:
  - Wire mainnet/testnet4 deploy and delete flows.

## Tasks (ordered)

1. T-001 - Finalize poller Worker spec

   - Scope:
     - Capture standalone Cloudflare poller requirements, scheduled runtime model, and deployment boundaries.
   - Output:
     - `specs/2026-03-11-cloudflare-poller-workers/*.md`
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-008, NFR-001, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-11-cloudflare-poller-workers" bash scripts/spec-lint.sh`
     - [ ] Expected result: spec lint passes and docs consistently describe standalone Cloudflare pollers.
     - [ ] Logs/metrics to check (if applicable): none

2. T-002 - Create thin Cloudflare poller deployment shell

   - Scope:
     - Add `deployments/cloudflare/payrune-poller/` with Wrangler config, scheduled entry shell, Go-Wasm loader, and Worker-focused docs/tests only.
   - Output:
     - Thin deployment shell for poller Workers.
   - Linked requirements: FR-001, FR-003, FR-004, NFR-001, NFR-002, NFR-005
   - Validation:
     - [ ] How to verify (manual steps or command): inspect deployment files and run Worker syntax/tests.
     - [ ] Expected result: deployment shell is thin and supports separate mainnet/testnet4 deployment targets.
     - [ ] Logs/metrics to check (if applicable): none

3. T-003 - Implement Go-Wasm poller entrypoint and scheduled request mapping

   - Scope:
     - Add Go/Wasm scheduled entrypoint plus Cloudflare poller request mapping inside the Worker composition root so it invokes `RunReceiptPollingCycleUseCase` without a separate adapter package.
   - Output:
     - `cmd/poller-worker/`
   - Linked requirements: FR-001, FR-002, FR-003, FR-008, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): focused Go tests plus Wasm build.
     - [ ] Expected result: scheduled poller envelope reaches the existing Go use case.
     - [ ] Logs/metrics to check (if applicable): poll cycle counters present in output.

4. T-004 - Reuse Cloudflare PostgreSQL adapter for poller paths

   - Scope:
     - Extend or wire Worker-compatible PostgreSQL stores so poller claim/save operations run under Worker runtime.
   - Output:
     - Poller-compatible use of `internal/adapters/outbound/persistence/cloudflarepostgres/`
   - Linked requirements: FR-002, FR-006, NFR-004
   - Validation:
     - [ ] How to verify (manual steps or command): focused Go tests for poller paths through Cloudflare PostgreSQL adapter.
     - [ ] Expected result: due-claim and save flows run in Worker runtime without `database/sql`.
     - [ ] Logs/metrics to check (if applicable): none

5. T-005 - Implement Worker-compatible Bitcoin observer path

   - Scope:
     - Add the Worker-compatible Bitcoin Esplora observer bridge/adapter needed by the polling use case.
   - Output:
     - Worker-compatible latest tip + address observation path for Bitcoin networks.
   - Linked requirements: FR-002, FR-007, NFR-004
   - Validation:
     - [ ] How to verify (manual steps or command): focused tests for latest block height and address observation under Worker runtime.
     - [ ] Expected result: poller Worker can observe Bitcoin addresses and preserve current receipt semantics.
     - [ ] Logs/metrics to check (if applicable): observer failures still surface as deterministic poller errors.

6. T-006 - Add mainnet/testnet4 deploy and teardown flows

   - Scope:
     - Expose a unified `make cf-up` / `make cf-down` flow that stacks the migration plus both poller deploy/delete scripts, with non-sensitive defaults in Worker config, repo-local `.env.cloudflare` loading, fail-fast PostgreSQL env checks, and deploy-time secret sync limited to PostgreSQL plus optional Esplora auth secrets.
   - Output:
     - Makefile and scripts for Cloudflare poller deployment.
   - Linked requirements: FR-001, FR-003, FR-005, NFR-005
   - Validation:
     - [ ] How to verify (manual steps or command): `make -n cf-up`, `make -n cf-down`, and Worker dry-run deploys.
     - [ ] Expected result: both poller workers are included in the unified rollout and teardown flows.
     - [ ] Logs/metrics to check (if applicable): deploy output clearly states migration plan.

7. T-007 - Verify clean Cloudflare-only scope
   - Scope:
     - Remove wrong-direction assumptions and ensure the final diff stays Cloudflare-specific.
   - Output:
     - Clean final diff for poller Worker slice.
   - Linked requirements: FR-004, NFR-001, NFR-003
   - Validation:
     - [ ] How to verify (manual steps or command): `git diff --stat`, focused tests, and spec lint.
     - [ ] Expected result: no unrelated compose/process-runtime changes remain in the final diff.
     - [ ] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001, T-002, T-003, T-006
- FR-002 -> T-001, T-003, T-004, T-005
- FR-003 -> T-001, T-002, T-003, T-006
- FR-004 -> T-001, T-002, T-007
- FR-005 -> T-001, T-006
- FR-006 -> T-004
- FR-007 -> T-005
- FR-008 -> T-001, T-003
- NFR-001 -> T-001, T-002, T-007
- NFR-002 -> T-001, T-002, T-003
- NFR-003 -> T-007
- NFR-004 -> T-004, T-005
- NFR-005 -> T-002, T-006

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - PostgreSQL migrations, if any are required at deploy time, run before Worker deploy.
- Rollback steps:
  - Delete the Cloudflare poller workers and keep existing compose/VM pollers as the active runtime until the Worker path is validated.
