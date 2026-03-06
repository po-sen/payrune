---
doc: 03_tasks
spec_date: 2026-03-06
slug: receipt-webhook-delivery
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-06-receipt-status-change-notification
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
- Rationale:
  - Includes schema migration, new worker runtime, new outbound HTTP integration, and retry/lease behavior.

## Tasks (ordered)

1. T-001 - Extend notification outbox schema for delivery lifecycle

- Scope:
  - Add migration for attempts, scheduling, lease, error, delivered timestamp, and pending-delivery indexes.
- Linked requirements: FR-001, FR-002, NFR-001, NFR-004
- Validation:
  - [x] Added `000007_receipt_status_notification_delivery` migration and validated spec lint.

1. T-002 - Add delivery domain/application contracts

- Scope:
  - Add delivery status value object/entity, dispatch DTOs, inbound port, notifier port, and repository methods.
- Linked requirements: FR-002, FR-004, FR-005, NFR-003
- Validation:
  - [x] `go test ./internal/domain/... ./internal/application/... -count=1`

1. T-003 - Implement postgres notification dispatch persistence

- Scope:
  - Implement claim and result update SQL paths with lease semantics.
- Linked requirements: FR-002, FR-005, NFR-001, NFR-004
- Validation:
  - [x] `go test ./internal/adapters/outbound/persistence/postgres -count=1`

1. T-004 - Implement fixed-endpoint webhook notifier adapter

- Scope:
  - Add HTTP webhook adapter with payload signing, timeout, and `2xx` success contract.
- Linked requirements: FR-003, FR-004, NFR-002, NFR-003
- Validation:
  - [x] `go test ./internal/adapters/outbound/webhook -count=1`

1. T-005 - Implement dispatch use case and runtime wiring

- Scope:
  - Add dedicated dispatcher use case, bootstrap, DI, env parsing, command, and Dockerfile.
- Linked requirements: FR-005, FR-006, NFR-001, NFR-004
- Validation:
  - [x] `go test ./internal/application/use_cases ./internal/infrastructure/di ./cmd/... ./internal/bootstrap -count=1`

1. T-006 - Final validation and spec sync

- Scope:
  - Move dispatcher service into `compose.yaml`, require webhook env via Compose interpolation, add committed fake test env values in `compose.test.env`, and keep fake webhook receiver wiring in `compose.test.yaml`.
  - Run short tests, precommit, and spec lint; update spec to final behavior.
- Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, NFR-001, NFR-002, NFR-003, NFR-004, NFR-005
- Validation:
  - [x] `go test ./... -short -count=1`
  - [x] `bash scripts/precommit-run.sh`
  - [x] `SPEC_DIR="specs/2026-03-06-receipt-webhook-delivery" bash scripts/spec-lint.sh`
  - [x] `docker compose --env-file deployments/compose/compose.test.env -f deployments/compose/compose.yaml -f deployments/compose/compose.test.yaml config`

## Traceability

- FR-001 -> T-001, T-006
- FR-002 -> T-001, T-002, T-003, T-006
- FR-003 -> T-004, T-006
- FR-004 -> T-002, T-004, T-006
- FR-005 -> T-002, T-003, T-005, T-006
- FR-006 -> T-005, T-006
- FR-007 -> T-006
- NFR-001 -> T-001, T-003, T-005, T-006
- NFR-002 -> T-004, T-006
- NFR-003 -> T-002, T-004, T-006
- NFR-004 -> T-001, T-003, T-005, T-006
- NFR-005 -> T-006
