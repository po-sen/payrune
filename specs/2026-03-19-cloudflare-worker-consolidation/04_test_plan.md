---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - Unified Cloudflare worker JS shell behavior for API requests and scheduled job routing.
  - Unified Go/Wasm command compilation and targeted bootstrap-path tests.
  - Deploy/delete script and active doc wiring updates for the unified worker.
- Not covered:
  - Live Cloudflare deployment verification against a real account.
  - End-to-end verification of external webhook delivery beyond the existing mock-binding path.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, FR-002, FR-003, NFR-006
  - Steps:
    - Run `GOCACHE=/tmp/go-build go test ./cmd/payrune-worker ./internal/bootstrap`
  - Expected:
    - The unified worker command builds and targeted bootstrap tests pass.
- TC-002:
  - Linked requirements: FR-002, FR-003, FR-004, NFR-001, NFR-002, NFR-005
  - Steps:
    - Run unified deployment-shell JS tests with `cd deployments/cloudflare/payrune && npm test`
  - Expected:
    - API route filtering, API envelope mapping, scheduled routing, and log-message helpers pass.

### Integration

- TC-101:
  - Linked requirements: FR-001, FR-005, NFR-006
  - Steps:
    - Run `cd deployments/cloudflare/payrune && npm run check`
  - Expected:
    - The unified worker shell builds its Wasm binary and Node syntax checks pass for runtime and
      bridge modules.
- TC-102:
  - Linked requirements: FR-001, FR-005, NFR-006
  - Steps:
    - Search active automation/docs with
      `rg -n "cf-api-worker|cf-poller-worker|cf-webhook-dispatcher-worker|deployments/cloudflare/payrune-api|deployments/cloudflare/payrune-poller|deployments/cloudflare/payrune-webhook-dispatcher" Makefile README.md scripts deployments/cloudflare`
  - Expected:
    - Active deploy flow references are updated to the unified payrune worker, with only
      intentional historical references left in specs.

### E2E (if applicable)

- Scenario 1:
  - Optional operator verification: deploy `receipt-webhook-mock`, deploy the unified payrune
    worker, then confirm one API request and one scheduled dispatcher run succeed.
- Scenario 2:
  - Optional operator verification: confirm scheduled poller logs still include explicit
    `chain=bitcoin network=mainnet|testnet4` context from the unified worker.

## Edge cases and failure modes

- Case:
  - An unsupported cron expression reaches the unified worker.
- Expected behavior:
  - The worker fails fast with a clear error instead of invoking the wrong handler.
- Case:
  - An API request hits a non-public path.
- Expected behavior:
  - The worker returns `404` without initializing scheduled-job-specific bridges.
- Case:
  - Dispatcher bridge registration is missing the `RECEIPT_WEBHOOK_MOCK` binding.
- Expected behavior:
  - Existing bridge tests still fail clearly, and the unified worker does not silently downgrade to
    a different transport.

## NFR verification

- Performance:
  - JS tests confirm scheduled-job selection is done by direct mapping logic without network or DB
    lookups.
- Reliability:
  - Go and JS tests confirm deterministic handler selection and explicit unmapped-cron failure.
- Security:
  - JS tests confirm only `/health` and `/v1/...` routes are public through `fetch()`.
