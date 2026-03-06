---
doc: 02_design
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

# Technical Design

## High-level approach

- 在 `RunReceiptPollingCycleUseCase` 判斷狀態是否轉換。
- 狀態轉換成立時，除既有 `SaveObservation` 外，同 transaction enqueue notification event。
- 新增 `PaymentReceiptStatusNotificationRepository` outbound port 與 postgres adapter。
- 本期不做 dispatch/notifier；事件先落為 `pending`。

## Key flows

- Flow 1 (observation transition):
  - claim tracking -> observe -> apply observation -> status changed -> tx: `SaveObservation` + `EnqueueStatusChanged` -> commit。
- Flow 2 (expiry transition):
  - claim tracking -> check expired -> mark expired -> status changed -> tx: `SaveObservation` + `EnqueueStatusChanged` -> commit。
- Flow 3 (no transition):
  - claim tracking -> status unchanged 或 polling error -> 不 enqueue notification。

## Data model

- New table: `payment_receipt_status_notifications`
  - `id BIGSERIAL PRIMARY KEY`
  - `payment_address_id BIGINT NOT NULL REFERENCES address_policy_allocations(id) ON DELETE CASCADE`
  - `customer_reference TEXT`
  - `previous_status TEXT NOT NULL`
  - `current_status TEXT NOT NULL`
  - `observed_total_minor BIGINT NOT NULL CHECK (observed_total_minor >= 0)`
  - `confirmed_total_minor BIGINT NOT NULL CHECK (confirmed_total_minor >= 0)`
  - `unconfirmed_total_minor BIGINT NOT NULL CHECK (unconfirmed_total_minor >= 0)`
  - `conflict_total_minor BIGINT NOT NULL CHECK (conflict_total_minor >= 0)`
  - `status_changed_at TIMESTAMPTZ NOT NULL`
  - `delivery_status TEXT NOT NULL CHECK (delivery_status IN ('pending', 'sent', 'failed'))`
  - `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Indexes:
  - `idx_payment_receipt_status_notifications_delivery_created` on `(delivery_status, created_at ASC)`.
  - `idx_payment_receipt_status_notifications_address_created` on `(payment_address_id, created_at DESC)`.

## Contracts

- New outbound port:
  - `PaymentReceiptStatusNotificationRepository`
    - `EnqueueStatusChanged(ctx, input) error`
- UnitOfWork tx repositories 新增:
  - `PaymentReceiptStatusNotification`
- Use case orchestration:
  - `RunReceiptPollingCycleUseCase` 僅在狀態轉換時呼叫 enqueue。

## Failure modes and resiliency

- 若 enqueue 失敗，transaction rollback，避免狀態與事件不一致。
- 若 customer_reference 為空，仍允許入庫（nullable），不阻斷 polling 主流程。

## Observability

- 本期以 DB 查詢為主:
  - pending backlog count
  - per-status transition event count
- 不新增外部通知 metrics（下一期補）。

## Security

- 事件表不存任何通知通道憑證。

## Configuration contract

- None for phase 1.
