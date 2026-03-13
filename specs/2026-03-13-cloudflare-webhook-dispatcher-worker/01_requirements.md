---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Dispatcher Worker:
  - A scheduled Cloudflare Worker that runs one receipt webhook dispatch cycle.
- Mock worker:
  - A Cloudflare Worker that receives dispatcher deliveries, verifies the signature, logs the
    request, and returns a deterministic result.

## Out-of-scope behaviors

- OOS1:
  - Rewriting webhook delivery in JavaScript.
- OOS2:
  - Adding worker-to-worker calls between dispatcher and other payrune workers.

## Functional requirements

### FR-001 - Standalone dispatcher worker runtime

- Description:
  - A standalone Cloudflare Worker deployment shell must exist for receipt webhook dispatch.
- Acceptance criteria:
  - [ ] A new `deployments/cloudflare/payrune-webhook-dispatcher/` worker shell exists.
  - [ ] The worker supports scheduled execution and does not expose public API routes.
- Notes:
  - The worker is cron-driven only.

### FR-002 - Reuse Go webhook dispatch use case

- Description:
  - The Cloudflare dispatcher worker must execute the existing Go
    `RunReceiptWebhookDispatchCycleUseCase`.
- Acceptance criteria:
  - [ ] A Go/Wasm entrypoint exists for the webhook dispatcher runtime.
  - [ ] The worker invokes the Go dispatch handler instead of duplicating dispatch logic in JS.
- Notes:
  - JS should remain bridge/shell code only.

### FR-003 - Cloudflare PostgreSQL and notifier adapters

- Description:
  - The worker runtime must provide the outbound adapters needed by the dispatch use case.
- Acceptance criteria:
  - [ ] The existing Cloudflare PostgreSQL adapter is wired for outbox claim/save flows.
  - [ ] The Cloudflare PostgreSQL scan layer supports `time.Time`, `*time.Time`, and
        `sql.NullTime` destinations required by webhook outbox rows.
  - [ ] A Cloudflare-compatible payment receipt status notifier exists for outbound webhook HTTP.
  - [ ] The notifier always uses an explicit Cloudflare binding transport mode for the internal
        mock path without relying on a fake internal URL.
- Notes:
  - The notifier may use a JS bridge or another Cloudflare-appropriate transport layer.

### FR-004 - Deployable Cloudflare receipt webhook mock worker

- Description:
  - The repo must include a standalone Cloudflare worker named `receipt-webhook-mock`.
- Acceptance criteria:
  - [ ] `deployments/cloudflare/receipt-webhook-mock/` exists with Wrangler config, JS entrypoint,
        tests, and docs.
  - [ ] The worker exposes `POST /receipt-status` and returns `204` for a valid request.
  - [ ] Non-matching routes return `404`.
- Notes:
  - This worker is a Cloudflare-only test helper.

### FR-005 - Deploy and teardown automation

- Description:
  - Repository-level Cloudflare automation must include dispatcher and mock worker deploy/delete
    flows.
- Acceptance criteria:
  - [ ] There is a dedicated deploy script for the dispatcher worker under `scripts/`.
  - [ ] There is a dedicated delete script for the dispatcher worker under `scripts/`.
  - [ ] There is a dedicated deploy script for the mock worker under `scripts/`.
  - [ ] There is a dedicated delete script for the mock worker under `scripts/`.
  - [ ] Top-level Cloudflare orchestration deploys mock before dispatcher and deletes dispatcher
        before mock.
- Notes:
  - Follow the existing `cf-*` script naming pattern.

### FR-006 - Worker defaults and configuration

- Description:
  - Non-secret dispatcher defaults must live in `wrangler.toml`; secrets must continue to come
    from the existing `.env.cloudflare` / Wrangler secret flow.
- Acceptance criteria:
  - [ ] `wrangler.toml` includes explicit defaults for dispatch cadence and batch config.
  - [ ] Deploy flow syncs required secrets without interactive prompt expansion beyond current
        Cloudflare conventions.
  - [ ] `.env.cloudflare.example` does not include `PAYMENT_RECEIPT_WEBHOOK_URL`.
- Notes:
  - Match the operational style already used by API and poller workers.

## Non-functional requirements

- Reliability (NFR-001):
  - Delivery semantics must match the current use case behavior for sent / retry / failed
    transitions.
- Security (NFR-002):
  - Shared webhook secret remains a Wrangler secret and is not moved into plain vars.
- Observability (NFR-003):
  - Dispatcher and mock workers must emit Cloudflare invocation logs by default.
- Maintainability (NFR-004):
  - Future dispatch feature work should primarily touch Go code under `cmd/`, `internal/adapters/`,
    `internal/application/`, and `internal/domain`, not `deployments/cloudflare/`.
- Architecture (NFR-005):
  - Dispatcher depends on a mock worker through a Cloudflare service binding, but must not depend
    on `payrune-api` or `payrune-poller`.

## Dependencies and integrations

- External systems:
  - Cloudflare Workers cron runtime.
  - PostgreSQL.
- Internal services:
  - Existing Go webhook dispatch use case and outbox store behavior.
  - `receipt-webhook-mock` worker.
