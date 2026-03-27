---
doc: 00_problem
spec_date: 2026-03-27
slug: application-inbound-error-mapping
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-application-error-boundaries
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background: `internal/application/usecases` 已把 shared application error 與 shared outbound contract error 分開，但 usecase 本身仍會把 unexpected outbound/adapter/private error 原封不動回給 inbound adapter。
- Users or stakeholders: 維護 usecase contract 的開發者，與依賴這些 usecase 的 HTTP / scheduler inbound adapter。
- Why now: 使用者明確要求 application layer 對外只暴露 inbound error，不能再讓 adapter/private error 穿過 usecase 邊界。

## Constraints (optional)

- Technical constraints: 不改 HTTP response schema、不新增 logging framework、不改 outbound port shape。
- Timeline/cost constraints: Quick mode；聚焦在 usecase boundary 與對應測試。
- Compliance/security constraints: 對外仍維持 generic internal failure response，不放大 internal error detail。

## Problem statement

- Current pain: 多個 usecase 在 unexpected 失敗時直接回傳 adapter/private error 或 `outport.Err...`，代表 inbound adapter 雖然通常會 genericize response，但 application contract 本身仍不夠乾淨。
- Evidence or examples:
  - `GenerateAddressUseCase` 會直接回傳 policy reader / deriver error。
  - `AllocatePaymentAddressUseCase` 會直接回傳 replay lookup、policy reader、issued address derivation、transaction persistence error。
  - `GetPaymentAddressStatusUseCase` 目前會直接回傳 finder error。
  - `RunReceiptPollingCycleUseCase` 與 `RunReceiptWebhookDispatchCycleUseCase` 會把 transaction/store/notifier failure 原封不動往外拋。

## Goals

- G1: 讓 `internal/application/usecases` 對外只回傳 `inport.Err...` 或 `nil`。
- G2: 保留既有 business/config/validation mapping，不把所有失敗都壓成同一種錯誤。
- G3: 將 unexpected dependency failure 與 unexpected internal consistency failure 收斂成明確 inbound contract error。
- G4: 更新測試，讓 repo 對這條 contract 有回歸保護。

## Non-goals (out of scope)

- NG1: 不把 adapter package 內所有 `errors.New(...)` 一次清空。
- NG2: 不改 inbound adapter 的 HTTP status code 或 worker response schema。
- NG3: 不重做 outbound port hierarchy 或新增 logging/telemetry 機制。

## Assumptions

- A1: 已知的 usecase-visible business/config/validation error 仍應維持現有 `inport.Err...` 映射。
- A2: outbound port contract error 與 adapter/private error 不該再直接穿過 usecase 邊界。

## Open questions

- Q1: 無；範圍已足夠明確。
- Q2:

## Success metrics

- Metric: `internal/application/usecases` production code 不再直接把 unexpected non-inbound error 回給 caller。
- Target: usecase 單元測試覆蓋 generic inbound error mapping，`go test ./internal/application/usecases ./internal/adapters/inbound/...` 與 spec lint 維持綠燈。
