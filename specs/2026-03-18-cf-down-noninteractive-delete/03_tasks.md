---
doc: 03_tasks
spec_date: 2026-03-18
slug: cf-down-noninteractive-delete
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
  - This is a small script-only operational change with no schema, integration-shape, or runtime
    behavior redesign.
- Upstream dependencies (`depends_on`):
  - []
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - The change is confined to existing shell scripts that already define the teardown flow.
  - What would trigger switching to Full mode:
    - Any change to Cloudflare resource design, runtime bindings, or new teardown orchestration.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Not skipped.

## Milestones

- M1:
  - Document the non-interactive teardown requirement in a quick spec.
- M2:
  - Patch the delete scripts and verify `cf-down` still resolves to the same delete sequence.

## Tasks (ordered)

1. T-001 - Patch worker delete scripts for non-interactive deletion

   - Scope:
     - Update every delete script reached by `make cf-down` to pass Wrangler's supported
       non-interactive delete flag.
   - Output:
     - Updated `scripts/cf-api-worker-delete.sh`
     - Updated `scripts/cf-poller-worker-delete.sh`
     - Updated `scripts/cf-receipt-webhook-mock-worker-delete.sh`
     - Updated `scripts/cf-webhook-dispatcher-worker-delete.sh`
   - Linked requirements: FR-001, NFR-001, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): inspect the four delete scripts for
           `wrangler delete --force`.
     - [ ] Expected result: each script invokes `wrangler delete --force` while keeping its existing
           positional arguments or env flags.
     - [ ] Logs/metrics to check (if applicable): none

2. T-002 - Verify cf-down entrypoint remains stable
   - Scope:
     - Confirm `make cf-down` still calls the same delete scripts in the same order after the patch.
   - Output:
     - Validation evidence from `make -n cf-down` and spec lint.
   - Linked requirements: FR-002, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): run `make -n cf-down` and `SPEC_DIR="specs/2026-03-18-cf-down-noninteractive-delete" bash scripts/spec-lint.sh`.
     - [ ] Expected result: `make -n cf-down` shows the same five delete commands in order, and
           spec-lint passes.
     - [ ] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-002
- NFR-001 -> T-001
- NFR-002 -> T-001, T-002

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Remove `--force` from the four delete scripts to restore interactive Wrangler confirmation.
