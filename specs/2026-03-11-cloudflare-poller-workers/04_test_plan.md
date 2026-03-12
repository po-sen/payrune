---
doc: 04_test_plan
spec_date: 2026-03-11
slug: cloudflare-poller-workers
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-05-blockchain-receipt-polling-service
  - 2026-03-09-shared-tip-height-polling
  - 2026-03-09-poller-interval-separation
  - 2026-03-09-receipt-expire-final-check
  - 2026-03-10-cloudflare-workers-postgres
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
  - Cloudflare standalone poller Worker runtime.
  - Go/Wasm scheduled poller entrypoint.
  - Worker-compatible PostgreSQL and Bitcoin observer paths.
  - Deploy/delete flows for mainnet and testnet4 pollers.
- Not covered:
  - Receipt webhook dispatcher migration to Cloudflare.
  - Cloudflare security hardening beyond basic deployability.

## Tests

### Unit

- TC-001:

  - Linked requirements: FR-002, FR-003, FR-008, NFR-002
  - Steps:
    - Run focused Go tests for the Go/Wasm poller entrypoint and scheduled-event mapping.
  - Expected:
    - A scheduled event invokes the existing Go polling use case with the correct chain/network scope and preserves current output counters.

- TC-002:
  - Linked requirements: FR-006, FR-007, NFR-004
  - Steps:
    - Run focused Go tests for Worker-compatible PostgreSQL poller paths and the Worker-compatible Bitcoin observer adapter.
  - Expected:
    - Claim/save polling paths and Bitcoin observation paths work under Worker runtime assumptions.

### Integration

- TC-101:

  - Linked requirements: FR-001, FR-003, FR-004, NFR-001, NFR-005
  - Steps:
    - Run Worker shell syntax/tests and inspect deployment-shell import structure.
  - Expected:
    - Deployment code remains thin and supports separate mainnet/testnet4 deployment targets.

- TC-102:
  - Linked requirements: FR-005, NFR-003, NFR-005
- Steps:
  - Run `make -n cf-up` and `make -n cf-down`.
- Expected:
  - `cf-up` stacks migration plus both poller deploy scripts in the expected order, `cf-down` stacks both poller delete scripts in the expected order, `.env.cloudflare` is auto-loaded, missing `POSTGRES_CONNECTION_STRING` fails fast, deploy output clearly states Worker secret sync behavior, and optional deploy-time secret sync is limited to Esplora auth secrets.

### E2E (if applicable)

- Scenario 1:
  - Deploy `payrune-poller-mainnet` and confirm one scheduled run executes a polling cycle with `chain=bitcoin`, `network=mainnet`.
- Scenario 2:
  - Deploy `payrune-poller-testnet4` and confirm one scheduled run executes a polling cycle with `chain=bitcoin`, `network=testnet4`.

## Edge cases and failure modes

- Case:

  - Missing `POSTGRES_CONNECTION_STRING`.
  - Expected behavior:
    - Worker poller fails deterministically and logs an explicit bootstrap or persistence error.

- Case:

  - Missing Esplora configuration for one network.
  - Expected behavior:
    - Poll cycle reports processing errors using the current polling semantics; it does not silently succeed.

- Case:
  - Overlapping scheduled runs for the same worker.
  - Expected behavior:
    - Existing claim/lease behavior prevents duplicate processing from corrupting receipt state.

## NFR verification

- Performance:
  - Confirm scheduled poller runtime preserves the current shared tip-height optimization and does not reintroduce per-address tip fetches.
- Reliability:
  - Confirm final-check expiry ordering and existing processing-error semantics remain unchanged.
- Security:
  - Confirm PostgreSQL and Esplora credentials remain secret-driven and are not hard-coded in Worker shell code.
