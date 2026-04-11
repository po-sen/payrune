---
doc: 03_tasks
spec_date: 2026-04-11
slug: cloudflare-env-location
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-09-compose-env-example
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
  - This is a focused deployment-file location cleanup affecting scripts, docs, and ignore rules, without changing runtime business logic.
- Upstream dependencies (`depends_on`):
  - `2026-04-09-compose-env-example`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - No new integration or architecture is being introduced.
  - What would trigger switching to Full mode:
    - If the change expanded into a broader Cloudflare deployment redesign.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Validation steps are listed under each task below.

## Milestones

- M1:
  - Move the checked-in Cloudflare env example and align docs.
- M2:
  - Update Cloudflare scripts to prefer the new deployment-local env path while keeping root fallback support.
- M3:
  - Validate Cloudflare helper behavior and repo checks.

## Tasks (ordered)

1. T-001 - Move the env example and update documentation
   - Scope:
     - Move `.env.cloudflare.example` into `deployments/cloudflare/cloudflare.env.example`, update README/doc references, and align ignore rules so only the new deployment-local env path stays ignored.
   - Output:
     - Cloudflare env example and docs aligned under `deployments/cloudflare/`.
   - Linked requirements: FR-001, FR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command):
       - Inspect `README.md`, `.gitignore`, and `deployments/cloudflare/**`.
     - [ ] Expected result:
       - The checked-in Cloudflare env example lives under `deployments/cloudflare/`, and docs point to the new path.
     - [ ] Logs/metrics to check (if applicable):
       - Not applicable.
2. T-002 - Update scripts to prefer the new env path with legacy fallback
   - Scope:
     - Update Cloudflare helper scripts to load `deployments/cloudflare/cloudflare.env` first and fall back to root `.env.cloudflare` when the new file is absent.
   - Output:
     - Backward-compatible env loading in Cloudflare helper scripts.
   - Linked requirements: FR-002, FR-003, NFR-001, NFR-002, NFR-005
   - Validation:
     - [ ] How to verify (manual steps or command):
       - Inspect `scripts/cf-*.sh`.
       - `make -n cf-up`
     - [ ] Expected result:
       - Scripts prefer the new deployment-local env path and still mention fallback compatibility for legacy root env files.
     - [ ] Logs/metrics to check (if applicable):
       - Script log messages mention the effective env-file behavior.
3. T-003 - Run validation
   - Scope:
     - Run spec lint, pre-commit, and a make dry-run for Cloudflare deployment entrypoints.
   - Output:
     - Validation evidence recorded in this spec.
   - Linked requirements: FR-002, FR-003, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command):
       - `make -n cf-up`
       - `SPEC_DIR="specs/2026-04-11-cloudflare-env-location" bash scripts/spec-lint.sh`
       - `bash scripts/precommit-run.sh`
     - [ ] Expected result:
       - Cloudflare helper entrypoints and repo checks pass after the path move.
     - [ ] Logs/metrics to check (if applicable):
       - Not applicable.

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-001, T-002, T-003
- FR-003 -> T-002, T-003
- NFR-001 -> T-002
- NFR-002 -> T-002, T-003
- NFR-005 -> T-002
- NFR-006 -> T-001, T-003

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Restore the root example path and script root-only loading if the deployment-local path causes operator confusion.

## Validation evidence

- `bash -n scripts/cf-cloudflare-migrate.sh`
- `bash -n scripts/cf-payrune-worker-deploy.sh`
- `bash -n scripts/cf-receipt-webhook-mock-worker-deploy.sh`
- `make -n cf-up`
- `SPEC_DIR="specs/2026-04-11-cloudflare-env-location" bash scripts/spec-lint.sh`
- `bash scripts/precommit-run.sh`
