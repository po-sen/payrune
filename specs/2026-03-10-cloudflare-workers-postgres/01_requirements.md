---
doc: 01_requirements
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

# Requirements

## Functional Requirements

### FR-001 Standalone Cloudflare Worker API

The Cloudflare Worker MUST execute the current public API directly without relying on a separately running Go origin service.

Acceptance criteria:

- The Worker serves `GET /health`.
- The Worker serves `GET /v1/chains/bitcoin/address-policies`.
- The Worker serves `GET /v1/chains/bitcoin/addresses`.
- The Worker serves `POST /v1/chains/bitcoin/payment-addresses`.
- The Worker serves `GET /v1/chains/bitcoin/payment-addresses/{paymentAddressId}`.
- The Worker does not require `PAYRUNE_ORIGIN_BASE_URL`.

### FR-002 Execute Go use cases

The Worker MUST execute the existing Go application layer for public API behavior.

Acceptance criteria:

- Public API flows are driven by the existing Go use cases rather than standalone JS business logic.
- The Worker runtime reuses the existing Go controller/use-case contract for:
  - health
  - address policies
  - address generation
  - payment address allocation
  - payment address status lookup

### FR-003 Contract parity

The Worker MUST preserve the current API contract and error semantics for the public routes it serves.

Acceptance criteria:

- Response JSON fields match the existing public API contract.
- Method restrictions still return `405` with `Allow` headers where applicable.
- Validation and business error mappings remain equivalent to the current Go API contract.
- `POST /v1/chains/bitcoin/payment-addresses` still supports `Idempotency-Key` and `Idempotency-Replayed`.

### FR-004 Deployment shell only

The deployment directory MUST remain a thin shell around the actual Worker application code.

Acceptance criteria:

- `deployments/cloudflare/payrune-api/` contains Wrangler config, thin bootstrap entrypoint, Go-Wasm loader, PostgreSQL JS bridge, package metadata, and deployment-focused tests/docs only.
- Main route handling and payment business flow implementation live outside `deployments/`.

### FR-005 Deploy and teardown entrypoints

The repo MUST provide simple deploy and delete flows for the Worker.

Acceptance criteria:

- `make cf-up` runs the shared Cloudflare migration and deploys the API Worker as part of the unified Cloudflare rollout flow.
- `make cf-down` deletes the API Worker as part of the unified Cloudflare teardown flow.
- `make cf-migrate` runs the shared Cloudflare PostgreSQL migration independently.
- Cloudflare deploy/migrate scripts auto-load repo-local `.env.cloudflare` when present.
- The repo provides `.env.cloudflare.example` as the local Cloudflare env template.
- `POSTGRES_CONNECTION_STRING` is required for `make cf-up` and `make cf-migrate`; missing it must fail fast instead of prompting.
- Optional xpub envs may be left blank and skipped during Worker secret sync.
- Non-sensitive Bitcoin confirmation and receipt-expiry defaults live in `wrangler.toml`, not in deploy-time secret sync.
- Cloudflare deployment docs and ignore rules reference only repo-root `.env.cloudflare`, not deployment-local `.env.local` files.
- Deploy clearly announces that `POSTGRES_CONNECTION_STRING` Worker secret sync will run before it starts the build/test/deploy steps.
- In an interactive terminal, deploy status messages use terminal colors so the operator can distinguish steps, warnings, and success quickly.
- Deploy builds the Go-Wasm artifact before publishing the Worker.

### FR-006 Worker-side PostgreSQL adapter

The Worker MUST provide a Cloudflare-compatible PostgreSQL adapter so Go use cases can run without `database/sql` at runtime.

Acceptance criteria:

- Go use cases can allocate payment addresses and look up payment address status inside Worker runtime.
- The Worker runtime does not depend on `database/sql` or a separately running Go process.

### FR-007 Cloudflare-only scope

The implementation MUST not leave unrelated thin-edge or origin-specific runtime code in place for this slice.

Acceptance criteria:

- No Worker code depends on an origin URL.
- No thin-edge proxy code remains as the active implementation.
- No unrelated Go server/runtime changes remain in the final diff.

## Non-Functional Requirements

### NFR-001 Deployment decoupling

Future public API feature work should usually not require editing `deployments/cloudflare/payrune-api/`.

Acceptance criteria:

- Adding or changing Worker route logic can be done primarily outside `deployments/`.
- Deployment bootstrap stays stable after this slice.

### NFR-002 Runtime simplicity

The standalone Worker should use the smallest practical Cloudflare runtime model while still executing Go use cases.

Acceptance criteria:

- Plain Worker handlers are used.
- A single generic Go-Wasm runtime shell is used.
- No thin-edge proxy/origin runtime path remains.

### NFR-003 Verification

The slice must be verifiable with focused local checks.

Acceptance criteria:

- Worker syntax and unit tests pass.
- Relevant Go validation for Cloudflare-specific code passes.
- Spec lint passes.
