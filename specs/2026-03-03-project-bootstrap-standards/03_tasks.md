---
doc: 03_tasks
spec_date: 2026-03-03
slug: project-bootstrap-standards
mode: Full
status: READY
owners:
  - payrune-team
depends_on: []
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
  - This initialization defines architecture boundaries, workflow policy, automation gate, and test strategy.
  - Multiple artifacts and quality gates must stay consistent across docs and code/config.
- Upstream dependencies (`depends_on`): []
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: not applicable
  - What would trigger switching to Full mode: not applicable
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): not applicable

## Milestones

- M1: Spec package drafted and linted.
- M2: AGENTS policy, Go bootstrap, and pre-commit hooks validated.

## Tasks (ordered)

1. T-001 - Build spec package for project bootstrap

   - Scope:
     - Create Full mode spec folder and all five docs with consistent headers and links.
   - Output:
     - `specs/2026-03-03-project-bootstrap-standards/*.md`
   - Linked requirements: FR-002, NFR-004
   - Validation:
     - [ ] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-03-project-bootstrap-standards" bash scripts/spec-lint.sh`
     - [ ] Expected result: lint reports all checks passed.
     - [ ] Logs/metrics to check (if applicable): N/A

2. T-002 - Create root AGENTS policy with embedded skill guidance

   - Scope:
     - Write `AGENTS.md` for project-local governance, including preserved guidance for required skills.
   - Output:
     - `AGENTS.md`
   - Linked requirements: FR-001, FR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): inspect `AGENTS.md` sections for all required workflow rules.
     - [ ] Expected result: all four skill policies are present and actionable.
     - [ ] Logs/metrics to check (if applicable): N/A

3. T-003 - Scaffold Go clean hexagonal bootstrap service

   - Scope:
     - Add module, command entrypoint, and internal layered packages for one health-check use case.
   - Output:
     - `go.mod`, `cmd/`, `internal/` code and unit tests.
   - Linked requirements: FR-003, NFR-001, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go mod tidy && go list ./... && go test ./... -short -count=1`
     - [ ] Expected result: commands succeed and tests pass.
     - [ ] Logs/metrics to check (if applicable): test output has no failures.

4. T-004 - Add pre-commit config and verification scripts
   - Scope:
     - Add provided hooks, markdownlint config, and repository scripts for spec lint and pre-commit validation.
   - Output:
     - `.pre-commit-config.yaml`, `.markdownlint.json`, `scripts/spec-lint.sh`, `scripts/precommit-run.sh`
   - Linked requirements: FR-004, NFR-003, NFR-004
   - Validation:
     - [ ] How to verify (manual steps or command): `bash scripts/precommit-run.sh`
     - [ ] Expected result: default-stage hooks pass across repository files.
     - [ ] Logs/metrics to check (if applicable): hook summary shows Passed/Skipped only.

## Traceability (optional)

- FR-001 -> T-002
- FR-002 -> T-001
- FR-003 -> T-003
- FR-004 -> T-004
- FR-005 -> T-002
- NFR-001 -> T-003
- NFR-002 -> T-003
- NFR-003 -> T-004
- NFR-004 -> T-001, T-004
- NFR-006 -> T-002, T-003

## Rollout and rollback

- Feature flag:
  - Not required for bootstrap artifacts.
- Migration sequencing:
  - Apply specs and policy files first, then code/config, then run validation.
- Rollback steps:
  - Revert created files if baseline causes blocker; re-apply incrementally by task.

## Ready-to-code checklist

- [x] Spec folder includes all Full mode documents.
- [x] Frontmatter values are consistent across all spec docs.
- [x] `owners` is set and `depends_on` is valid.
- [x] Mode decision and rationale are documented.
- [x] Requirement, task, and test traceability IDs are present.
- [x] `SPEC_DIR=\"specs/2026-03-03-project-bootstrap-standards\" bash scripts/spec-lint.sh` passes.
- [x] Spec status is `READY` across all documents.
