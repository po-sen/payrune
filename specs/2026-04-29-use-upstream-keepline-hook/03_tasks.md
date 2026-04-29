---
doc: 03_tasks
spec_date: 2026-04-29
slug: use-upstream-keepline-hook
mode: Quick
status: DONE
owners:
  - repo-maintainers
depends_on:
  - 2026-04-29-keepline-architecture-policy
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
- Rationale: This only changes pre-commit hook metadata and does not affect runtime code or
  architecture policy rules.
- Upstream dependencies (`depends_on`):
  - `2026-04-29-keepline-architecture-policy`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`.
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: no runtime flow, data model, or package boundary design changes.
  - What would trigger switching to Full mode: changing keepline policy semantics or Go package
    structure.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): task validation below.

## Milestones

- M1: Switch to the upstream keepline project-check hook.
- M2: Validate pre-commit.

## Tasks (ordered)

1. T-001 - Update keepline provider hook configuration
   - Scope: Change the keepline provider rev and replace the local import hook with upstream
     `keepline-check`.
   - Output: Updated `.pre-commit-config.yaml`.
   - Linked requirements: FR-001 / FR-002 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): inspect `.pre-commit-config.yaml`.
     - [x] Expected result: `rev: v0.19.0` and `keepline-check` are present; local keepline hook is
           absent.
     - [x] Logs/metrics to check (if applicable): file diff.
2. T-002 - Validate updated hooks
   - Scope: Run the upstream hook and the repository pre-commit workflow.
   - Output: Passing validation.
   - Linked requirements: FR-001 / FR-002 / NFR-001 / NFR-002 / NFR-003 / NFR-004 / NFR-005
   - Validation:
     - [x] How to verify (manual steps or command): `pre-commit run keepline-check --all-files` and
           `bash scripts/precommit-run.sh`.
     - [x] Expected result: both pass.
     - [x] Logs/metrics to check (if applicable): pre-commit output.

## Validation evidence

- `SPEC_DIR="specs/2026-04-29-use-upstream-keepline-hook" bash scripts/spec-lint.sh`: passed.
- `pre-commit run keepline-check --all-files`: passed with `po-sen/keepline` v0.19.0.
- `bash scripts/precommit-run.sh`: passed.

## Traceability (optional)

- FR-001 -> T-001, T-002
- FR-002 -> T-001, T-002
- NFR-001 -> T-002
- NFR-002 -> T-002
- NFR-003 -> T-002
- NFR-004 -> T-002
- NFR-005 -> T-002
- NFR-006 -> T-001

## Rollout and rollback

- Feature flag: none.
- Migration sequencing: update spec, update pre-commit config, validate, mark spec `DONE`.
- Rollback steps: restore the local hook only if the upstream hook is removed or broken.
