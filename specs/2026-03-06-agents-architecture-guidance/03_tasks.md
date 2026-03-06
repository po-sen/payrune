---
doc: 03_tasks
spec_date: 2026-03-06
slug: agents-architecture-guidance
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
  - This change augments repository guidance and architecture expectations in documentation only, without new runtime behavior, schema, or external integration.

## Tasks (ordered)

1. T-001 - Capture the architecture guidance gaps

- Scope:
  - Identify the ambiguities in the current `AGENTS.md` that caused repeated corrections during recent feature work.
- Linked requirements: FR-001, FR-002, FR-003, FR-004, NFR-001, NFR-002
- Validation:
  - [x] The replacement document addresses layer boundaries, naming, modeling, and review triggers directly.

1. T-002 - Rewrite `AGENTS.md` as an agent operating manual

- Scope:
  - Preserve the existing important sections, then refine the repo-specific override into clearer best-practice guidance for the coding agent without changing the meaning of correct instructions.
- Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, NFR-001, NFR-002, NFR-003
- Validation:
  - [x] `AGENTS.md` clearly states workflow rules, layer responsibilities, naming rules, and anti-patterns.
  - [x] `AGENTS.md` preserves the existing important sections and adds repo-specific override guidance.
  - [x] `AGENTS.md` refines the repo-specific section into better best-practice wording without changing the meaning of correct existing guidance.

1. T-003 - Validate the new guidance and sync the spec

- Scope:
  - Run spec lint and pre-commit checks; update spec docs to final state after the rewrite.
- Linked requirements: FR-001, FR-002, FR-003, FR-004, NFR-001, NFR-002, NFR-003
- Validation:
  - [x] `SPEC_DIR="specs/2026-03-06-agents-architecture-guidance" bash scripts/spec-lint.sh`
  - [x] `bash scripts/precommit-run.sh`

## Traceability

- FR-001 -> T-001, T-002, T-003
- FR-002 -> T-001, T-002, T-003
- FR-003 -> T-001, T-002, T-003
- FR-004 -> T-001, T-002, T-003
- FR-005 -> T-002, T-003
- FR-006 -> T-002, T-003
- NFR-001 -> T-001, T-002, T-003
- NFR-002 -> T-001, T-002, T-003
- NFR-003 -> T-002, T-003
