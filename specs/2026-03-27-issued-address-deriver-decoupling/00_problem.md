---
doc: 00_problem
spec_date: 2026-03-27
slug: issued-address-deriver-decoupling
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-application-error-boundaries
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null # set to 02_design.md in Full mode
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background: `internal/adapters/outbound/blockchain/issued_payment_address_deriver.go` 目前雖然已移除對 concrete ethereum type 的直接 import，但 `blockchain` package 仍然知道 `ethereum/create2` 的特殊流程。使用者明確指出 `blockchain` 應保持中立，不應承載 chain-specific issued derivation 細節。
- Users or stakeholders: 維護 outbound adapter 邊界的人，以及後續閱讀 bootstrap / adapter wiring 的開發者。
- Why now: error ownership cleanup 已收尾，使用者要求接著處理這個 concrete coupling，讓 outbound adapter 組裝更乾淨。

## Constraints (optional)

- Technical constraints: 這輪只做最小 decoupling；不能把單一實作硬抽成通用 framework，也不能改變地址導出行為。
- Timeline/cost constraints: Quick mode；限於 `issued_payment_address_deriver`、其 constructor call sites、與相關 tests。
- Compliance/security constraints: 不改變 create2 salt derivation 演算法、不新增 runtime config。

## Problem statement

- Current pain: `blockchain` package 目前仍持有 `ethereum/create2` 的 chain-specific issued derivation 細節，違反它應該保持中立、只做 dispatch 的定位。
- Evidence or examples:
  - `blockchain/issued_payment_address_deriver.go` 仍直接分支 `ethereum/create2` 邏輯。
  - `deriveCreate2SaltForIDFn` 只是把 concrete coupling 換成 function dependency，但沒有讓 `blockchain` 回到中立。
  - 真正的 create2 salt / relative reference 邏輯應該屬於 `ethereum` adapter，而不是 `blockchain`。

## Goals

- G1: 讓 `blockchain` package 只負責 multi-chain dispatch，不再知道 `ethereum/create2` issued derivation 細節。
- G2: 把 bitcoin / ethereum issued address derivation 拆回各自 package。
- G3: 保持目前 `create2` relative reference / salt derivation 行為完全不變。
- G4: 不引入泛用 registry / framework，維持 concrete、可讀的組裝方式。

## Non-goals (out of scope)

- NG1: 不改變 `Create2SaltDeriver` 的演算法或 env/config wiring。
- NG2: 不把所有 outbound adapter dependency 都抽成大型 interface catalog。
- NG3: 不改 application port shape。

## Assumptions

- A1: 最合理的修法是 chain-specific issued deriver 實作加上一個 `blockchain` dispatch owner，而不是留一個中間 function hook。
- A2: `IssuedPaymentAddressDeriver` 仍然是 outbound adapter 內部協作者，不需要因此重畫 application layer boundary。

## Open questions

- Q1: 無；方向已足夠明確。
- Q2:

## Success metrics

- Metric: `blockchain` package 不再包含 `ethereum/create2` issued derivation knowledge。
- Target: bitcoin / ethereum issued derivation tests 與 bootstrap wiring 測試維持通過。
