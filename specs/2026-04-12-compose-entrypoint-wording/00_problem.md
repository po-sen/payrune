---
doc: 00_problem
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

# Problem & Goals

## Context

- Background:
  - The repo now uses a single `compose.yaml` with a `development` profile for dev-only services.
  - The unprofiled base stack still includes the mainnet pollers.
- Users or stakeholders:
  - Developers using `make up` and `make up-mainnet`.
  - Operators reading the local deployment instructions in `README.md`.
- Why now:
  - The current `Makefile` help text and README wording imply a cleaner separation than the actual Compose behavior provides.

## Constraints (optional)

- Technical constraints:
  - Keep the current Compose topology and service selection behavior unchanged.
  - Do not rename existing `make` targets in this cleanup.
- Timeline/cost constraints:
  - Small Quick-mode cleanup only.
- Compliance/security constraints:
  - None beyond keeping operator instructions accurate.

## Problem statement

- Current pain:
  - `make up` looks like a pure development stack entrypoint, but it still starts the unprofiled base services, including mainnet pollers.
  - `make up-mainnet` is described as a formal/mainnet-style path, but the distinction from `make up` is actually about whether the development profile is added.
- Evidence or examples:
  - `Makefile` help text currently says `up` starts the local development stack and `up-mainnet` starts the formal/mainnet-style local stack.
  - `README.md` currently describes `up-mainnet` as a formal/mainnet-style compose path instead of explicitly saying it is the base stack without development-profile services.

## Goals

- G1:
  - Make `Makefile` help text describe the current Compose behavior accurately.
- G2:
  - Update `README.md` so the local Compose entrypoints are explained in terms of base stack versus added development-profile services.

## Non-goals (out of scope)

- NG1:
  - Changing which services belong to the base stack or the `development` profile.
- NG2:
  - Renaming `make up-mainnet` or redesigning the Compose topology.

## Assumptions

- A1:
  - Preserving current behavior is preferable to reworking Compose service selection in this cleanup.
- A2:
  - Clear wording is enough to remove the main source of confusion.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - Local entrypoint descriptions match actual Compose behavior.
- Target:
  - `Makefile` help and `README.md` no longer describe `make up` as if it excluded the base mainnet pollers, and they explicitly explain that the `development` profile adds extra dev-only services on top of the base stack.
