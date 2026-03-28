---
doc: 00_problem
spec_date: 2026-03-28
slug: allocate-usecase-readability
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-26-allocate-usecase-decomposition
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Problem & Goals

## Context

- Background: `internal/application/usecases/allocate_payment_address_use_case.go` 已經比之前乾淨，但 transaction flow 仍有 side-effect state 和錯誤回傳路徑交錯，尤其是 derivation failure 會透過外部變數再補回真正錯誤。
- Users or stakeholders: 維護 allocate payment address flow 與 review application usecase 的開發者。
- Why now: 使用者明確要求把這支收成更容易順著讀的 production code，而不是只做檔案切分。

## Constraints (optional)

- Technical constraints: 不改 usecase outward contract、不改 transaction boundary、不改 store/domain behavior。
- Timeline/cost constraints: Quick mode，專注在 `allocate_payment_address_use_case.go` 的順讀改善。
- Compliance/security constraints: 不引入新的 outward error detail。

## Problem statement

- Current pain: `issueAllocation(...)` 目前有 `derivationFailureErr` 這種 transaction 外部 side-effect state，讓 callback 回傳錯誤和最終 outward error 不是同一條線；讀者得來回比對 callback 內外才知道最後會回什麼。
- Evidence or examples:
  - `internal/application/usecases/allocate_payment_address_use_case.go`
  - derivation failure flow currently sets a side variable and later overrides the outward result

## Goals

- G1: 讓 `issueAllocation(...)` 的交易主體更接近單一路徑順讀。
- G2: 移除不必要的 side-effect error state，讓 derivation failure path 和最終回傳錯誤對齊。
- G3: 保持 allocation/idempotency/receipt tracking 的既有 transaction behavior 不變。

## Non-goals (out of scope)

- NG1: 不重新設計 `UnitOfWork`。
- NG2: 不重寫整支 usecase 成多檔或多個 public usecase。

## Assumptions

- A1: 這輪應該優先收順交易主體，而不是增加更多新的 helper/type。
- A2: 若 helper 需要保留，必須是為了讓主流程更直，而不是單純搬動程式碼。

## Open questions

- Q1: 無
- Q2:

## Success metrics

- Metric: `issueAllocation(...)` 的錯誤路徑是否不再依賴 transaction 外部 side-effect state
- Target: derivation failure flow 改成直接沿交易主體回傳最終 outward error，且 existing tests/full suite 維持通過
