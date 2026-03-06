---
doc: 03_tasks
spec_date: 2026-03-06
slug: receipt-status-change-notification
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-06-receipt-polling-expiration-guard
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
  - 需要新 migration、新 outbound port/repository，並修改 polling 交易流程，屬跨層 async 事件基礎建設。

## Tasks (ordered)

1. T-001 - Add notification outbox schema

- Scope:
  - 新增 migration 建立 `payment_receipt_status_notifications` table 與索引。
- Linked requirements: FR-003, NFR-003, NFR-004
- Validation:
  - [x] 新增 `000006_receipt_status_notifications.up.sql` / `.down.sql`。

1. T-002 - Add notification repository port and tx wiring

- Scope:
  - 新增 `PaymentReceiptStatusNotificationRepository` port 與 enqueue input contract。
  - 擴充 `TxRepositories` 與 postgres `NewTxRepositories` wiring。
- Linked requirements: FR-002, FR-003, NFR-002
- Validation:
  - [x] `go test ./internal/application/ports/out ./internal/adapters/outbound/persistence/postgres -count=1`

1. T-003 - Implement postgres notification repository

- Scope:
  - 實作 `EnqueueStatusChanged` SQL，透過 `payment_address_id` 取 `customer_reference` 寫入 pending 事件。
- Linked requirements: FR-003, NFR-002
- Validation:
  - [x] `go test ./internal/adapters/outbound/persistence/postgres -count=1`

1. T-004 - Enqueue status-changed events in polling use case

- Scope:
  - observation/expiry 產生狀態轉換時，在同 tx 內執行 SaveObservation + EnqueueStatusChanged。
  - 狀態未變或 polling error 不 enqueue。
- Linked requirements: FR-001, FR-002, FR-004, NFR-001
- Validation:
  - [x] `go test ./internal/application/use_cases -count=1`

1. T-005 - Final validation and spec sync

- Scope:
  - 跑 short tests、precommit、spec-lint，更新 spec 完成狀態。
- Linked requirements: FR-001, FR-002, FR-003, FR-004, NFR-001, NFR-002, NFR-003, NFR-004
- Validation:
  - [x] `go test ./... -short -count=1`
  - [x] `bash scripts/precommit-run.sh`
  - [x] `SPEC_DIR="specs/2026-03-06-receipt-status-change-notification" bash scripts/spec-lint.sh`

## Traceability

- FR-001 -> T-004, T-005
- FR-002 -> T-002, T-004, T-005
- FR-003 -> T-001, T-002, T-003, T-005
- FR-004 -> T-004, T-005
- NFR-001 -> T-004, T-005
- NFR-002 -> T-002, T-003, T-005
- NFR-003 -> T-001, T-005
- NFR-004 -> T-001, T-005
