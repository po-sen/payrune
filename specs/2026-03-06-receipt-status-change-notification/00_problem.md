---
doc: 00_problem
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

# Problem & Goals

## Context

- Background:
  - Receipt polling 已可更新 `payment_receipt_trackings`，但目前沒有把「狀態改變」可靠地落成通知事件。
- Users or stakeholders:
  - 發起交易申請的商戶系統、平台後端與維運團隊。
- Why now:
  - 先把狀態變更事件可靠入庫，才能在下一期安全接 webhook/其他通道。

## Problem statement

- Current pain:
  - 雖然 tracking 狀態會更新，但沒有事件層可供通知系統消費。
- Evidence or examples:
  - Poller 目前只做 `SaveObservation`/`SavePollingError`，沒有 outbox enqueue。
  - 若未來直接在 poller 內同步外呼，會把外部通道不穩定性耦合到核心輪詢。

## Goals

- G1:
  - 每次 tracking `receipt_status` 發生轉換時，建立一筆 notification 事件。
- G2:
  - tracking 狀態更新與 notification 事件建立必須在同一 transaction，避免漏事件。
- G3:
  - 事件資料需包含申請方識別資訊（至少 `customer_reference`）與轉換上下文。

## Non-goals (out of scope)

- NG1:
  - 本期不實作 webhook/dispatcher 實際投遞流程。
- NG2:
  - 本期不變更 receipt 狀態判斷規則。

## Assumptions

- A1:
  - `address_policy_allocations.customer_reference` 可作為申請方關聯識別。
- A2:
  - 本期事件預設 `delivery_status = 'pending'`，由下一期通知派送流程消費。

## Success metrics

- Metric:
  - 狀態變更事件完整性。
- Target:
  - 每次 `previous_status != current_status` 都會落一筆 outbox 事件。
- Metric:
  - 交易一致性。
- Target:
  - tracking 更新成功時必有對應事件；事件寫入失敗時 tracking 更新不提交。
- Metric:
  - 事件可追溯性。
- Target:
  - 事件列可查到 `payment_address_id`, `customer_reference`, `from/to status`, `status_changed_at` 與金額快照。
