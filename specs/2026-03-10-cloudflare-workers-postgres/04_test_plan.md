---
doc: 04_test_plan
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

# Test Plan

## Scope

- Go-Wasm Worker public API
- Cloudflare PostgreSQL adapter for create/status flow
- Deployment shell thinness

## Test Cases

### TC-001 Health route

- Linked requirements: FR-001, FR-002, FR-003
- Method:
  - Run focused Go tests or smoke checks for the Go-Wasm health route.
- Expected result:
  - Worker returns `200` with the current health JSON contract.

### TC-002 Address policy listing

- Linked requirements: FR-001, FR-002, FR-003
- Method:
  - Run focused Go tests or smoke checks for `GET /v1/chains/bitcoin/address-policies`.
- Expected result:
  - Worker returns policy metadata with correct `enabled` values from env.

### TC-003 Address generation

- Linked requirements: FR-001, FR-002, FR-003
- Method:
  - Run focused Go tests or smoke checks for `GET /v1/chains/bitcoin/addresses`.
- Expected result:
  - Worker validates input and derives the same address families expected by the contract.

### TC-004 Payment address creation

- Linked requirements: FR-001, FR-002, FR-003, FR-006
- Method:
  - Run create-flow tests through the Go-Wasm runtime and Cloudflare PostgreSQL adapter for success, invalid body, disabled policy, idempotency replay, and idempotency conflict.
- Expected result:
  - Status codes, headers, and payloads match the current contract.

### TC-005 Payment address status lookup

- Linked requirements: FR-001, FR-002, FR-003, FR-006
- Method:
  - Run status lookup tests through the Go-Wasm runtime and Cloudflare PostgreSQL adapter for found/not-found/invalid-id paths.
- Expected result:
  - Worker returns the current status contract or the expected errors.

### TC-006 Deploy helper shape

- Linked requirements: FR-005
- Method:
  - Run `make -n cf-up` and `make -n cf-down`.
  - Manually review deploy output for both cases:
    - `POSTGRES_CONNECTION_STRING` provided via `.env.cloudflare`
    - `POSTGRES_CONNECTION_STRING` missing
- Expected result:
  - `cf-up` stacks migration plus API/poller deploy scripts in the expected order, `cf-down` stacks teardown scripts in the expected order, `.env.cloudflare` is auto-loaded, missing `POSTGRES_CONNECTION_STRING` fails fast, optional xpub/config envs can stay blank, and deploy status output remains colorized.

### TC-007 Shell thinness

- Linked requirements: FR-004, NFR-001, NFR-002
- Method:
  - Inspect deployment entrypoint and import structure.
- Expected result:
  - Deployment code stays a thin runtime shell and does not own route/business-flow logic.
