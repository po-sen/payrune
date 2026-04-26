---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Keepline: The Go CLI at `github.com/po-sen/keepline/cmd/keepline@v0.1.0` used to validate commit messages.
- Commit message hook: A pre-commit hook that receives the commit message file at the `commit-msg` Git hook stage.

## Out-of-scope behaviors

- OOS1: Repository code behavior, runtime services, and application architecture are unchanged.
- OOS2: Commit message policy design remains owned by Keepline and is not reimplemented here.

## Functional requirements

### FR-001 - Commit message hook installation type

- Description: `.pre-commit-config.yaml` must declare `commit-msg` as a default install hook type.
- Acceptance criteria:
  - [x] `default_install_hook_types` exists at the top level.
  - [x] `commit-msg` is included in `default_install_hook_types`.
- Notes: Existing hook repositories and regular pre-commit stage behavior should remain valid.

### FR-002 - Keepline local hook

- Description: `.pre-commit-config.yaml` must include a local hook named `keepline-commit-msg`.
- Acceptance criteria:
  - [x] The hook entry is `go run github.com/po-sen/keepline/cmd/keepline@v0.1.0 commit-msg`.
  - [x] The hook uses `language: system`.
  - [x] The hook runs at `stages: [commit-msg]`.
  - [x] The hook passes filenames to Keepline.
- Notes: The hook should live in the existing pre-commit configuration rather than a separate helper script unless validation requires one.

### FR-003 - Hook execution verification

- Description: The configured hook must be tested locally against a sample commit message file.
- Acceptance criteria:
  - [x] A direct or pre-commit-driven command proves `github.com/po-sen/keepline/cmd/keepline@v0.1.0` can be executed.
  - [x] Validation output distinguishes a tool execution problem from a commit message policy failure.
- Notes: Network access may be required for the first `go run` download if the module is not already cached.

## Non-functional requirements

- Performance (NFR-001): The commit message hook should only run during `commit-msg` stage and not add runtime cost to normal Go tests or service startup.
- Availability/Reliability (NFR-002): Existing default pre-commit hooks must remain syntactically valid after adding the commit message hook.
- Security/Privacy (NFR-003): The change must not disable existing secret detection or large-file checks.
- Compliance (NFR-004): The spec folder must pass `SPEC_DIR="specs/2026-04-26-keepline-commit-msg-hook" bash scripts/spec-lint.sh`.
- Observability (NFR-005): Validation commands must produce enough output to tell whether Keepline ran.
- Maintainability (NFR-006): The hook configuration should be local, explicit, and limited to the requested Keepline command.

## Dependencies and integrations

- External systems: Go module proxy or GitHub module fetch for `github.com/po-sen/keepline/cmd/keepline@v0.1.0` when not already cached.
- Internal services: `.pre-commit-config.yaml`, `scripts/spec-lint.sh`.
