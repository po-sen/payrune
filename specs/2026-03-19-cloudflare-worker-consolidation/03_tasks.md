---
doc: 03_tasks
spec_date: 2026-03-19
slug: cloudflare-worker-consolidation
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-11-cloudflare-poller-workers
  - 2026-03-12-api-worker-naming-unification
  - 2026-03-13-cloudflare-webhook-dispatcher-worker
  - 2026-03-14-runtime-entrypoint-alignment
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
  - The change alters Cloudflare deployment/runtime boundaries, cron routing, Wasm entrypoints,
    deployment scripts, and active operational docs. It is broader than a small runtime-shell
    tweak and needs an explicit design plus test plan.
- Upstream dependencies (`depends_on`):
  - `2026-03-11-cloudflare-poller-workers`
  - `2026-03-12-api-worker-naming-unification`
  - `2026-03-13-cloudflare-webhook-dispatcher-worker`
  - `2026-03-14-runtime-entrypoint-alignment`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Not applicable; Full mode is required.
- If `04_test_plan.md` is skipped:
  - Not applicable; Full mode requires `04_test_plan.md`.

## Milestones

- M1:
  - Unified spec and runtime design are in place and lint-clean.
- M2:
  - Unified worker shell, scripts, docs, and tests replace the split API/poller/dispatcher worker
    shells.
- M3:
  - Verification passes and the spec can be marked `DONE`.

## Tasks (ordered)

1. T-001 - Finalize the unified worker spec package

   - Scope:
     - Fill the problem, requirements, design, tasks, and test-plan docs for the consolidated
       Cloudflare worker shape and validate dependency readiness.
   - Output:
     - `specs/2026-03-19-cloudflare-worker-consolidation/*.md`
   - Linked requirements: FR-001, FR-003, FR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-19-cloudflare-worker-consolidation" bash scripts/spec-lint.sh`
     - [ ] Expected result: spec-lint passes and all produced docs stay frontmatter-consistent.
     - [ ] Logs/metrics to check (if applicable): none.

2. T-002 - Add the unified Wasm entrypoint and deployment shell

   - Scope:
     - Introduce the unified Go/Wasm worker entrypoint, unified runtime loader, unified
       `deployments/cloudflare/payrune/` shell, and runtime-specific bridge dispatch for API,
       poller, and dispatcher.
   - Output:
     - New unified Cloudflare worker deployment directory and Wasm build script.
     - Updated JS tests covering API fetch path plus scheduled-job routing.
     - Removal or replacement of the split API/poller/dispatcher deployment shells as active
       deployment targets.
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, NFR-001, NFR-002, NFR-003, NFR-005,
     NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): run targeted Go tests for the unified worker
           command and JS tests for the unified Cloudflare deployment shell.
     - [ ] Expected result: the unified worker can route API, poller, and dispatcher operations
           without changing the existing bootstrap contracts.
     - [ ] Logs/metrics to check (if applicable): poller and dispatcher log-message tests keep
           their current summary format.

3. T-003 - Consolidate deploy/delete automation and docs

   - Scope:
     - Update `Makefile`, Cloudflare deploy/delete scripts, and active docs/README content to use
       the unified payrune worker while keeping `receipt-webhook-mock` explicit.
   - Output:
     - Updated top-level deploy/delete flow.
     - Unified payrune worker README.
     - Updated active references in top-level operational docs.
   - Linked requirements: FR-001, FR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): search for active references to the split
           worker deploy flow and run the unified worker package checks/tests.
     - [ ] Expected result: `cf-up` / `cf-down` wiring references one payrune worker deploy/delete
           script pair plus the mock worker pair.
     - [ ] Logs/metrics to check (if applicable): none.

4. T-004 - Verify the consolidated runtime and close the spec
   - Scope:
     - Run targeted formatting/tests/search checks, record validation evidence, and update the spec
       status to `DONE` if implementation passes.
   - Output:
     - Passing targeted verification commands and final spec status update.
   - Linked requirements: FR-002, FR-003, FR-004, FR-005, NFR-001, NFR-002, NFR-003, NFR-005,
     NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): run the commands listed in `04_test_plan.md`
           plus targeted repo searches for removed split deployment references.
     - [ ] Expected result: tests pass, active references are updated, and the spec can move to
           `DONE`.
     - [ ] Logs/metrics to check (if applicable): confirm test output covers API, poller, and
           dispatcher runtime paths.

## Traceability (optional)

- FR-001 -> T-001, T-002, T-003
- FR-002 -> T-002, T-004
- FR-003 -> T-001, T-002, T-004
- FR-004 -> T-002, T-004
- FR-005 -> T-001, T-003, T-004
- NFR-001 -> T-002, T-004
- NFR-002 -> T-002, T-004
- NFR-003 -> T-002, T-004
- NFR-005 -> T-002, T-004
- NFR-006 -> T-001, T-002, T-003

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - No database migration change is required.
  - Deploy `receipt-webhook-mock` first, then deploy the unified payrune worker.
- Rollback steps:
  - Restore the split worker deployment directories and lifecycle scripts from git if unified
    deployment verification fails.
  - Redeploy the previously split workers if rollback is required in a live environment.
