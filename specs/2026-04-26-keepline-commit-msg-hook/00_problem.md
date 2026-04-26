---
doc: 00_problem
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

# Problem & Goals

## Context

- Background: The repository already uses pre-commit for formatting, linting, tests, and spec validation, but it does not yet install or run a `commit-msg` hook for commit message validation.
- Users or stakeholders: Maintainers who want local commits to be checked by `github.com/po-sen/keepline/cmd/keepline@v0.1.0`.
- Why now: The user wants to verify the newly released Keepline CLI can be used from this repository through pre-commit's commit message stage.

## Constraints (optional)

- Technical constraints: Keep helper automation under `scripts/` if new scripts are needed; this change should only update pre-commit configuration unless validation shows a repository helper is required.
- Timeline/cost constraints: Quick mode, scoped to one local pre-commit hook addition and command-level validation.
- Compliance/security constraints: The hook must not weaken existing pre-commit hooks or bypass existing secret detection.

## Problem statement

- Current pain: Commit message validation is not wired into pre-commit, so `pre-commit install --hook-type commit-msg` would not install the requested Keepline commit message hook.
- Evidence or examples:
  - `.pre-commit-config.yaml` has default hook repositories and local hooks, but no `default_install_hook_types` or `commit-msg` stage hook.

## Goals

- G1: Add `default_install_hook_types: [commit-msg]` behavior in YAML block form so pre-commit installs the commit message hook by default.
- G2: Add a local `keepline-commit-msg` hook that runs `go run github.com/po-sen/keepline/cmd/keepline@v0.1.0 commit-msg`.
- G3: Validate that the configured hook can run against a sample commit message file.

## Non-goals (out of scope)

- NG1: Do not change Keepline's validation rules in this repository.
- NG2: Do not change existing non-commit-message pre-commit hooks except as needed to keep YAML valid.

## Assumptions

- A1: The requested `go run ...@v0.1.0` command is the intended installation path for testing the release.
- A2: A sample Conventional Commit message is sufficient to verify the hook is executable from pre-commit.

## Open questions

- Q1: None.

## Success metrics

- Metric: Local hook availability and execution.
- Target: `pre-commit run keepline-commit-msg --hook-stage commit-msg --commit-msg-filename <sample file>` completes successfully, or reports a Keepline validation result that proves the CLI was invoked.
