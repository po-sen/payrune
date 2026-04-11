---
doc: 03_tasks
spec_date: 2026-04-12
slug: compose-entrypoint-wording
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
  - This is a wording/documentation cleanup around existing Compose behavior. It changes no runtime topology, schema, or external integration behavior.
- Upstream dependencies (`depends_on`):
  - `2026-04-09-compose-env-example`
- Dependency gate before `READY`:
  - The dependency must remain folder-wide `DONE`.
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - There is no design change, only help text and README alignment.
  - What would trigger switching to Full mode:
    - If the work expanded into changing Compose topology or service/profile ownership.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Validation is listed in the task below.

## Milestones

- M1:
  - Align compose entrypoint wording with actual behavior.

## Tasks (ordered)

1. T-001 - Align compose entrypoint wording
   - Scope:
     - Update `Makefile` help text so `up`/`down`/`config` are described as base stack plus development-profile services.
     - Update `Makefile` help text so `up-mainnet`/`down-mainnet`/`config-mainnet` are described as base stack only.
     - Update `README.md` to explain the same distinction clearly and remove “formal/mainnet-style” wording that overstates the separation.
   - Output:
     - Honest `Makefile` help text and matching README instructions.
   - Linked requirements: FR-001, FR-002, NFR-001, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command):
       - Inspect `Makefile` and `README.md`.
       - `make help`
       - `docker compose --env-file deployments/compose/compose.dev.env --profile development -f deployments/compose/compose.yaml config --services`
     - [ ] Expected result:
       - The docs/help describe `make up` as base stack plus development-profile services and `make up-mainnet` as base stack only, matching the rendered service list.
     - [ ] Logs/metrics to check (if applicable):
       - Not applicable.

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-001
- NFR-001 -> T-001
- NFR-002 -> T-001
- NFR-006 -> T-001

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Restore the previous wording if the new phrasing proves less clear for operators.

## Validation evidence

- `make help`
- `docker compose --env-file deployments/compose/compose.dev.env --profile development -f deployments/compose/compose.yaml config --services`
- `SPEC_DIR="specs/2026-04-12-compose-entrypoint-wording" bash scripts/spec-lint.sh`
