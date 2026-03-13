---
doc: 02_design
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

# Technical Design

## High-level approach

- Summary:
  - Add a standalone Cloudflare dispatcher worker plus a JS-only `receipt-webhook-mock` worker,
    with dispatcher delivery fixed to the internal mock binding path.
- Key decisions:
  - Reuse `RunReceiptWebhookDispatchCycleUseCase`.
  - Keep deployment shell thin and route-free.
  - Use direct PostgreSQL access and direct outbound webhook delivery from the worker.
  - Do not call `payrune-api` or `payrune-poller` workers as part of dispatch.
  - For the delivery target, use an explicit Cloudflare service-binding transport mode instead of a
    sentinel webhook URL hostname.

## System context

- Components:
  - `cmd/webhook-dispatcher-worker/`
  - `internal/infrastructure/di/cloudflare_webhook_dispatcher_worker.go`
  - `internal/adapters/inbound/cloudflareworker/`
  - `internal/adapters/outbound/persistence/cloudflarepostgres/`
  - `internal/adapters/outbound/webhook/` or a new Cloudflare-specific notifier bridge
  - `deployments/cloudflare/payrune-webhook-dispatcher/`
  - `deployments/cloudflare/receipt-webhook-mock/`
- Interfaces:
  - Scheduled invocation from Cloudflare cron.
  - PostgreSQL outbox claim/save.
  - Cloudflare service binding from dispatcher to mock worker.

## Key flows

- Flow 1:
  - Cron trigger fires -> JS shell builds scheduled envelope -> Go/Wasm handler runs one dispatch
    cycle -> logs dispatch summary.
- Flow 2:
  - Go use case claims pending notifications -> notifier bridge sends request through
    `RECEIPT_WEBHOOK_MOCK` service binding -> use case saves sent / retry / failed delivery result.
- Flow 3:
  - Mock worker verifies `X-Payrune-Signature-256`, logs the request, and returns `204` or `401`.

## Data model

- Entities:
  - No new domain entities.
- Schema changes or migrations:
  - None.
- Consistency and idempotency:
  - Reuse current outbox claim and delivery-result persistence model.

## API or contracts

- Endpoints or events:
  - No public HTTP routes.
  - Scheduled worker envelope mirrors the poller worker style: env snapshot + schedule metadata.
  - Mock worker endpoints:
    - `POST /receipt-status`
    - `GET /health`
- Request/response examples:
  - Response body is internal worker transport only; deploy shell logs output and returns 404 on
    fetch.
  - Mock delivery request is always sent through explicit binding metadata.

## Backward compatibility (optional)

- API compatibility:
  - Existing webhook payload contract remains unchanged.
- Data migration compatibility:
  - No schema change.

## Failure modes and resiliency

- Retries/timeouts:
  - Reuse current retry delay / max attempts semantics from the Go use case.
  - Webhook notifier timeout remains configurable through worker secrets/vars.
- Backpressure/limits:
  - Batch size remains configurable through `wrangler.toml`.
- Degradation strategy:
  - Failed deliveries stay in the outbox with pending/failed result according to existing policy.
  - If the service binding is missing, dispatcher fails fast with a clear runtime error.

## Observability

- Logs:
  - Dispatcher logs cycle summary and failures.
  - Mock worker logs received request summary and signature verification result.
  - Cloudflare invocation logs enabled in both `wrangler.toml` files.
- Metrics:
  - Use existing claimed / sent / retried / failed counters from use case output.
- Traces:
  - Not introduced in this change.
- Alerts:
  - Not introduced in this change.

## Security

- Authentication/authorization:
  - No inbound public surface.
- Secrets:
  - Webhook secret stays in Wrangler secret sync flow.
  - PostgreSQL connection string stays in Wrangler secret sync flow.
  - Cloudflare binding metadata is fixed runtime configuration, not secret input.
- Abuse cases:
  - Dispatcher has no public route to abuse.
  - Mock worker is a test helper and does not persist sensitive data.

## Alternatives considered

- Option A:
  - Call `payrune-api` or another internal worker through public URL.
- Option B:
  - Use a fake internal sentinel webhook URL and map it to a binding in JS shell code.
- Why chosen:
  - The Cloudflare stack is intentionally closed and internal. A fixed service-binding target is
    simpler, safer, and removes unnecessary deploy-time branching.

## Risks

- Risk:
  - Go/Wasm webhook dispatch runtime may have similar CPU/runtime sensitivity as the poller.
- Mitigation:
  - Keep batch defaults conservative and add stage-level logging if needed.
- Risk:
  - Existing Go `net/http` notifier may not be suitable for Cloudflare Go/Wasm runtime.
- Mitigation:
  - Introduce a Cloudflare-specific notifier bridge instead of forcing the existing notifier
    adapter.
- Risk:
  - Binding name drift may break dispatcher-to-mock delivery.
- Mitigation:
  - Keep explicit worker names and cover the path with deploy dry-run and JS bridge tests.
