---
doc: 00_problem
spec_date: 2026-03-18
slug: cf-down-noninteractive-delete
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

# Problem & Goals

## Context

- Background:
  - `make cf-down` tears down the deployed Cloudflare Workers by calling the worker delete scripts
    under `scripts/`.
- Users or stakeholders:
  - payrune maintainers operating Cloudflare teardown flows.
- Why now:
  - The current delete flow prompts for confirmation during `wrangler delete`, adding repeated
    manual input to a command that is already explicitly destructive.

## Constraints (optional)

- Technical constraints:
  - Keep the change limited to the existing delete scripts used by `make cf-down`.
  - Preserve the existing worker deletion order and per-script arguments.
- Compliance/security constraints:
  - Do not broaden the change into deploy or migrate flows.

## Problem statement

- Current pain:
  - `make cf-down` requires an extra interactive confirmation for each worker delete step.
- Evidence or examples:
  - The delete scripts currently call `npm exec -- wrangler delete ...` without a non-interactive
    flag.

## Goals

- G1:
  - Make `make cf-down` run without manual confirmation prompts.
- G2:
  - Keep the teardown behavior explicit and scoped to the existing Cloudflare worker delete flows.

## Non-goals (out of scope)

- NG1:
  - Changing `make cf-up`, migration, or worker runtime behavior.
- NG2:
  - Reordering teardown steps or redesigning Cloudflare resource ownership.

## Assumptions

- A1:
  - `wrangler delete --force` is the supported way to bypass the interactive delete confirmation.
- A2:
  - `make cf-down` is an intentionally destructive operator command, so requiring a second prompt is
    unnecessary.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - All delete scripts used by `make cf-down` invoke Wrangler in non-interactive mode.
- Target:
  - Each `scripts/cf-*-worker-delete.sh` used by `make cf-down` includes `--force` on
    `wrangler delete`.
- Metric:
  - Teardown entrypoint remains unchanged for operators.
- Target:
  - `make -n cf-down` still expands to the same five delete scripts in the same order.
