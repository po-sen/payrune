---
doc: 00_problem
spec_date: 2026-03-19
slug: cloudflare-worker-consolidation
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-11-cloudflare-poller-workers
  - 2026-03-12-api-worker-naming-unification
  - 2026-03-13-cloudflare-webhook-dispatcher-worker
  - 2026-03-14-runtime-entrypoint-alignment
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background:
  - `make cf-up` currently deploys separate Cloudflare Workers for the public API, the Bitcoin
    poller worker environments, the webhook dispatcher, and the receipt webhook mock.
  - The Go application logic already treats API handling, polling, and webhook dispatch as one
    product area, but the Cloudflare deployment shell is fragmented across multiple worker
    directories, scripts, and Wasm entrypoints.
- Users or stakeholders:
  - payrune maintainers operating the Cloudflare deployment.
  - downstream teams that already manage additional app/API/www workers and want the payrune worker
    inventory to stay compact and service-oriented.
- Why now:
  - The desired operating model is one Cloudflare worker per backend service where practical.
  - Maintaining three separate payrune worker deployments adds control-plane clutter without
    matching the intended deployment or fault-domain boundaries.

## Constraints (optional)

- Technical constraints:
  - Keep Clean Architecture boundaries in Go intact; the change is a deployment/runtime
    consolidation, not a use-case rewrite.
  - Preserve existing public API routes, scheduled use-case behavior, Cloudflare PostgreSQL bridge
    usage, Bitcoin bridge usage, and webhook service-binding behavior.
  - Keep `receipt-webhook-mock` as a separate helper worker for the internal binding path in this
    slice.
- Timeline/cost constraints:
  - Reuse the existing Cloudflare bootstrap and bridge code where possible.
- Compliance/security constraints:
  - Public HTTP routes must remain limited to the existing API surface.
  - Worker secret sync must continue to handle PostgreSQL, xpub, Esplora, and webhook secrets
    explicitly.

## Problem statement

- Current pain:
  - One logical payrune backend service is currently represented by multiple Cloudflare workers,
    multiple Wasm binaries, and multiple deploy/delete scripts.
  - The split makes the Cloudflare resource list noisy and pushes operator attention toward worker
    plumbing instead of the product boundary that the team actually manages.
  - Existing worker shells duplicate loader and bridge setup patterns even though they all invoke
    the same internal bootstrap layer.
- Evidence or examples:
  - `cf-up` deploys `payrune-api`, `payrune-poller-mainnet`, `payrune-poller-testnet4`,
    `receipt-webhook-mock`, and `payrune-webhook-dispatcher` separately today.
  - API, poller, and dispatcher each ship their own deployment directory, package metadata,
    runtime loader, Wasm build script, and Wrangler lifecycle scripts.

## Goals

- G1:
  - Consolidate the payrune API, poller, and webhook dispatcher into one Cloudflare worker
    deployment named and operated as a single payrune service.
- G2:
  - Preserve current API behavior, poller network scopes, dispatcher behavior, and internal service
    binding semantics while reducing deployment-shell duplication.
- G3:
  - Keep runtime-specific code boundaries explicit inside Go and JavaScript even though the
    deployment target becomes one worker.

## Non-goals (out of scope)

- NG1:
  - Changing application/domain behavior for address allocation, polling lifecycle, or webhook
    dispatch policy.
- NG2:
  - Replacing the existing `receipt-webhook-mock` service-binding target with a different test or
    production transport.
- NG3:
  - Redesigning database schema, migrations, or the Cloudflare PostgreSQL bridge contract.

## Assumptions

- A1:
  - A shared Cloudflare worker failure domain for the payrune API, poller, and dispatcher is
    acceptable and desired by the maintainers.
- A2:
  - One worker can host the public `fetch()` surface and the scheduled job entrypoint without
    requiring the internal Clean Architecture boundaries to collapse.
- A3:
  - Poller mainnet, poller testnet4, and dispatcher schedules can be routed deterministically from
    one Worker `scheduled()` entrypoint.

## Open questions

- Q1:
  - Should the legacy split worker directories and scripts be deleted immediately as part of the
    consolidation, or temporarily kept only if verification shows a hidden dependency?
- Q2:
  - Should the worker be named `payrune`, matching the service boundary, or keep a longer
    Cloudflare-specific name? Assumption for this spec: use `payrune`.

## Success metrics

- Metric:
  - Cloudflare worker count for the core payrune service deployment shell.
- Target:
  - API, poller, and dispatcher are deployed through one Cloudflare worker shell and one primary
    deploy/delete script pair.
- Metric:
  - Runtime behavior parity.
- Target:
  - Existing API JS tests, scheduler JS tests, and targeted Go tests pass after consolidation.
- Metric:
  - Deployment-shell duplication.
- Target:
  - One payrune Cloudflare deployment directory owns the Wasm loader, Wrangler config, package
    metadata, and deployment-focused tests for API plus scheduled jobs.
