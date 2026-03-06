---
doc: 01_requirements
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

# Requirements

## Functional requirements

### FR-001 - Detect status transition in polling

- Description:
  - Polling cycle 必須在 observation/expiry 判斷後，比較 `previous_status` 與 `current_status`。
- Acceptance criteria:
  - [x] `previous_status != current_status` 時視為狀態轉換。
  - [x] 狀態未變時不得建立 notification 事件。

### FR-002 - Transactional enqueue with tracking update

- Description:
  - `SaveObservation` 與 `EnqueueStatusChanged` 必須在同一 UnitOfWork transaction。
- Acceptance criteria:
  - [x] enqueue 發生錯誤時，該筆 tracking 更新不提交。
  - [x] transaction 成功時，同時可見 tracking 新狀態與 pending 事件。

### FR-003 - Persist notification outbox event snapshot

- Description:
  - 新增 outbox table 儲存狀態轉換事件快照。
- Acceptance criteria:
  - [x] 事件包含 `payment_address_id`, `customer_reference`, `previous_status`, `current_status`, `status_changed_at`。
  - [x] 事件包含 `observed_total_minor`, `confirmed_total_minor`, `unconfirmed_total_minor`, `conflict_total_minor`。
  - [x] 事件初始 `delivery_status` 為 `pending`。

### FR-004 - Cover all status-change paths

- Description:
  - 本期需涵蓋 observation 路徑與 expiry 路徑的狀態轉換事件。
- Acceptance criteria:
  - [x] `watching -> paid_*` 等 observation 轉換會 enqueue。
  - [x] `* -> failed_expired` 轉換會 enqueue。
  - [x] polling error 且狀態未改變時不 enqueue。

## Non-functional requirements

- Reliability (NFR-001):
  - 已提交狀態轉換不得遺失事件（transactional outbox）。
- Maintainability (NFR-002):
  - 通知事件入庫透過獨立 outbound port/repository，避免 use case 直接耦合 SQL。
- Operability (NFR-003):
  - 事件表需可依 `delivery_status` 與時間查詢 backlog。
- Scalability (NFR-004):
  - 事件表需具備可支援後續 dispatcher claim 的索引基礎。

## Dependencies and integrations

- External systems:
  - 無（本期不做外部投遞）。
- Internal services:
  - `RunReceiptPollingCycleUseCase`
  - PostgreSQL persistence + UnitOfWork transaction
