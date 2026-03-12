---
doc: 03_tasks
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

# Task Plan

## Mode decision

- Selected mode: Full
- Rationale: This change affects runtime shape, Cloudflare deployment behavior, PostgreSQL access, and future development boundaries.

## Tasks (ordered)

1. T-001 - Re-baseline the Cloudflare spec

   - Scope: Update the spec from thin-edge or standalone JS directions to `Cloudflare-only + Go/Wasm + reuse Go use cases`.
   - Linked requirements: FR-001, FR-002, FR-004, FR-007, NFR-001, NFR-002
   - Validation:
     - [ ] Review the spec.
     - [ ] Confirm no requirement still assumes a Go origin.
     - [ ] Confirm no requirement still assumes standalone JS business logic.

2. T-002 - Freeze deployment shell

   - Scope: Keep `deployments/cloudflare/payrune-api/` to Wrangler config, JS shell, Go-Wasm loader, PostgreSQL bridge, and deployment-focused tests/docs only.
   - Linked requirements: FR-004, NFR-001, NFR-002
   - Validation:
     - [ ] `deployments/cloudflare/payrune-api/src/index.mjs` stays thin.
     - [ ] Future `/v1/...` route additions do not require touching deployment shell code.
     - [ ] The Wasm artifact is imported directly instead of being base64-wrapped into JavaScript.

3. T-003 - Implement Go-Wasm worker entrypoint

   - Scope: Add a Go-Wasm entrypoint plus Worker inbound adapter that dispatches to the existing Go controllers and use cases.
   - Linked requirements: FR-001, FR-002, FR-003, NFR-002
   - Validation:
     - [ ] Focused Go tests for Worker inbound adapter pass.
     - [ ] Go-Wasm build succeeds.

4. T-004 - Implement Cloudflare PostgreSQL outbound adapter

   - Scope: Add a Worker-compatible PostgreSQL adapter that implements the outbound ports used by the public Go use cases.
   - Linked requirements: FR-002, FR-006
   - Validation:
     - [ ] Focused Go tests for `internal/adapters/outbound/persistence/cloudflarepostgres` pass.
     - [ ] `POST /v1/chains/bitcoin/payment-addresses` can run through Go use cases in Worker runtime.

5. T-005 - Wire deploy and teardown flows

   - Scope: Expose a unified `make cf-up` / `make cf-down` flow that stacks the Cloudflare scripts, with repo-local `.env.cloudflare` loading, fail-fast required env checks, optional xpub secret sync, and Go-Wasm artifact build.
   - Linked requirements: FR-005
   - Validation:
     - [ ] `make -n cf-up`
     - [ ] `make -n cf-down`

6. T-006 - Remove obsolete wrong-direction code and verify clean scope

   - Scope: Remove origin proxy code, standalone JS business logic, and unrelated non-Cloudflare changes.
   - Linked requirements: FR-007, NFR-003
   - Validation:
     - [ ] `git diff --stat` only shows Cloudflare-related files and intentionally touched deploy hooks.
     - [ ] `SPEC_DIR="specs/2026-03-10-cloudflare-workers-postgres" bash scripts/spec-lint.sh`

## Traceability (optional)

- FR-001 -> T-001, T-003
- FR-002 -> T-001, T-003, T-004
- FR-003 -> T-003
- FR-004 -> T-001, T-002
- FR-005 -> T-005
- FR-006 -> T-004
- FR-007 -> T-001, T-006
- NFR-001 -> T-001, T-002
- NFR-002 -> T-001, T-002, T-003
- NFR-003 -> T-006
