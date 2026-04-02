---
doc: 00_problem
spec_date: 2026-04-02
slug: domain-model-boundary-cleanup
mode: Full
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-07-architecture-naming-refactor
  - 2026-03-24-architecture-conformance-refactor
  - 2026-03-28-allocation-failure-reason-typing
  - 2026-03-29-allocation-issuance-naming
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
  - repo 已經將核心邏輯集中在 `internal/domain`，但目前 `entities`、`policies`、`valueobjects` 內仍混入幾種不同語義: 真正的 business model、deployment catalog、outbox workflow transition helper、health response enum、以及 persistence backward-compatibility parsing。
  - 上一輪架構與命名整理已把很多責任拉回正確層，但 domain 內部的型別分類仍不夠乾淨，導致 package 名稱與實際責任開始脫鉤。
  - 上一輪 cleanup 後，runtime 還殘留一條 `entity -> policy` 依賴，代表 domain dependency direction 還沒有被硬性收斂。
- Users or stakeholders:
  - 維護 payrune domain/application/adapters 的開發者。
  - 未來要繼續做 domain refactor 或新增 payment flow 的人。
  - 依 `AGENTS.md` 工作的 agent / reviewer。
- Why now:
  - 使用者已直接指出 `internal/domain` 內有「看起來不太對勁」的分類。
  - 若不先把 domain category 邊界講清楚，後續 feature 很容易再把 deployment config、workflow row state、或 adapter compatibility logic 放回 domain。

## Constraints (optional)

- Technical constraints:
  - 維持既有 Clean Architecture + Hexagonal 邊界，不新增新的 top-level architecture。
  - 採用務實的簡潔架構整理，不把這輪擴成完整 DDD package strategy migration。
  - 維持目前 `internal/domain/{entities,events,valueobjects,policies}` 的 top-level 結構；本輪不新增 `domain/services`、`domain/enums` 等新 bucket。
  - 優先做 behavior-preserving cleanup；如無必要，不改 API contract、不改 schema。
  - 若型別仍屬於核心商業概念，可調整 package/責任，但不要為了純命名美化做大規模 churn。
- Timeline/cost constraints:
  - 本輪先定義並整理 domain model boundary 與 repo contract，不要求同時做 unrelated feature work。
- Compliance/security constraints:
  - 支付狀態轉移、webhook dispatch、address issuance 行為不可因分類重整而被弱化。

## Problem statement

- Current pain:
  - `internal/domain/entities` 內有些型別沒有 entity 應有的 identity + lifecycle + business behavior，只是在承載 metadata 或從別的 config 推導 `Enabled`。
  - `internal/domain/policies` 內有些型別其實是 outbox workflow result 或 pass-through wrapper，不是真正的 domain policy。
  - `internal/domain/valueobjects` 內有些 parser 承擔了 legacy DB text alias 與 unknown fallback，這比較像 adapter/persistence compatibility，而不是核心 VO 應該持有的 canonical rule。
  - `AGENTS.md` 目前對 entity / aggregate root / repository / store / DAO 的判準仍偏 generic，還不足以阻止這些錯位再發生。
  - `PaymentAddressAllocation` 的 issued transition 仍直接接受 `AddressIssuancePolicy`，使 entity 對 policy 產生 runtime import。
  - `AddressPolicyID` 與 address `Scheme` 仍在多個核心型別與 application port 中以裸字串流動，缺少 dedicated typed domain scalar。
  - `AddressPolicyID` 雖已被提升成 typed scalar，但 repo 內建 policy IDs 的集中定義若被拆成過度零碎的 sibling file，也會讓這個 VO 看起來被不必要地切散。
  - malformed `AddressPolicyID` 目前在部分 application/read path 仍會先被 `Normalize()` 吃成 zero value，再掉進 `not found`，讓 invalid input 與 unknown policy 混成同一個結果。
  - 在 `AddressScheme` 已進入 domain 後，`BitcoinAddressScheme` 仍留在 `internal/domain/valueobjects`，形成兩套重疊的 scheme truth。
  - `NetworkID` 已是 domain canonical scalar，但 `BitcoinNetwork` 仍以另一個 domain type 並存，讓 network 也出現雙軌模型。
- Evidence or examples:
  - `internal/domain/entities/address_policy.go` 只有 metadata 與 normalize，沒有 entity lifecycle。
  - `internal/domain/entities/address_issuance_policy.go` 會在 normalize 時把 `Enabled` 回填到內層 `AddressPolicy`，顯示狀態來源不在該 entity 自身。
  - `internal/domain/policies/payment_receipt_tracking_lifecycle.go` 幾乎只是包一層 entity method。
  - `internal/domain/policies/payment_receipt_status_notification_delivery.go` 的 result type 實際只被 webhook dispatch use case 與 outbox port 使用。
  - `internal/domain/valueobjects/*failure_reason.go` 目前吸收 legacy alias 與 unknown-text fallback。
  - `internal/domain/valueobjects/service_status.go` 只被 health check response 使用，並非付款領域概念。

## Goals

- G1:
  - 讓 `internal/domain` 內每個 exported type 都能清楚回答自己屬於 `entity`、`event`、`value object`、`policy`，或根本不該留在 domain。
- G2:
  - 把 list/query metadata、deployment catalog、workflow result、health response enum、persistence compatibility logic 從錯誤的 domain category 中移開。
- G3:
  - 保留真正穩定的 domain core: `PaymentAddressAllocation`、`PaymentReceiptTracking`、`PaymentReceiptStatusChanged` 與 canonical value objects。
- G4:
  - 讓 address-policy / issuance-policy 這組模型的責任重新清楚，避免 public metadata、issuance capability、deployment config 混成單一 entity bucket。
- G5:
  - 更新 `AGENTS.md`，將本 repo 對 `entity / aggregate root / event / value object / policy` 的判準與 review trigger 寫明，成為後續評審契約。
- G6:
  - 用 targeted test 與 compile checks 證明本輪整理不改變既有行為。
- G7:
  - 將 repo stance 定義為 pragmatic clean architecture：允許 domain-oriented modeling，但不要求 full DDD、feature-oriented package tree、或 repository-over-store 教條。
- G8:
  - 將 `AddressPolicyID` 與 address `Scheme` 明確建模成 typed domain scalar / value object，減少核心模型中的裸字串。
- G9:
  - 讓 domain 內只保留一套 canonical scheme model；Bitcoin-specific scheme routing 若仍需要，應下沉到 bitcoin adapter 內部。
- G10:
  - 讓 domain 內只保留一套 canonical network model；Bitcoin-specific network routing 若仍需要，應下沉到 bitcoin adapter 內部。
- G11:
  - 讓 `AddressPolicyID` 保持 open typed identifier，但同時補齊 repo 內建 built-in policy IDs 與 helper，避免 runtime code 持續散落裸字串。
- G12:
  - 讓 malformed `AddressPolicyID` 在 application 與 persistence 邊界都能被明確辨識為 invalid，而不是靜默降級成 `not found` 或 zero value。

## Non-goals (out of scope)

- NG1:
  - 不在本輪新增新的支付功能、鏈支援或外部整合。
- NG2:
  - 不為了追求理論完美而重寫整個 application / adapter 層。
- NG3:
  - 不以純 stylistic rename 為目標；只有在責任更清楚時才改名或搬 package。
- NG4:
  - 不主動引入新的 DB schema 或 migration，除非後續實作證明沒有 schema 無法達成。
- NG5:
  - 不在本輪新增 `internal/domain/services`、`internal/domain/enums` 或其他新的 domain top-level bucket。
- NG6:
  - 不把 `internal/domain` 或 `internal/application` 全面改成依業務概念切 package 的新目錄策略。

## Assumptions

- A1:
  - 可接受將部分現有 domain type 移到 application/adapters/bootstrap 邊界，只要行為不變且責任更清楚。
- A2:
  - `AddressPolicy` 類的 public list metadata 若只服務查詢與輸出，可視情況降到 application read model，而不是硬留在 domain entity bucket。
- A3:
  - `codex` 可作為這份 spec owner，先把 domain boundary contract 定到可實作狀態。
- A4:
  - 既有 `Store` 命名與 port shape 只要語義誠實即可保留，不需要為了 DDD 名詞一致性硬改成 `Repository`。
- A5:
  - aggregate root 若存在，在本 repo 預設仍放在 `internal/domain/entities`，不另外新增 `internal/domain/aggregates`。

## Open questions

- Q1:
  - 若 `AddressPolicy` 離開 `entities` 後仍保留 domain 內部存在，最終名稱要維持 `AddressPolicy` 還是改成更明確的 metadata/descriptor 名稱？
  - 本 spec 先以「責任正確優先於命名完美」處理，必要 rename 可在實作時一併完成。

## Success metrics

- Metric:
  - `internal/domain` exported type 的分類一致性。
- Target:
  - 完成後不存在明顯分類錯位的 exported type，例如 query metadata 被放在 `entities`、workflow result 被放在 `policies`、health enum 被放在 `valueobjects`。
- Metric:
  - Address policy / issuance policy responsibility clarity。
- Target:
  - list-facing metadata、issuance capability/config、allocation/preview rule 的 ownership 清楚分離，`ListByChain` 不再必須回傳 `entities.AddressPolicy`。
- Metric:
  - Value object purity。
- Target:
  - canonical VO 不再內建 legacy storage alias 與 unknown-message fallback；這類相容層邏輯移到 adapter/persistence normalizer。
- Metric:
  - Repo contract alignment。
- Target:
  - `AGENTS.md` 明確描述 `entity / aggregate root / event / value object / policy` 的判準，並寫清楚 `Repository / Store / DAO` 的命名邊界；同時新增對 deployment config、workflow result、compatibility parsing 的 review trigger，並寫明不新增 `domain/services` / `domain/enums`、不機械推動 full DDD package migration。
- Metric:
  - Regression safety。
- Target:
  - `go list ./...`、targeted `go test`、以及最終 `go test ./...` 通過。
