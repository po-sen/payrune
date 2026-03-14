---
doc: 02_design
spec_date: 2026-03-14
slug: runtime-entrypoint-alignment
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-10-cloudflare-workers-postgres
  - 2026-03-11-cloudflare-poller-workers
  - 2026-03-12-api-worker-naming-unification
  - 2026-03-13-cloudflare-webhook-dispatcher-worker
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Design

## Summary

- Split runtime entrypoint responsibilities by layer:
  - `cmd/*` remains runtime glue only.
  - `internal/bootstrap` owns runtime orchestration.
  - `internal/adapters/inbound/http` owns public HTTP transport mapping.
  - `internal/adapters/inbound/scheduler` owns one-cycle scheduler transport mapping.
  - `internal/infrastructure/di` wires concrete handlers and runtime dependencies.
- Normalize bootstrap naming so runtime nouns match file/function/type names.

## Target package layout

- `internal/adapters/inbound/http/handler.go`
  - Shared public HTTP handler assembly point.
- `internal/adapters/inbound/http/cloudflare/bridge.go`
  - Cloudflare API request/response bridge to the shared HTTP handler.
- `internal/adapters/inbound/scheduler/*.go`
  - Poller and receipt webhook dispatcher one-cycle handlers.
- `internal/bootstrap/api.go`
  - Standalone API runtime orchestration.
- `internal/bootstrap/api_worker.go`
  - Cloudflare API worker request orchestration.
- `internal/bootstrap/poller.go`
  - Standalone poller loop orchestration.
- `internal/bootstrap/poller_worker.go`
  - Cloudflare poller worker request orchestration.
- `internal/bootstrap/receipt_webhook_dispatcher.go`
  - Standalone receipt webhook dispatcher loop orchestration.
- `internal/bootstrap/receipt_webhook_dispatcher_worker.go`
  - Cloudflare receipt webhook dispatcher worker request orchestration.

## Flows

### Standalone API runtime

1. `cmd/api/main.go` calls `bootstrap.RunAPI`.
2. Bootstrap asks DI for the public API handler.
3. Shared HTTP assembly registers controllers and middleware.
4. `http.Server` serves the handler.

### Cloudflare API worker runtime

1. `cmd/api-worker/main.go` provides JS runtime glue only.
2. It delegates payload handling to `bootstrap.HandleCloudflareAPIRequestJSON`.
3. Bootstrap decodes JSON, builds the shared HTTP handler through DI, and calls the HTTP
   Cloudflare bridge.
4. The existing response envelope shape is returned unchanged.

### Standalone scheduler runtimes

1. `cmd/poller/main.go` and `cmd/webhook-dispatcher/main.go` parse env/config and call bootstrap.
2. Bootstrap owns ticker lifecycle.
3. Each tick is executed through the shared scheduler handlers.

### Cloudflare scheduler worker runtimes

1. Worker `cmd/*` packages provide JS runtime glue only.
2. Bootstrap decodes payload JSON and builds scheduler handlers through DI.
3. Scheduler handlers execute one cycle and return the existing response payloads.

## Contracts

- Public API routes and response semantics are unchanged.
- Cloudflare API request envelope remains:
  - `request`, `env`, `bridgeId`
- Cloudflare poller worker request envelope remains:
  - `env`, `postgresBridgeId`, `bitcoinBridgeId`, `scheduledTime`, `cron`
- Cloudflare webhook dispatcher worker request envelope remains:
  - `env`, `postgresBridgeId`, `scheduledTime`, `cron`
- Scheduler response counter fields remain unchanged.

## Failure modes

- Nil shared HTTP or scheduler handlers must return clear configuration errors rather than panic.
- Partial rename application must be caught by compile/test/search verification.
- Removing the mixed package without updating imports would break builds, so import/search checks are
  part of validation.

## Observability

- No new logging or metrics behavior is introduced.
- Existing cycle log lines and API behavior remain intact.

## Security

- Public HTTP middleware and validation stay in the shared HTTP handler path for all API runtimes.
- This refactor does not alter secret handling or external transport security posture.

## Trade-offs

- The refactor prefers concrete runtime-specific bootstrap helpers over abstract frameworks.
- One merged spec is less granular historically, but it is a clearer source of truth for the actual
  implementation that landed.
