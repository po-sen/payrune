---
doc: 00_problem
spec_date: 2026-03-10
slug: cloudflare-workers-postgres
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-08-payment-address-idempotency-key
  - 2026-03-08-payment-address-status-api
  - 2026-03-09-receipt-expire-final-check
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background: Payrune already has a working Go API. This slice must deploy that public API to Cloudflare Workers without a separately running Go origin service.
- Users or stakeholders: Payrune operators and backend teams.
- Why now: The previous Cloudflare exploration drifted into thin-edge proxying and required an origin URL, which does not satisfy the requirement to run the API directly on Cloudflare.

## Constraints (optional)

- Technical constraints:
  - The deployed runtime must be a standalone Cloudflare Worker.
  - External PostgreSQL remains the persistence backend.
  - `deployments/cloudflare/payrune-api/` must stay a thin runtime shell; future API feature work should primarily happen outside `deployments/`.
  - Plain Worker handlers are preferred; no Hono.
- Timeline/cost constraints:
  - Reuse current API contract and persistence schema rather than redesigning the product surface.
  - Keep changes tightly scoped to Cloudflare-related code.
- Compliance/security constraints:
  - Security hardening beyond the minimum runnable Worker API can be handled in later slices.

## Problem statement

- Current pain:
  - A thin-edge Worker that forwards to a Go origin does not satisfy the requirement to run independently on Cloudflare.
  - A standalone JS Worker reimplementation drifts from the existing Go application layer and does not actually reuse Go use cases.
  - The final direction must keep Cloudflare-only deployment while still running the existing Go use cases.
- Evidence or examples:
  - A Worker-only deployment must not require `PAYRUNE_ORIGIN_BASE_URL`.
  - Reimplementing `POST /v1/chains/bitcoin/payment-addresses` in JS is not acceptable when the source of truth already exists in Go use cases.

## Goals

- G1: Run the public Payrune API directly inside Cloudflare Workers with no separate Go origin.
- G2: Execute the existing Go use cases, not a parallel JS business implementation.
- G3: Keep `deployments/cloudflare/payrune-api/` as deployment/bootstrap/runtime shell only.
- G4: Ensure future API feature work usually changes Go code, not `deployments/`.
- G5: Keep the diff limited to Cloudflare-related code and remove leftover wrong-direction experiments.

## Non-goals (out of scope)

- NG1: Moving poller or receipt webhook dispatcher into Cloudflare in this slice.
- NG2: Replacing PostgreSQL with D1 or SQLite.
- NG3: Adding edge auth, origin hardening, or other security controls in this slice.
- NG4: Rewriting business logic into standalone JS.
- NG5: Changing payment semantics, status transitions, or database schema beyond what Worker execution needs.

## Assumptions

- A1: Operators can provide a PostgreSQL connection string to the Worker.
- A2: The Worker can use a JS bridge to access PostgreSQL while Go/Wasm remains the application runtime.
- A3: The current Go API behavior and use cases remain the source contract that the Worker implementation must run.

## Open questions

- Q1: None for this slice.

## Success metrics

- Metric: Worker runs the public API without requiring an origin URL.
- Target: Public API routes succeed against Worker runtime with PostgreSQL configured.
- Metric: Deployment shell remains thin.
- Target: `deployments/cloudflare/payrune-api/` only contains Wrangler/bootstrap/runtime shell code, not business-flow implementation.
- Metric: Future API route work stays out of deployment shell.
- Target: New `/v1/...` behavior can be added primarily in Go code under `cmd/` and `internal/`.
