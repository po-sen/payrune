---
doc: 00_problem
spec_date: 2026-03-28
slug: process-error-reason-cleanup
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-application-error-boundaries
  - 2026-03-27-domain-error-contracts
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Problem & Goals

## Context

- Background: 前一輪先把 raw `err.Error()` 換成 controlled string constant，但使用者指出這仍只是把字串抽出來，還沒有把 reason 當成 domain concept 建模。
- Users or stakeholders: 維護 payment status polling、webhook delivery、以及對外 status response 的開發者與 reviewer。
- Why now: 使用者剛確認下一步要先收這塊，避免 adapter/private wording 繼續留在 domain/process state。

## Constraints (optional)

- Technical constraints: 不做 DB migration；資料庫仍可存字串，但 application/domain 內部要改成 typed reason code。
- Timeline/cost constraints: Quick mode，集中重構 receipt tracking 與 webhook delivery 這兩條 reason flow。
- Compliance/security constraints: 不增加新的 outward error detail。

## Problem statement

- Current pain: 即使不再直接寫 `err.Error()`，現在的 reason 仍只是 usecase-local string constant，不是 domain-owned type；這讓 process state 的語意仍然留在 application layer。
- Evidence or examples:
  - `internal/application/usecases/run_receipt_polling_cycle_use_case.go`
  - `internal/application/usecases/run_receipt_webhook_dispatch_cycle_use_case.go`
  - `internal/application/usecases/get_payment_address_status_use_case.go` 目前仍直接把 persisted string reason 回 DTO
  - `internal/domain/entities/payment_receipt_tracking.go` 與 `internal/domain/policies/payment_receipt_status_notification_delivery.go` 都還是自由字串

## Goals

- G1: 把 receipt tracking 與 webhook delivery 的 failure reason 收成 domain typed code，而不是自由字串
- G2: 讓 usecase 只負責 `lower-level error -> domain reason code` mapping
- G3: read side 對外回 public text，而不是把 persisted code 或歷史 raw text 直接穿出去

## Non-goals (out of scope)

- NG1: 不處理 `PaymentAddressAllocation.FailureReason` 這條線
- NG2: 不處理 `PaymentAddressAllocation` 這條 derivation failure path

## Assumptions

- A1: 這些 persisted reason 屬於 domain/process state，而不是 logging/debug channel
- A2: 舊資料庫裡可能已經存在 legacy raw text，需要在 read-side parse 時安全降級

## Open questions

- Q1: 無
- Q2:

## Success metrics

- Metric: receipt process reason 是否由 domain typed code 擁有，而不是 usecase-local string
- Target: polling 與 webhook dispatch 的 domain model / policy 改吃 typed reason，usecase 不再傳自由字串；status read path 對外回 public text
