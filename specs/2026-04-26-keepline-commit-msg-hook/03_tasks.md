---
doc: 03_tasks
spec_date: 2026-04-26
slug: keepline-commit-msg-hook
mode: Quick
status: DONE
owners:
  - codex
depends_on: []
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null # set to 02_design.md in Full mode
  tasks: 03_tasks.md
  test_plan: null # set to 04_test_plan.md if produced
---

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale: This is a small repository tooling configuration change with no new runtime integration, database schema, async flow, or architecture impact.
- Upstream dependencies (`depends_on`): []
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: The requested behavior is fully represented by one pre-commit YAML addition plus command validation.
  - What would trigger switching to Full mode: Switch to Full only if Keepline requires a repository wrapper script, persistent configuration file, or broader release workflow changes.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): Each task includes explicit spec-lint, YAML, and hook execution checks.

## Milestones

- M1: Add the Keepline commit message hook to pre-commit configuration.
- M2: Validate the spec and configured hook execution.

## Tasks (ordered)

1. T-001 - Add Keepline commit message hook
   - Scope: Update `.pre-commit-config.yaml` with `default_install_hook_types` and the requested local `keepline-commit-msg` hook.
   - Output: Pre-commit configuration contains a `commit-msg` hook that invokes Keepline v0.1.0 through `go run`.
   - Linked requirements: FR-001 / FR-002 / NFR-001 / NFR-002 / NFR-003 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `pre-commit run keepline-commit-msg --hook-stage commit-msg --commit-msg-filename /tmp/keepline-commit-msg-test.txt`
     - [x] Expected result: The hook runs Keepline against the sample commit message file.
     - [x] Logs/metrics to check (if applicable): Pre-commit output shows the `keepline commit message` hook result.
2. T-002 - Validate spec and hook behavior
   - Scope: Run repository spec lint and command-level validation for the new hook.
   - Output: Validation evidence records whether Keepline v0.1.0 is usable from this repo.
   - Linked requirements: FR-003 / NFR-004 / NFR-005
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-04-26-keepline-commit-msg-hook" bash scripts/spec-lint.sh`
     - [x] Expected result: Spec lint passes for this spec package.
     - [x] Logs/metrics to check (if applicable): Keepline or Go output identifies module download/execution problems separately from policy failures.

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-001
- FR-003 -> T-002
- NFR-001 -> T-001
- NFR-002 -> T-001
- NFR-003 -> T-001
- NFR-004 -> T-002
- NFR-005 -> T-002
- NFR-006 -> T-001

## Rollout and rollback

- Feature flag: None
- Migration sequencing: Update spec first, add pre-commit YAML, run hook validation, then mark spec done if validation completes.
- Rollback steps: Remove the top-level `default_install_hook_types` block and the `keepline-commit-msg` local hook if it blocks commits unexpectedly.

## Validation evidence

- `SPEC_DIR="specs/2026-04-26-keepline-commit-msg-hook" bash scripts/spec-lint.sh` passed before implementation.
- `go run github.com/po-sen/keepline/cmd/keepline@v0.1.0 commit-msg /tmp/keepline-commit-msg-test.txt` passed after downloading `github.com/po-sen/keepline v0.1.0`.
- `pre-commit run keepline-commit-msg --hook-stage commit-msg --commit-msg-filename /tmp/keepline-commit-msg-test.txt` passed with output `keepline commit message ... Passed`.
