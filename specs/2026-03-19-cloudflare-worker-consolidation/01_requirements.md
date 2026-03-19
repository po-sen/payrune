---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Unified payrune worker:
  - The single Cloudflare worker deployment that owns public API handling plus scheduled poller and
    dispatcher execution.
- Scheduled job route:
  - The deterministic mapping from a Cloudflare cron trigger to either the mainnet poller,
    testnet4 poller, or webhook dispatcher handler.

## Out-of-scope behaviors

- OOS1:
  - Changing the API route table, response payloads, or controller semantics.
- OOS2:
  - Changing poller cycle logic, dispatcher retry policy, or webhook payload schema.
- OOS3:
  - Eliminating `receipt-webhook-mock` in this implementation slice.

## Functional requirements

### FR-001 - Single Cloudflare worker deployment shell

- Description:
  - API, poller, and webhook dispatcher must deploy through one Cloudflare worker shell and one
    unified Wasm runtime entrypoint.
- Acceptance criteria:
  - [ ] A new Cloudflare deployment directory exists for the unified payrune worker.
  - [ ] One unified Wasm build target exists for the Cloudflare API, poller, and dispatcher entry
        handlers.
  - [ ] `make cf-up` deploys the payrune service runtime through one primary worker deploy script
        instead of three separate scripts for API, poller, and dispatcher.
- Notes:
  - `receipt-webhook-mock` remains separate in this slice.

### FR-002 - Preserve public API behavior

- Description:
  - The unified worker must preserve the existing public HTTP route filtering and API request/response
    behavior.
- Acceptance criteria:
  - [ ] The worker still serves `/health` and `/v1/...` routes and returns `404` for unsupported
        public paths.
  - [ ] The Cloudflare API request envelope contract remains compatible with the existing Go
        bootstrap path.
  - [ ] API JS tests continue to verify request-envelope and response mapping behavior.
- Notes:
  - This is a runtime-shell consolidation, not an API redesign.

### FR-003 - Preserve scheduled poller and dispatcher behavior

- Description:
  - The unified worker must preserve existing poller and dispatcher one-cycle execution behavior
    while routing all scheduled triggers through one `scheduled()` entrypoint.
- Acceptance criteria:
  - [ ] Scheduled triggers for Bitcoin mainnet poller, Bitcoin testnet4 poller, and webhook
        dispatcher are all configured in the unified worker.
  - [ ] Each scheduled trigger is routed deterministically to the correct bootstrap handler with
        the required envelope fields and bridge registrations.
  - [ ] Unsupported or unmapped scheduled triggers return a clear runtime error instead of silently
        running the wrong job.
- Notes:
  - Poller network scope must remain explicit; no inferred network fallback is allowed.

### FR-004 - Preserve bridge and binding behavior

- Description:
  - The unified worker must continue using the correct technical bridges for each runtime path.
- Acceptance criteria:
  - [ ] API requests register only the PostgreSQL bridge they need.
  - [ ] Poller scheduled runs register PostgreSQL plus Bitcoin observer bridges.
  - [ ] Dispatcher scheduled runs register PostgreSQL plus webhook notifier bridge and continue to
        use the `RECEIPT_WEBHOOK_MOCK` service binding path.
- Notes:
  - Bridge registration should stay scoped to the runtime path being executed.

### FR-005 - Consolidate deployment automation and docs

- Description:
  - Repo automation and Cloudflare deployment docs must describe and operate the unified payrune
    worker shape.
- Acceptance criteria:
  - [ ] Top-level deploy/delete flows no longer mention separate API, poller, and dispatcher worker
        deploy scripts.
  - [ ] The unified worker README documents public API handling, scheduled job routing, required
        secrets, and cron coverage.
  - [ ] References in active operational docs point to the unified worker deployment shape.
- Notes:
  - Historical specs remain as historical records and do not need retroactive content rewrites.

## Non-functional requirements

- Performance (NFR-001):
  - Scheduled-job dispatch selection inside the unified worker must remain in-process and
    constant-time; no extra HTTP hop or database lookup may be introduced just to determine which
    job to run.
- Availability/Reliability (NFR-002):
  - The unified worker must emit a deterministic error for unmapped cron triggers, and existing
    poller/dispatcher summary log messages must still be reachable through targeted tests.
- Security/Privacy (NFR-003):
  - Only the existing public API routes may be reachable through `fetch()`; scheduled runtime
    functionality must not expose new public endpoints.
- Compliance (NFR-004):
  - No new compliance requirements apply in this slice.
- Observability (NFR-005):
  - API error logs, poller cycle logs, and dispatcher cycle logs must remain distinguishable by
    message content after consolidation.
- Maintainability (NFR-006):
  - Duplicate Cloudflare deployment-shell code for API, poller, and dispatcher should be reduced to
    one runtime loader, one Wrangler config, one primary package manifest, and one primary deploy
    script pair for the payrune service.

## Dependencies and integrations

- External systems:
  - Cloudflare Workers `fetch()` and `scheduled()` runtime surfaces.
  - Cloudflare Wrangler CLI.
  - PostgreSQL accessed through the Cloudflare JS bridge.
  - Bitcoin Esplora APIs for poller observation.
- Internal services:
  - Existing Go bootstrap handlers under `internal/bootstrap`.
  - `receipt-webhook-mock` Cloudflare worker service binding.
