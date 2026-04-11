---
doc: 03_tasks
spec_date: 2026-04-11
slug: runtime-defaults-alignment
mode: Quick
status: DONE
owners:
  - codex
depends_on: []
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale:
  - This is a small checked-in default change across existing config/runtime files; no new flow, schema, or integration is added.
- Upstream dependencies (`depends_on`):
  - `[]`
- Dependency gate before `READY`:
  - No prerequisite specs.
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - No new design work is needed for a default-value update.
  - What would trigger switching to Full mode:
    - If the request expanded into changing runtime flow semantics rather than just defaults.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Validation steps are listed under each task below.

## Milestones

- M1:
  - Update poller reschedule defaults to `5m`.
- M2:
  - Update Sepolia confirmations defaults to `12`.
- M3:
  - Validate the combined default-alignment change.

## Tasks (ordered)

1. T-001 - Update poller reschedule checked-in defaults
   - Scope:
     - Update Compose poller defaults, env example values, Cloudflare poller defaults, and worker fallback from `10m` to `5m`.
   - Output:
     - Aligned `5m` poller reschedule defaults across checked-in sources.
   - Linked requirements: FR-001, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command):
       - Inspect `deployments/compose/compose.yaml`, `deployments/compose/compose.env.example`, `deployments/cloudflare/payrune/wrangler.toml`, and `internal/bootstrap/poller_worker.go`.
     - [ ] Expected result:
       - Every checked-in poller reschedule default uses `5m`.
     - [ ] Logs/metrics to check (if applicable):
       - Not applicable.
2. T-002 - Update Sepolia confirmations checked-in defaults
   - Scope:
     - Update Compose defaults, env example values, Cloudflare vars, and bootstrap fallback from `1` to `12`.
   - Output:
     - Aligned `12` Sepolia confirmations defaults across checked-in sources.
   - Linked requirements: FR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command):
       - Inspect `deployments/compose/compose.yaml`, `deployments/compose/compose.env.example`, `deployments/cloudflare/payrune/wrangler.toml`, and `internal/bootstrap/api.go`.
     - [ ] Expected result:
       - Every checked-in Sepolia confirmations default uses `12`.
     - [ ] Logs/metrics to check (if applicable):
       - Not applicable.
3. T-003 - Run focused validation and repo checks
   - Scope:
     - Run the focused tests and config validation for the combined default changes.
   - Output:
     - Passing validation for both updated defaults.
   - Linked requirements: FR-003, NFR-001, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command):
       - `go test ./internal/bootstrap`
       - `docker compose --env-file deployments/compose/compose.env.example -f deployments/compose/compose.yaml config`
       - `SPEC_DIR="specs/2026-04-11-runtime-defaults-alignment" bash scripts/spec-lint.sh`
       - `bash scripts/precommit-run.sh`
     - [ ] Expected result:
       - Tests and repo checks pass after both default changes.
     - [ ] Logs/metrics to check (if applicable):
       - Not applicable.

## Traceability (optional)

- FR-001 -> T-001, T-003
- FR-002 -> T-002, T-003
- FR-003 -> T-003
- NFR-001 -> T-003
- NFR-002 -> T-003
- NFR-006 -> T-001, T-002

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Restore the previous checked-in defaults if either updated value is rejected.

## Validation evidence

- `go test ./internal/bootstrap`
- `docker compose --env-file deployments/compose/compose.env.example -f deployments/compose/compose.yaml config`
- `SPEC_DIR="specs/2026-04-11-runtime-defaults-alignment" bash scripts/spec-lint.sh`
- `bash scripts/precommit-run.sh`
