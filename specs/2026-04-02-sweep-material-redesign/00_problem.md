---
doc: 00_problem
spec_date: 2026-04-02
slug: sweep-material-redesign
mode: Full
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-02-domain-model-boundary-cleanup
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
  - `PaymentAddressAllocation` 原本直接持有 `SweepMaterialJSON string`，讓 operator recovery document 的序列化表示進到 domain entity。
  - 前一輪又曾嘗試把這份資料改成新的 application mini-model 與額外 wiring，結果把 `internal/application/ports/outbound`、`UnitOfWork`、bootstrap 都拉複雜了。
  - 這個 repo 真正需要的其實很小: domain 不要背 JSON，deriver 直接產出 JSON，use case 原樣傳給既有 store，store 只寫欄位。
- Users or stakeholders:
  - 維護 address issuance flow、allocation persistence、以及 operator recovery payload 的開發者。
  - 依 `AGENTS.md` 做 review 的 agent / reviewer。
- Why now:
  - 目前這條邊界仍不乾淨，且使用者已明確指出先前的 port-model 寫法過度設計、難以理解。

## Constraints (optional)

- Technical constraints:
  - 維持既有 Clean Architecture + Hexagonal 邊界。
  - 不新增新的 top-level architecture 或 generic framework。
  - 不更改 `address_policy_allocations.sweep_material_json` schema 與現有 operator-facing JSON 格式。
  - `internal/application/ports/**` 必須維持 plain contract 風格，不新增 constructor/validation-heavy model API。
- Timeline/cost constraints:
  - 本輪只重設 `SweepMaterial` 相關邊界，不順手擴成另一輪 domain 大重構。
- Compliance/security constraints:
  - operator recovery payload 不能遺失既有必要欄位。
  - `Complete` 失敗時不得留下半套 issued state。

## Problem statement

- Current pain:
  - domain entity 持有 JSON representation，不符合 domain 只持有 business state 的原則。
  - 若為了搬走這個欄位而新增新的 application model、builder dispatcher、第二套 `UnitOfWork` 或 bootstrap helper，infra 會被一個新欄位拖到過度設計。
  - 現有設計需要明確回答: `sweep_material_json` 不是 domain state，它只是 deriver 產出的 operator document，經過 use case 搬運後由 store 寫入。
- Evidence or examples:
  - [`internal/domain/entities/payment_address_allocation.go`](/Users/posen/Desktop/payrune/internal/domain/entities/payment_address_allocation.go) 曾直接保存 `SweepMaterialJSON string`。
  - 先前嘗試引入額外 builder / wiring 後，generic `UnitOfWork` 與 bootstrap 開始知道 allocation-specific collaborator，違反 repo 對簡潔 wiring 的要求。
  - bitcoin / ethereum deriver 本來就最接近產出這份 operator document 所需的鏈別資料。

## Goals

- G1:
  - 從 domain entity 拿掉 `SweepMaterialJSON string`。
- G2:
  - 不再引入新的 `SweepMaterial` application mini-model；`internal/application/ports/outbound` 維持 boring contract 風格。
- G3:
  - 將 `sweep_material_json` 保留為 operator/persistence document，由 chain-specific deriver adapter 組裝。
- G4:
  - 讓 `PaymentAddressAllocationStore.Complete` 只接收 store 真正需要寫入的資料: issued allocation、JSON、issued time。
- G5:
  - 保持既有 DB column 與 JSON payload shape 不變，避免 migration 與 operator breakage。
- G6:
  - 用 targeted tests 證明新的 ownership 比目前更清楚，且不會再把 port package 寫成 overdesigned model layer。
- G7:
  - 不污染 generic `UnitOfWork` constructor，也不新增只為組 sweep-material builder 存在的 `internal/bootstrap` helper 檔。

## Non-goals (out of scope)

- NG1:
  - 不改 `sweep_material_json` 的 DB schema。
- NG2:
  - 不重新設計整個 issuance flow 或 multi-chain deriver abstraction。
- NG3:
  - 不新增 `internal/application/contracts`、`models`、`services` 等新 bucket 來安置這個問題。
- NG4:
  - 不把 operator payload 提升成 domain value object，也不為它發明新的 application mini-model。
- NG5:
  - 不為了存一個新欄位而新增第二套 UnitOfWork 流程或多餘的 adapter-private wiring tree。

## Assumptions

- A1:
  - 組裝 `sweep_material_json` 所需的資料，由 chain-specific deriver 在 issuance 當下就能取得。
- A2:
  - `sweep_material_json` 的主要消費者是 persistence / operator tooling，而不是 domain 或一般 application use case。
- A3:
  - `sweep_material_json` 雖然是技術 payload，但由 chain-specific issued deriver 直接產出，比再多切一層 persistence-side builder 更符合這個 repo 對簡潔度的要求。

## Open questions

- Q1:
  - 無。這份 spec 採最小設計，不再讓 `Complete` input 背 `IssuanceRefKind` / `IssuanceRef` / `IssuancePolicy`。

## Success metrics

- Metric:
  - Domain / application boundary cleanliness.
- Target:
  - `rg -n "SweepMaterialJSON" internal/domain` 不再命中 runtime code，且 `internal/application/ports/outbound` 只在既有 deriver/store contract file 內出現 plain string field。
- Metric:
  - Port simplicity.
- Target:
  - `internal/application/ports/outbound` 不新增 standalone `sweep_material.go`，也不新增 `New...` / `Validate...` 類的 `SweepMaterial` helper API；只在既有 deriver/store contract 增加 plain string field。
- Metric:
  - Regression safety.
- Target:
  - `go list ./...`、`go test ./...`、`SPEC_DIR="specs/2026-04-02-sweep-material-redesign" bash scripts/spec-lint.sh` 通過。
