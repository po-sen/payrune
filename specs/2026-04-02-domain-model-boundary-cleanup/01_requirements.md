---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Entity:
  - 具有 identity、可追蹤 lifecycle，且其 business behavior 會決定合法狀態轉換的 domain object。
- Aggregate:
  - 一個需要一致性保護的 business boundary；它是邊界角色，不是另一種 top-level package 分類。
- Aggregate root:
  - aggregate 對外唯一入口；在本 repo 中通常同時也是 entity，且預設仍放在 `internal/domain/entities`。
- Event:
  - 由 domain state transition 產生的不可變 business fact。
- Value object:
  - 以值定義、可驗證、可 canonicalize 的 domain concept，不靠 identity 表示自身；在本 repo 中也可包含 enum-like 的 canonical domain scalar code。
- Domain policy:
  - 不能自然只放在單一 entity method 內的商業決策規則，通常跨越多個 object、status set、或 configuration-backed rule。
- Workflow result:
  - 某個 use case / outbox / delivery pipeline 執行後要持久化的技術狀態，不等於 domain policy。
- Compatibility normalizer:
  - 將 legacy DB text、vendor error 或歷史 alias 映射成 canonical code 的邏輯；這是 adapter/persistence concern，不是 core VO concern。
- Repository:
  - 代表 aggregate-like domain object collection 的 outbound port，主語意是 load/save aggregate。
- Store:
  - 代表流程型或技術型 persistence 的 outbound port，主語意是 claim/reserve/lease/retry/cursor 等 process operation。
- DAO:
  - adapter 內部的 DB/table/query helper 術語，不是 application outbound port 命名。

## Out-of-scope behaviors

- OOS1:
  - 新增 payment lifecycle 狀態、通知 payload 欄位、或 issuance capability。
- OOS2:
  - 更改 schema 來追求新的 domain taxonomy，除非實作證明沒有 schema 無法保持正確邊界。
- OOS3:
  - 將所有現有型別改成全新命名，只為了視覺一致。
- OOS4:
  - 將 `internal/domain` 改造成 feature-oriented package tree。
- OOS5:
  - 為了分類方便新增 `internal/domain/services` 或 `internal/domain/enums`。

## Functional requirements

### FR-001 - Audit and reclassify every current domain export

- Description:
  - 本輪必須盤點目前 `internal/domain` 的 exported type，對每個型別做出 keep/move/remove decision，不能保留「名字像某分類、內容卻不是」的灰色地帶。
- Acceptance criteria:
  - [ ] `internal/domain` 內每個 exported type 都能被標記為 `entity`、`event`、`value object`、`policy`，或被移出 domain。
  - [ ] 盤點與決策結果在 spec 與 code layout 中一致。
  - [ ] 不再保留明顯靠 package 名誤導 review 的型別。
- Notes:
  - 這是本 spec 的總體要求，其他 FR 會針對高訊號區塊細化。

### FR-002 - Keep entities limited to identity, lifecycle, and business transitions

- Description:
  - `internal/domain/entities` 只能保留真正具有 identity、lifecycle、business transitions 的型別；list metadata、deployment config、或由外部 config 推導狀態的 bundle 不得再放在這裡。
- Acceptance criteria:
  - [ ] `PaymentAddressAllocation` 與 `PaymentReceiptTracking` 保持為 entity，並繼續擁有其狀態轉移行為。
  - [ ] 若本輪明確標記 aggregate root，仍以業務名詞放在 `internal/domain/entities`，不新增 `internal/domain/aggregates`。
  - [ ] `internal/domain/entities` 不再保留只提供 metadata/normalize 的型別。
  - [ ] 不再有 entity 的核心狀態是由巢狀 config 在 normalize 時回填出來。
  - [ ] `internal/domain/entities` 的 runtime code 不得 import `internal/domain/policies`；若 policy 需要影響 entity，必須先算出 plain values 或 value-object snapshot 再傳入 entity。
- Notes:
  - 像 `AddressPolicy` 這類 list-facing metadata，若沒有 entity lifecycle，應移出 `entities`。

### FR-003 - Model address-policy and issuance-policy responsibilities explicitly

- Description:
  - public list metadata、issuance capability/rule、以及 deployment/runtime config 必須被清楚拆開，不能再共用一個「看起來像 entity」的 bucket。
- Acceptance criteria:
  - [ ] `ListByChain` read path 不再必須回傳 `entities.AddressPolicy`。
  - [ ] issuance rule/capability 若仍屬 domain，應放在 `internal/domain/policies` 或等價的正確 category，而不是 `entities`。
  - [ ] bootstrap / adapter 組 policy catalog 的流程不再迫使 domain 承擔 query-only metadata shape。
  - [ ] `AddressPolicyID` 不再只作為裸字串出現在 domain 核心型別中；應有對應的 typed domain scalar / value object。
  - [ ] repo 內建的 address policy IDs 應以 well-known constants 或 helper 集中定義，而不是在 runtime code 中散落硬編碼字串。
  - [ ] `AddressPolicyID` 的 constants/helper 若規模仍小，應與該 VO 保持高 locality，不應為了機械分檔而拆成多餘 sibling file。
  - [ ] malformed `AddressPolicyID` 在 application boundary 必須回傳 explicit invalid-input error，不能與 unsupported/unknown policy 共用同一個 `not found` 結果。
  - [ ] persistence/read adapter 若讀到 malformed persisted `AddressPolicyID`，必須回傳 explicit `outport.Err...PersistedAddressPolicyIDInvalid` contract error，而不是靜默 normalize 成 zero value。
  - [ ] `AddressPolicyID` 的 public construction surface 必須表達明確用途；不要保留只被同檔 `Normalize()` 內部使用的多餘 exported parser API。
  - [ ] address `Scheme` 不再只作為裸字串出現在 domain 核心型別中；應有對應的 typed domain scalar / value object。
  - [ ] `internal/domain/valueobjects` 不再同時維持 `AddressScheme` 與另一套 Bitcoin-only scheme truth。
  - [ ] `internal/domain/valueobjects` 不再同時維持 `NetworkID` 與另一套 Bitcoin-only network truth。
- Notes:
  - 若 list path 只是 query/read model，application 層 read model 比 domain entity 更合適。
  - `AddressPolicyID` 仍是 open identifier，不應為了 repo 目前的 built-in catalog 被誤建模成 closed enum。

### FR-004 - Keep policies limited to real business decision rules

- Description:
  - `internal/domain/policies` 只能保留真正的商業決策規則；pass-through wrapper、workflow save-result carrier、或純 retry bookkeeping 不得再以 domain policy 名義存在。
- Acceptance criteria:
  - [ ] `internal/domain/policies` 不再保留只包一層 entity method 的 thin wrapper。
  - [ ] webhook delivery sent/retry/failed 的持久化 result carrier 若仍存在，必須移到 application/outbox 或其他非-domain workflow boundary。
  - [ ] 若 polling eligibility status set 仍屬 domain rule，必須由明確的 domain policy 表達，而不是塞在 entity package 的 helper function。
- Notes:
  - policy 的核心是 decision，不是 result row。

### FR-005 - Keep value objects canonical and free from adapter compatibility logic

- Description:
  - `internal/domain/valueobjects` 應表達 canonical business values、validation、有限的 canonicalization，以及必要的 enum-like domain scalar code；legacy alias 與 unknown-text fallback 必須移出 core value layer。
- Acceptance criteria:
  - [ ] failure-reason VO 不再直接吸收 legacy DB text alias 或 unknown vendor message fallback。
  - [ ] persistence adapters 在 scan/parse 邊界自行完成 legacy text 到 canonical VO 的映射。
  - [ ] 非支付領域的 health/status enum 不再留在 `internal/domain/valueobjects`。
  - [ ] `internal/domain/valueobjects` 的 runtime code 不得 import `internal/domain/policies`。
  - [ ] Bitcoin-specific address encoder routing type 若仍存在，必須屬於 bitcoin adapter 內部，不再放在 domain valueobjects。
  - [ ] Bitcoin-specific network/chaincfg routing type 若仍存在，必須屬於 bitcoin adapter 內部，不再放在 domain valueobjects。
  - [ ] `AddressPolicyID` 與 address `Scheme` 若被提升為 value object / typed domain scalar，需在 value layer 提供 canonical normalization。
- Notes:
  - trim/lowercase 這類 canonicalization 可以留在 VO；歷史 storage compatibility 不應留在 VO。
  - 本 repo 不為此新增獨立的 `domain/enums` bucket。

### FR-006 - Preserve and clarify domain events

- Description:
  - 真正由 domain transition 產生的事件應保留在 `internal/domain/events`，並維持 immutable fact 的語義。
- Acceptance criteria:
  - [ ] `PaymentReceiptStatusChanged` 仍由 domain transition 產生。
  - [ ] `internal/domain/events` 不新增 outbox row、delivery attempt、或技術型 workflow message。
  - [ ] event construction validation 與 call site 維持清楚且可測。
- Notes:
  - 這一條主要是保護正確的 event，不讓 cleanup 過度把真正 event 也抽走。

### FR-007 - Align AGENTS.md with the stricter domain modeling contract

- Description:
  - repo-level `AGENTS.md` 必須明確寫出 entity / event / value object / policy 的分類規則，以及本 repo 特別容易犯錯的反例，同時定義本 repo 採用的是 pragmatic clean architecture，而非 full DDD package migration。
- Acceptance criteria:
  - [ ] `AGENTS.md` 新增或更新 domain category rules。
  - [ ] `AGENTS.md` 明確寫出 entity 與 aggregate root 的關係，並說明 aggregate root 在本 repo 預設仍放在 `entities`。
  - [ ] `AGENTS.md` 明確寫出何時使用 `Repository`、`Store`、`Reader/Finder`，以及 `DAO` 只屬於 adapter 內部實作。
  - [ ] `AGENTS.md` 明確寫出 domain dependency direction：`valueobjects` 不依賴 `policies`，`entities` 也不依賴 `policies`，而是消費 plain values 或 value-object snapshots。
  - [ ] `AGENTS.md` 明確指出 deployment/runtime catalog、query-only shape、workflow result、health response、compatibility normalizer 預設不屬於 domain。
  - [ ] `AGENTS.md` 新增對「entity 狀態由外部 config 推導」與「VO 吸收 legacy alias」等 review trigger。
  - [ ] `AGENTS.md` 明確寫出本輪不新增 `domain/services`、`domain/enums`，且不機械同步重組 `internal/application` package strategy。
- Notes:
  - `AGENTS.md` 需要與這份 spec 同步，避免實作完但 repo contract 仍模糊。

### FR-008 - Keep cleanup behavior-preserving

- Description:
  - 本輪是 domain boundary cleanup，不是功能改版；應透過 targeted tests 與 compile checks 證明行為未改。
- Acceptance criteria:
  - [ ] `go list ./...` 通過。
  - [ ] `go test ./internal/domain/...` 通過。
  - [ ] 受 domain category 搬動影響的 application/adapters/bootstrap 測試通過。
- Notes:
  - 若型別搬 package，需連同 call sites 與 tests 一起更新。

## Non-functional requirements

- Performance (NFR-001):
  - 本輪不得引入新的 runtime network/DB IO。
  - 驗證方式: `go test ./...` 時間級別保持在一般本地可接受範圍。
- Availability/Reliability (NFR-002):
  - receipt polling、address allocation、address preview、receipt webhook dispatch 的成功/失敗行為不變。
  - 驗證方式: 既有相關 unit/integration tests 維持通過。
- Security/Privacy (NFR-003):
  - webhook delivery / receipt status / address issuance 的安全假設不得因模型搬移而放寬。
  - 驗證方式: 無新增 bypass path，既有 tests 維持通過。
- Compliance (NFR-004):
  - 不適用，本輪無新增法規或資料分類需求。
- Observability (NFR-005):
  - 本輪不要求新增 metrics/logs，但不得刪除既有診斷所需的 error identity 與測試訊號。
- Maintainability (NFR-006):
  - 完成後 reviewer 可以只看 package 與型別名稱，就大致判斷責任方向是否合理。
  - 驗證方式: `internal/domain` 不再出現高訊號錯位型別；`AGENTS.md` 與實作一致。

## Dependencies and integrations

- External systems:
  - 無新增 external integration。
- Internal services:
  - `internal/domain`
  - `internal/application/usecases`
  - `internal/application/ports/outbound`
  - `internal/application/outbox`
  - `internal/adapters/outbound/persistence/*`
  - `internal/bootstrap`
