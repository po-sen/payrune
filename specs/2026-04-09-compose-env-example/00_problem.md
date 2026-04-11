---
doc: 00_problem
spec_date: 2026-04-09
slug: compose-env-example
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-07-ethereum-contract-readiness
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
  - The repo already ships [`compose.yaml`](/Users/posen/Desktop/payrune/deployments/compose/compose.yaml) for local services and a checked-in fake-value env file for the local development path.
  - [`deployments/compose/compose.env.example`](/Users/posen/Desktop/payrune/deployments/compose/compose.env.example) is the main checked-in compose env example and can act as a readable superset for both the formal/mainnet-style path and the local development path.
  - The current [`Makefile`](/Users/posen/Desktop/payrune/Makefile) still auto-selects between formal/mainnet-style and local development behavior instead of exposing explicit local targets.
  - The current compose setup still spreads local development topology across a second compose file instead of keeping one obvious compose entrypoint.
- Users or stakeholders:
  - Developers and operators starting either the formal/mainnet-style local stack or the local development path.
  - Reviewers trying to confirm which envs are relevant to the formal path versus the development path.
- Why now:
  - The env contract recently became more explicit with per-policy `*_ENABLED` flags and grouped network-specific settings.
  - The user explicitly wants a `deployments/compose/compose.env.example` checked in.

## Constraints (optional)

- Technical constraints:
  - Keep this as a Quick-mode operational/docs change; no runtime behavior changes.
  - Keep one checked-in env example file readable even if it covers both the formal/mainnet-style path and the local development path.
  - Keep grouping aligned with the current env layout convention: same network/kind together, with paired Bitcoin `ENABLED + XPUB`.
- Timeline/cost constraints:
  - Prefer a small, explicit example file over a broader env documentation rewrite.
- Compliance/security constraints:
  - Do not commit real secrets.

## Problem statement

- Current pain:
  - Operators currently jump between the main example file and the checked-in development env when trying to understand the full local compose contract.
  - If `compose.env.example` is going to stay the main checked-in reference, it still needs the development-chain blocks to be grouped clearly and kept disabled by default.
  - The checked-in `compose.dev.env` still repeats several values that already match `compose.yaml` defaults, which makes the development env file noisier than it needs to be.
  - `make up`/`make down` should be obvious local development commands, while formal/mainnet-style usage should be explicit instead of hidden behind env-file presence checks.
  - Cloudflare migration should not need its own Make target, but `make cf-up` still needs to run migration before deploying workers.
  - The current `Makefile` compose targets have started to accumulate command-assembly variables that make a small local workflow look more abstract than it is.
  - The local operator entrypoints are still spread across README prose; there is no simple `make help` summary.
  - The remaining `mainnet` profile is no longer buying much; it makes the formal/mainnet-style path look more special than it really is, while the real special-case services are the local development helpers.
- Evidence or examples:
  - [`deployments/compose/compose.yaml`](/Users/posen/Desktop/payrune/deployments/compose/compose.yaml)
  - [`deployments/compose/compose.dev.env`](/Users/posen/Desktop/payrune/deployments/compose/compose.dev.env)

## Goals

- G1:
  - Keep `deployments/compose/compose.env.example` as the main checked-in example file for compose envs.
- G2:
  - Expand that example so it also includes local development-chain env blocks for Bitcoin testnet4 and Ethereum Sepolia.
- G3:
  - Make local compose targets explicit: `up/down/config` for the local development path, plus `up-mainnet/down-mainnet/config-mainnet` for the formal/mainnet-style path.
- G4:
  - Collapse local development topology into a single `compose.yaml`, with one dedicated development profile for dev-only services.
- G5:
  - Keep Cloudflare migration inside `make cf-up`, while still removing the standalone `cf-migrate` Make target.
- G6:
  - Make the formal/mainnet-style example safe by default.
- G7:
  - Keep the `Makefile` entrypoints discoverable through a simple built-in help target.
- G8:
  - Keep `compose.dev.env` focused on development-only overrides instead of restating compose defaults.

## Non-goals (out of scope)

- NG1:
  - Changing the underlying compose env contract or env names.
- NG2:
  - Removing the committed fake-value development env file; that file can still exist as the default local development env source.
- NG3:
  - Changing the actual default values in `compose.yaml`.

## Assumptions

- A1:
  - The example file can be a superset env file: extra development-path vars are acceptable when their policy flags stay disabled by default in the formal/mainnet-style example.
- A2:
  - Example values may include public non-secret references such as the canonical mainnet USDT asset reference.
- A3:
  - `deployments/compose/compose.env` is intended to be an untracked operator-owned local env file derived from `compose.env.example`.

## Open questions

- None for this scoped change.

## Success metrics

- Metric:
  - A checked-in env example exists for both the formal/mainnet-style compose path and the local development compose path, and a separate checked-in fake-value env file still exists for the development path.
- Target:
  - `docker compose --env-file deployments/compose/compose.env.example -f deployments/compose/compose.yaml config` succeeds.
  - `docker compose --env-file deployments/compose/compose.env.example --profile development -f deployments/compose/compose.yaml config` succeeds.
  - `docker compose --env-file deployments/compose/compose.dev.env --profile development -f deployments/compose/compose.yaml config` succeeds.
  - `deployments/compose/compose.dev.env` contains only the development-specific overrides needed to enable the local development path or satisfy required development-only settings.
  - `make up/down/config` has one obvious behavior: use the checked-in local development env with the `development` profile, while `make up-mainnet/down-mainnet/config-mainnet` uses `deployments/compose/compose.env` without a `mainnet` profile.
