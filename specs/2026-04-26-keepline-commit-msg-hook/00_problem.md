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

- Background: The repository already uses pre-commit for formatting, linting, tests, and spec validation, and now needs commit message validation through Keepline's own pre-commit hook repository metadata without losing the default installation path for the existing pre-commit-stage hooks.
- Users or stakeholders: Maintainers who want local commits to be checked by Keepline `v0.1.0`.
- Why now: The user wants the pre-commit configuration to use the standard `repo` plus `rev: v0.1.0` form and wants `pre-commit install` to install both the regular pre-commit hook and the Keepline commit message hook.

## Constraints (optional)

- Technical constraints: Keep helper automation under `scripts/` if new scripts are needed; this change should only update pre-commit configuration unless validation shows a repository helper is required.
- Timeline/cost constraints: Quick mode, scoped to one local pre-commit hook addition and command-level validation.
- Compliance/security constraints: The hook must not weaken existing pre-commit hooks or bypass existing secret detection.

## Problem statement

- Current pain: The existing Keepline hook works, but it is configured as a local `go run` hook rather than as a versioned pre-commit hook repository, and the default install hook types currently only cover `commit-msg`.
- Evidence or examples:
  - `.pre-commit-config.yaml` uses `repo: local` with `entry: go run github.com/po-sen/keepline/cmd/keepline@v0.1.0 commit-msg`.

## Goals

- G1: Configure `default_install_hook_types` in YAML block form so `pre-commit install` installs both `pre-commit` and `commit-msg` Git hooks by default.
- G2: Configure Keepline as a remote pre-commit hook repository pinned with `rev: v0.1.0`.
- G3: Validate that the configured hook can run against a sample commit message file.

## Non-goals (out of scope)

- NG1: Do not change Keepline's validation rules in this repository.
- NG2: Do not change existing non-commit-message pre-commit hooks except as needed to keep YAML valid.

## Assumptions

- A1: Keepline `v0.1.0` exposes a `.pre-commit-hooks.yaml` manifest with hook id `keepline-commit-msg`.
- A2: A sample Conventional Commit message is sufficient to verify the hook is executable from pre-commit.

## Open questions

- Q1: None.

## Success metrics

- Metric: Local hook availability and execution.
- Target: `pre-commit run keepline-commit-msg --hook-stage commit-msg --commit-msg-filename <sample file>` completes successfully using `repo: https://github.com/po-sen/keepline` and `rev: v0.1.0`, or reports a Keepline validation result that proves the CLI was invoked.
