---
doc: 00_problem
spec_date: 2026-04-03
slug: notification-delivery-boundary
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-02-domain-model-boundary-cleanup
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background:
  - `PaymentReceiptNotificationDeliveryStatus` 與 `PaymentReceiptNotificationDeliveryFailureReason` 目前放在 `internal/domain/valueobjects`。
  - 這兩個型別實際上描述的是 receipt webhook / outbox delivery workflow，而不是 payment receipt domain fact。
  - 它們雖然會被寫進 DB，但「需要持久化」不代表「屬於 domain」。
- Users or stakeholders:
  - 維護 receipt webhook dispatch、outbox store、以及 domain/application boundary 的開發者。
  - 依 `AGENTS.md` 做 review 的 agent / reviewer。
- Why now:
  - 使用者已明確指出這兩個型別看起來不屬於 domain，希望把 boundary 再收乾淨一輪。

## Constraints (optional)

- Technical constraints:
  - 維持既有 DB schema 與 persisted string values，不做 migration。
  - 不擴成另一輪大範圍 domain cleanup；本 spec 只處理 notification delivery workflow types 的 ownership。
  - 不新增新的 top-level architecture bucket。
- Timeline/cost constraints:
  - 本輪要用小而準的 refactor 完成，不引入新的抽象層。
- Compliance/security constraints:
  - webhook delivery retry / sent / failed 行為不可改變。

## Problem statement

- Current pain:
  - `internal/domain/valueobjects` 目前仍混入 notification delivery workflow status/reason，讓 domain bucket 不夠純。
  - 這會讓 reviewer 誤以為 `pending/sent/failed` 是 payment domain 狀態，而不是 outbox delivery 狀態。
  - 相關 parse / canonicalization 目前也跟著留在 domain，與實際使用場景脫節。
- Evidence or examples:
  - [`internal/domain/valueobjects/payment_receipt_notification_delivery_status.go`](/Users/posen/Desktop/payrune/internal/domain/valueobjects/payment_receipt_notification_delivery_status.go)
  - [`internal/domain/valueobjects/payment_receipt_notification_delivery_failure_reason.go`](/Users/posen/Desktop/payrune/internal/domain/valueobjects/payment_receipt_notification_delivery_failure_reason.go)
  - 真正主要使用者是 [`internal/application/outbox/payment_receipt_status_notification_delivery_result.go`](/Users/posen/Desktop/payrune/internal/application/outbox/payment_receipt_status_notification_delivery_result.go) 與 persistence outbox stores。

## Goals

- G1:
  - 將 notification delivery status / failure reason 從 domain 移到更合適的 application/outbox boundary。
- G2:
  - 保持 persisted string values、DB schema、與 webhook dispatch 行為不變。
- G3:
  - 讓 `internal/domain` 只保留 payment receipt 本身的 domain 狀態，不再混入 delivery workflow state。

## Non-goals (out of scope)

- NG1:
  - 不改 `payment_receipt_status_notification_outbox` 的 schema 或資料值。
- NG2:
  - 不重寫 webhook dispatch use case 或 outbox store 流程。

## Assumptions

- A1:
  - delivery `pending/sent/failed` 與 `delivery_failed` 屬於 application/outbox workflow state，即使它們會被存進 DB。
- A2:
  - 這兩個型別搬移後，application/outbox 會是最自然的 owning boundary。

## Open questions

- Q1:
  - 無；本 spec 只做 ownership 更正，不改語意與值集合。

## Success metrics

- Metric:
  - Domain purity.
- Target:
  - `internal/domain` 不再匯出 notification delivery workflow status / failure reason type。
- Metric:
  - Regression safety.
- Target:
  - `go test ./internal/domain/... ./internal/application/... ./internal/adapters/outbound/persistence/...` 通過，且 persisted values 不變。
