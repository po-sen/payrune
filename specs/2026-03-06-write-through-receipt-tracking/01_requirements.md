---
doc: 01_requirements
spec_date: 2026-03-06
slug: write-through-receipt-tracking
mode: Full
status: DONE
owners:
  - payrune-team
depends_on: []
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Requirements

## Glossary (optional)

- Write-through registration:
  - Create `payment_receipt_trackings` row in the same transaction when allocation is marked `issued`.

## Out-of-scope behaviors

- OOS1:
  - Reworking BTC observer amount aggregation logic.
- OOS2:
  - Adding per-policy `required_confirmations` configuration.

## Functional requirements

### FR-001 - Register tracking at allocation issue time

- Description:
  - Allocation issue flow must register tracking row in the same UoW transaction.
- Acceptance criteria:
  - [x] `AllocatePaymentAddressUseCase` calls receipt-tracking registration after `Complete` in the same transaction.
  - [x] Registration uses `ON CONFLICT DO NOTHING` semantics keyed by `payment_address_id`.
  - [x] If tracking registration fails, allocation issue transaction fails.

### FR-002 - Remove per-cycle register scan from poller

- Description:
  - Poller cycle should claim due trackings directly without full registration scan.
- Acceptance criteria:
  - [x] `RunReceiptPollingCycleUseCase` no longer calls `RegisterMissingIssued`.
  - [x] Cycle still performs claim, observe, save observation/error logic.
  - [x] Poller output/logs expose only active cycle counters (`claimed`, `updated`, `failed`) without `registered`.

### FR-003 - Backfill missing rows for existing issued allocations

- Description:
  - Existing issued allocations missing tracking rows must be inserted by migration.
- Acceptance criteria:
  - [x] Migration inserts missing rows using `INSERT ... SELECT ... ON CONFLICT DO NOTHING`.
  - [x] Backfill only includes `allocation_status='issued'` rows with non-null network/address.
  - [x] Backfill preserves source `issued_at` and sets initial status `watching`.

### FR-004 - Keep chain/network decoupled contract

- Description:
  - Registration API should not re-couple to Bitcoin-only types.
- Acceptance criteria:
  - [x] Registration source fields (`chain`, `network`, `address`) are copied from allocation row.
  - [x] No new Bitcoin-only type appears in application port signatures.

### FR-005 - Remove dead poller configuration surfaces

- Description:
  - Remove poller configuration/API fields that no longer affect behavior after write-through registration.
- Acceptance criteria:
  - [x] `POLL_REQUIRED_CONFIRMATIONS` is no longer parsed/used by `cmd/poller`.
  - [x] Poller compose overrides no longer define `POLL_REQUIRED_CONFIRMATIONS`.
  - [x] Polling DTOs no longer expose unused `DefaultRequiredConfirmations` and `RegisteredCount`.

### FR-006 - Network-specific required confirmations via env

- Description:
  - Allocation issue flow must support separate required-confirmations defaults for `bitcoin/mainnet` and `bitcoin/testnet4` through explicit env configuration.
- Acceptance criteria:
  - [x] `AllocatePaymentAddressUseCase` receives network-scoped default confirmations from DI/config instead of hardcoded constant only.
  - [x] `BITCOIN_MAINNET_REQUIRED_CONFIRMATIONS` and `BITCOIN_TESTNET4_REQUIRED_CONFIRMATIONS` are parsed and validated as positive integers.
  - [x] `compose.bitcoin.mainnet.yaml` and `compose.bitcoin.testnet4.yaml` define separate env keys for the two networks.
  - [x] When env is missing, each network falls back to `1`.

## Non-functional requirements

- Performance (NFR-001):
  - Poller cycle must avoid allocation-table scan query path.
- Availability/Reliability (NFR-002):
  - Registration and allocation issue remain atomic within one transaction.
- Observability (NFR-005):
  - Poller cycle summary logging reports active counters (`claimed`, `updated`, `failed`) only.
- Maintainability (NFR-006):
  - Keep use case orchestration in application layer and SQL in Postgres adapter.

## Dependencies and integrations

- External systems:
  - PostgreSQL.
- Internal services:
  - Allocation use case and poller cycle use case.
