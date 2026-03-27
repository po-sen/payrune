---
doc: 00_problem
spec_date: 2026-03-27
slug: domain-error-contracts
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-application-inbound-error-mapping
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background: `internal/domain` 仍有多處匿名 `errors.New(...)`，只有少數 domain rule 已使用具名 sentinel，例如 [address_issuance_policy.go](/Users/posen/Desktop/payrune/internal/domain/entities/address_issuance_policy.go)。
- Users or stakeholders: 維護 clean architecture 邊界的開發者，與需要將 domain rule 轉譯為 application error 的 usecase 作者。
- Why now: 使用者希望先把 domain error 抽出來，避免 usecase 未來只能靠錯誤字串處理 domain invariant。

## Constraints (optional)

- Technical constraints: 不重做 package 結構；維持既有 domain package 邊界。
- Timeline/cost constraints: Quick mode；以 domain error contract 抽取與測試為主。
- Compliance/security constraints: 不放寬既有 validation；只把 error contract 明文化。

## Problem statement

- Current pain: `internal/domain/entities`, `internal/domain/valueobjects`, `internal/domain/events`, `internal/domain/policies` 仍有許多匿名 `errors.New(...)`。這讓 application layer 很難穩定 `errors.Is(...)`，也容易在 refactor 時不小心靠字串耦合。
- Evidence or examples:
  - [payment_receipt_tracking.go](/Users/posen/Desktop/payrune/internal/domain/entities/payment_receipt_tracking.go) 的 constructor / transitions 幾乎全是匿名錯誤。
  - [payment_address_allocation.go](/Users/posen/Desktop/payrune/internal/domain/entities/payment_address_allocation.go) 與 [payment_receipt_observation.go](/Users/posen/Desktop/payrune/internal/domain/valueobjects/payment_receipt_observation.go) 也有相同情況。
  - application usecase 目前只有部分 domain rule 能靠 stable sentinel 做精準 mapping。

## Goals

- G1: 將 domain package 中有跨層判斷價值的匿名錯誤改為 stable sentinel errors。
- G2: 讓 application/usecase 可以透過 `errors.Is(...)` 處理 domain rule，而不是依賴字串。
- G3: 維持 concrete naming，不引入一個過度泛化的 global error registry。
- G4: 補回歸測試，鎖定新的 domain error contract。

## Non-goals (out of scope)

- NG1: 不處理 application / adapter / bootstrap error contract。
- NG2: 不調整 HTTP response 或 scheduler response schema。
- NG3: 不為每一個僅限 package 內部、沒有跨層價值的 one-off error 強行抽象成複雜 hierarchy。

## Assumptions

- A1: 以 package 為單位定義 domain sentinel error，比起單一全域 error 檔更符合 repo 的 concrete naming 風格。
- A2: 既有已經具名的 domain errors，例如 `ErrAddressPolicyNotEnabled`，應保持不變。

## Open questions

- Q1: 無；範圍已足夠明確。
- Q2:

## Success metrics

- Metric: `internal/domain` 中原本匿名且具跨層價值的 invariant error 改為 stable sentinel。
- Target: `go test ./internal/domain/... ./internal/application/usecases ./...` 與 `SPEC_DIR="specs/2026-03-27-domain-error-contracts" bash scripts/spec-lint.sh` 通過。
