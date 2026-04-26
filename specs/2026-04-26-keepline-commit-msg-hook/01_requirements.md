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

- Keepline: The Go-based pre-commit hook repository at `https://github.com/po-sen/keepline`, pinned to `rev: v0.1.0`.
- Commit message hook: A pre-commit hook that receives the commit message file at the `commit-msg` Git hook stage.

## Out-of-scope behaviors

- OOS1: Repository code behavior, runtime services, and application architecture are unchanged.
- OOS2: Commit message policy design remains owned by Keepline and is not reimplemented here.

## Functional requirements

### FR-001 - Default install hook types

- Description: `.pre-commit-config.yaml` must declare both regular pre-commit checks and commit message checks as default install hook types.
- Acceptance criteria:
  - [x] `default_install_hook_types` exists at the top level.
  - [x] `pre-commit` is included in `default_install_hook_types`.
  - [x] `commit-msg` is included in `default_install_hook_types`.
- Notes: Existing hook repositories remain on their normal stages; this only changes what `pre-commit install` installs by default.

### FR-002 - Keepline versioned hook repository

- Description: `.pre-commit-config.yaml` must include Keepline as a versioned pre-commit hook repository.
- Acceptance criteria:
  - [x] The repo is `https://github.com/po-sen/keepline`.
  - [x] The repo is pinned with `rev: v0.1.0`.
  - [x] The configured hook id is `keepline-commit-msg`.
  - [x] The configuration does not duplicate Keepline's hook manifest details as a local `go run` hook.
- Notes: Keepline `v0.1.0` provides `entry: keepline commit-msg`, `language: golang`, `stages: [commit-msg]`, and `pass_filenames: true` through its hook manifest.

### FR-003 - Hook execution verification

- Description: The configured hook must be tested locally against a sample commit message file.
- Acceptance criteria:
  - [x] A pre-commit-driven command proves Keepline `v0.1.0` can be installed and executed through the remote hook repository.
  - [x] Validation output distinguishes a tool execution problem from a commit message policy failure.
- Notes: Network access may be required for the first pre-commit hook environment installation if the repository is not already cached.

## Non-functional requirements

- Performance (NFR-001): The commit message hook should only run during `commit-msg` stage and not add runtime cost to normal Go tests or service startup.
- Availability/Reliability (NFR-002): Existing default pre-commit hooks must remain syntactically valid after adding the commit message hook.
- Security/Privacy (NFR-003): The change must not disable existing secret detection or large-file checks.
- Compliance (NFR-004): The spec folder must pass `SPEC_DIR="specs/2026-04-26-keepline-commit-msg-hook" bash scripts/spec-lint.sh`.
- Observability (NFR-005): Validation commands must produce enough output to tell whether Keepline ran.
- Maintainability (NFR-006): The hook configuration should use Keepline's published pre-commit hook metadata rather than duplicating its entry command locally.

## Dependencies and integrations

- External systems: GitHub/pre-commit hook fetch for `https://github.com/po-sen/keepline` at `v0.1.0`, plus Go toolchain installation for Keepline's `language: golang` hook when not already cached.
- Internal services: `.pre-commit-config.yaml`, `scripts/spec-lint.sh`.
