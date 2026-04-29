---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Term: `keepline-check`
- Definition: Upstream pre-commit hook from `po-sen/keepline` that runs
  `keepline check --scope staged`.

## Out-of-scope behaviors

- OOS1: Modify keepline import policy rules.
- OOS2: Refactor application code.

## Functional requirements

### FR-001 - Use upstream keepline check hook

- Description: `.pre-commit-config.yaml` must use `po-sen/keepline` v0.19.0 and configure the
  upstream `keepline-check` hook.
- Acceptance criteria:
  - [x] The keepline provider `rev` is updated to `v0.19.0`.
  - [x] `keepline-check` is listed under the upstream keepline provider hooks.
  - [x] The local `repo: local` keepline import hook is removed.
- Notes: v0.19.0 is the tagged release that includes the upstream hook metadata.

### FR-002 - Preserve commit and file/import enforcement

- Description: The updated hook list must still validate commit messages and project policies.
- Acceptance criteria:
  - [x] `keepline-commit-msg` remains configured.
  - [x] `keepline-check` runs in the pre-commit stage.
  - [x] `bash scripts/precommit-run.sh` passes.
- Notes: `keepline-check` replaces the separate local import-check hook.

## Non-functional requirements

- Performance (NFR-001): No more than one upstream keepline pre-commit project-check hook runs.
- Availability/Reliability (NFR-002): Pre-commit validation passes after the hook change.
- Security/Privacy (NFR-003): Existing denied-file and secret checks remain enabled.
- Compliance (NFR-004): Conventional Commit validation remains enabled.
- Observability (NFR-005): Validation evidence records the hook command outcomes.
- Maintainability (NFR-006): Keepline hook metadata is owned by the upstream provider, not duplicated
  locally.

## Dependencies and integrations

- External systems: `https://github.com/po-sen/keepline` v0.19.0 and pre-commit.
- Internal services: none.
