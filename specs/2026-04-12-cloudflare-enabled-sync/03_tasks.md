---
doc: 03_tasks
spec_date: 2026-04-12
slug: cloudflare-enabled-sync
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-11-cloudflare-env-location
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
  - This is a small Cloudflare deploy-contract bug fix with matching docs updates. It does not add a new integration or schema.
- Upstream dependencies (`depends_on`):
  - `2026-04-11-cloudflare-env-location`
- Dependency gate before `READY`:
  - The dependency is already folder-wide `DONE`.
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - The fix is limited to env sync wiring and docs alignment.
  - What would trigger switching to Full mode:
    - Replacing Wrangler secret sync with a different runtime env delivery mechanism.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Validation is listed in the task below.

## Milestones

- M1:
  - Align Cloudflare worker env sync with the advertised operator intent flags.

## Tasks (ordered)

1. T-001 - Sync Cloudflare `*_ENABLED` flags and align docs
   - Scope:
     - Update `scripts/cf-payrune-worker-deploy.sh` so all Bitcoin and Ethereum `*_ENABLED` keys exposed in `deployments/cloudflare/cloudflare.env.example` are passed as plain Wrangler deploy vars instead of secrets.
     - Update Cloudflare README text so the supported optional env-backed worker values match the script behavior and clearly distinguish non-secret flags from secret-backed values.
   - Output:
     - Working Cloudflare env sync for policy enablement flags as non-secret deploy vars, plus aligned docs.
   - Linked requirements: FR-001, FR-002, NFR-001, NFR-002, NFR-003, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command):
       - Inspect `deployments/cloudflare/cloudflare.env.example`, `deployments/cloudflare/payrune/README.md`, and `scripts/cf-payrune-worker-deploy.sh`.
       - `bash -n scripts/cf-payrune-worker-deploy.sh`
       - `cd deployments/cloudflare/payrune && npm exec -- wrangler deploy --dry-run --var TEST_FLAG:true`
       - `make -n cf-up`
     - [ ] Expected result:
       - Every advertised `*_ENABLED` key is included in the non-secret deploy-var path, none of them stay in the secret-sync list, and Cloudflare docs no longer overpromise values that `make cf-up` ignores or misclassify as secrets.
     - [ ] Logs/metrics to check (if applicable):
       - Not applicable.

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-001
- NFR-001 -> T-001
- NFR-002 -> T-001
- NFR-003 -> T-001
- NFR-006 -> T-001

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Remove the `*_ENABLED` keys from the sync list and revert the docs if the worker should intentionally keep these values unsynced.

## Validation evidence

- `bash -n scripts/cf-payrune-worker-deploy.sh`
- `cd deployments/cloudflare/payrune && npm exec -- wrangler deploy --dry-run --var TEST_FLAG:true`
- `SPEC_DIR="specs/2026-04-12-cloudflare-enabled-sync" bash scripts/spec-lint.sh`
- `bash scripts/precommit-run.sh`
