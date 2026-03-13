---
doc: 00_problem
spec_date: 2026-03-13
slug: cloudflare-webhook-dispatcher-worker
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-06-receipt-webhook-delivery
  - 2026-03-10-cloudflare-workers-postgres
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
  - `payrune-api` and `payrune-poller` already run as standalone Cloudflare Workers.
  - `payrune-receipt-webhook-dispatcher` still runs as a separate long-running Go process.
  - Local compose has a `receipt-webhook-mock`, but Cloudflare deployment needs its own internal
    mock target.
- Users or stakeholders:
  - payrune maintainers operating Cloudflare-based deployment.
- Why now:
  - The remaining receipt webhook dispatcher should match the Cloudflare-only deployment shape.

## Constraints (optional)

- Technical constraints:
  - Reuse the existing Go `RunReceiptWebhookDispatchCycleUseCase`.
  - Keep `deployments/cloudflare/` as thin deployment/runtime shell code.
  - Avoid introducing a dependency on `payrune-api` worker for webhook dispatch.

## Problem statement

- Current pain:
  - Webhook delivery still requires a separate process runtime even though API and poller have
    already moved to Cloudflare Workers.
- Evidence or examples:
  - Current runtime lives in `cmd/webhook-dispatcher/main.go` plus
    `internal/bootstrap/receipt_webhook_dispatcher.go`.

## Goals

- G1:
  - Run receipt webhook dispatch cycles in a standalone Cloudflare Worker.
- G2:
  - Reuse the existing Go webhook dispatch use case rather than reimplementing delivery logic in JS.
- G3:
  - Keep future webhook dispatch feature work primarily in Go, not in `deployments/cloudflare/`.
- G4:
  - Provide a Cloudflare-native mock webhook target and make it the default dispatcher destination.

## Non-goals (out of scope)

- NG1:
  - Replacing webhook delivery with Cloudflare Queues or Workflows.
- NG2:
  - Making the dispatcher call `payrune-api` or `payrune-poller` workers as part of dispatch.
- NG3:
  - Changing webhook payload contract, retry policy, or delivery semantics.

## Assumptions

- A1:
  - Dispatcher can run as a scheduled Cloudflare Worker similar to the existing poller worker.
- A2:
  - The best path is direct PostgreSQL access plus direct outbound webhook fetch from the worker.
- A3:
  - The dispatcher always targets a Cloudflare `receipt-webhook-mock` worker through a service
    binding in this runtime.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - Receipt webhook dispatch no longer requires the standalone process runtime.
- Target:
  - A dedicated Cloudflare Worker can run the dispatch cycle end-to-end.
- Metric:
  - Future feature work remains mostly in Go.
- Target:
  - Deployment shell stays thin and route-free; business behavior lives in Go use case / adapters.
- Metric:
  - Default Cloudflare stack no longer needs a public webhook URL.
- Target:
  - `cf-up` deploys mock + dispatcher and dispatcher uses the internal binding path by default.
