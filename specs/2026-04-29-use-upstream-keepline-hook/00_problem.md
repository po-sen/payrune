---
doc: 00_problem
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

# Problem & Goals

## Context

- Background: The repository temporarily used a local pre-commit hook to run
  `keepline import-check` because `po-sen/keepline` v0.17.0 did not publish an upstream
  pre-commit hook for whole-project checks.
- Users or stakeholders: repository maintainers and contributors.
- Why now: The upstream keepline repository now publishes `keepline-check` in v0.19.0, so this
  repository can remove the local hook and use the provider hook directly.

## Constraints (optional)

- Technical constraints: Preserve commit-message and import-policy enforcement; keep
  `.pre-commit-config.yaml` pinned to a tagged keepline release.
- Timeline/cost constraints: Limit the change to pre-commit configuration and this spec.
- Compliance/security constraints: Do not weaken denied-file or Conventional Commit checks.

## Problem statement

- Current pain: The local `keepline-import-check` hook duplicates metadata that now belongs in the
  upstream keepline hook repository.
- Evidence or examples: `po-sen/keepline` v0.19.0 includes `keepline-check` with entry
  `keepline check --scope staged`.

## Goals

- G1: Update `.pre-commit-config.yaml` to use `po-sen/keepline` v0.19.0.
- G2: Replace the local keepline import hook with the upstream `keepline-check` hook.
- G3: Validate the updated pre-commit workflow.

## Non-goals (out of scope)

- NG1: Change `keepline.toml` architecture policy.
- NG2: Change Go application code.

## Assumptions

- A1: `keepline-check` is the upstream hook intended to run project checks, including import policy.
- A2: Keeping `keepline-commit-msg` separate remains necessary for commit-message hook stage.

## Open questions

- None.

## Success metrics

- Metric: `pre-commit run keepline-check --all-files`
- Target: passes with the upstream hook from `po-sen/keepline` v0.19.0.
- Metric: `bash scripts/precommit-run.sh`
- Target: passes.
