---
doc: 00_problem
spec_date: 2026-04-11
slug: cloudflare-env-location
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
  - Compose env files already live under `deployments/compose`, but the Cloudflare env template still lives at repo root as `.env.cloudflare.example`.
- Users or stakeholders:
  - Operators deploying Cloudflare workers via `make cf-up`.
  - Developers looking for deployment-local configuration near `wrangler.toml`.
- Why now:
  - The current location is inconsistent with the repo's deployment-file organization and makes the root directory noisier than necessary.

## Constraints (optional)

- Technical constraints:
  - Keep `make cf-up` and the Cloudflare scripts working during the transition.
  - Do not move business logic or Cloudflare runtime code.
- Timeline/cost constraints:
  - Small repo-local cleanup; keep the change in Quick mode.
- Compliance/security constraints:
  - Preserve the current behavior that local shell env can override env-file values.

## Problem statement

- Current pain:
  - The Cloudflare env template is the only checked-in deployment env example still living at repo root.
  - Cloudflare deploy scripts hardcode the root `.env.cloudflare` path, which makes the current location part of the runtime contract.
- Evidence or examples:
  - Root README tells operators to copy `.env.cloudflare.example` to `.env.cloudflare`.
  - Cloudflare deploy and migrate scripts load `"$ROOT_DIR/.env.cloudflare"` directly.

## Goals

- G1:
  - Move the checked-in Cloudflare env template under `deployments/cloudflare/`.
- G2:
  - Make Cloudflare scripts and docs prefer the new deployment-local env path.
- G3:
  - Avoid breaking existing local setups immediately by keeping a compatibility path for the old root env file.

## Non-goals (out of scope)

- NG1:
  - Reworking Cloudflare secret sync behavior or Wrangler config.
- NG2:
  - Changing Compose env file locations.

## Assumptions

- A1:
  - `deployments/cloudflare/cloudflare.env.example` and `deployments/cloudflare/cloudflare.env` are clearer names than root dotfiles for deployment-local config.
- A2:
  - Temporary fallback support for root `.env.cloudflare` is preferable to a hard breaking cutover.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - Cloudflare env example is no longer stored at repo root.
- Target:

  - The checked-in example file lives under `deployments/cloudflare/`.

- Metric:
  - Existing operator workflows keep working during the transition.
- Target:
  - Scripts accept the new deployment-local env file and still support the old root file as fallback.
