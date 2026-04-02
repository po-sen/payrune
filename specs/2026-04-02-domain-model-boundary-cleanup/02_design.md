---
doc: 02_design
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

# Technical Design

## High-level approach

- Summary:
  - 先以 exported type 為單位做 domain audit，再按責任重新安置型別，而不是先從 package 名倒推理論分類。
  - 本輪採用 pragmatic clean architecture：先把 type ownership 做準，不把 cleanup 升級成 full DDD、feature-oriented package tree、或 repository-over-store 的全面遷移。
  - 本輪的核心原則是:
    - 真正的 business state transition 留在 domain entity / event。
    - 真正的 business decision 留在 domain policy。
    - canonical code/value 與必要的 enum-like domain scalar 留在 value object layer。
    - query-only metadata、deployment catalog、workflow result、compatibility parsing、health enum 離開錯誤的 domain bucket。
  - Key decisions:
    - 維持目前 `internal/domain/{entities,events,valueobjects,policies}` 的 top-level 目錄，不新增 `services`、`enums` 等新 bucket。
    - `PaymentAddressAllocation` 與 `PaymentReceiptTracking` 保留為 domain entities，並可視為目前的單一-entity aggregate roots。
    - aggregate root 在本 repo 預設仍留在 `internal/domain/entities`，不額外新增 `internal/domain/aggregates`，也不以 `...Aggregate` suffix 重新命名。
    - domain dependency direction 以 `valueobjects <- entities <- policies` 為準；`valueobjects` 不依賴 `policies`，`entities` 也不依賴 `policies`。
    - `AddressPolicyID` 以 dedicated typed domain scalar 建模，而不是在 domain 核心持續使用裸字串。
    - `AddressPolicyID` 保持 open typed identifier；repo 既知 built-in policy IDs 以 non-exhaustive constants/helper 補齊，不升級成 closed enum。若常數集仍小，維持在同一個 VO 檔案即可，不額外拆分多餘 sibling file。
    - `AddressPolicyID` 對外提供明確 constructor/validator surface；內部 parse helper 不再作為無實際 consumer 的 exported API 暴露。
    - application inbound path 需在 lookup 前先明確驗證 `AddressPolicyID`，讓 malformed input 回到 explicit invalid-input error，而不是被 `Normalize()` 吃成 zero value 後誤判為 policy not found。
    - persistence/read adapter 在 scan raw `address_policy_id` 後，必須先顯式 parse 成 canonical `AddressPolicyID`；若 persisted value malformed，回傳對應的 `outport.Err...PersistedAddressPolicyIDInvalid` contract error。
    - address `Scheme` 以 dedicated typed domain scalar 建模，統一 Bitcoin 與 Ethereum 已知 scheme 的 canonical 值；DTO/HTTP 仍可保留字串。
    - `BitcoinAddressScheme` 不再作為 domain VO 與 `AddressScheme` 並存；若 bitcoin adapter 仍需要 encoder-routing type，改成 adapter-local 型別。
    - `BitcoinNetwork` 不再作為 domain VO 與 `NetworkID` 並存；若 bitcoin adapter 仍需要 chaincfg/esplora routing type，改成 adapter-local 型別。
    - `PaymentReceiptStatusChanged` 保留為 domain event。
    - `AddressIssuanceConfig` 可保留為 value object，只要它仍代表 canonical issuance configuration value，而不是 process env reader。
  - `AddressIssuancePolicy` 不再視為 entity；若保留在 domain，應移到 `internal/domain/policies` 或等價位置。
  - list-facing `AddressPolicy` metadata 不再留在 `internal/domain/entities`；優先降到 application read model，若仍需 domain 內共享，則只能作為 VO/descriptor，不作為 entity。
  - `PaymentReceiptTrackingLifecyclePolicy` 這類 thin wrapper 應被刪除或折回真正的 domain owner。
  - `PaymentReceiptStatusNotificationDeliveryResult` 與 sent/retry/fail helper 移到 application/outbox or outbound workflow boundary。
  - failure-reason 的 legacy alias / fallback mapping 改由 persistence adapter normalizer 持有。
  - `ServiceStatus` 移出 domain，回到 health-check boundary。
    - `Repository` 只保留給 aggregate collection 語意明確的 outbound port；`Store` 仍用於 claim/reserve/lease/retry/cursor 等流程型 persistence。
    - `Reader` / `Finder` / `Outbox` 等更窄的 port 名稱應優先於 generic `Repository` 或 `Store`，只要它們更能表達實際邊界。
    - `DAO` 視為 adapter 內部 row/query helper 術語，不作為 application outbound port 命名。
    - `internal/application` 只做配合 domain cleanup 的 call site 與 type-boundary 調整，不做全面 package strategy 統一。

## System context

- Components:
  - Domain:
    - `PaymentAddressAllocation`
    - `PaymentReceiptTracking`
    - `PaymentReceiptStatusChanged`
    - canonical IDs / statuses / observation / failure codes
    - issuance rules after reclassification
  - Application:
    - allocate / generate / list / receipt polling / webhook dispatch use cases
    - read models and outbox workflow payloads
  - Adapters:
    - persistence scanners / normalizers
    - policy reader adapter
    - webhook notifier / outbox persistence
  - Bootstrap:
    - deployment-owned policy catalog assembly
- Interfaces:
  - Outbound ports should consume either domain-native business types or application read/workflow models, not mislabeled entities.
  - Persistence adapters are responsible for turning raw storage text into canonical domain values.
  - Aggregate roots, if called out explicitly, remain ordinary domain entities from the package layout perspective.
  - 如果 policy 需要提供 entity transition 所需資訊，application/policy 要先攤平成 plain values 或 value-object snapshot，再呼叫 entity method。
  - `AddressPolicyReader.FindIssuanceByID` 這類已吃 typed scalar 的 port，應假設 caller 傳入的已是 canonical ID，不再在 port 內部偷偷 normalize 來掩蓋 boundary validation 漏洞。

## Key flows

- Flow 1:
  - Address policy listing:
    - bootstrap assembles runtime policy catalog
    - policy reader exposes list-facing read model
    - list use case maps read model to response DTO
    - no fake entity is needed just to carry query metadata
- Flow 2:
  - Address allocation / preview validation:
    - use case loads issuance rule object
    - issuance capability and validation remain in domain policy
    - use case orchestrates ports without assuming the rule object is an entity
- Flow 3:
  - Receipt polling lifecycle:
    - entity applies observation and owns sticky timestamps / state transitions
    - if any remaining cross-object lifecycle rule still exists, it is expressed as a real domain policy
    - otherwise the thin wrapper policy is removed
- Flow 4:
  - Webhook dispatch delivery bookkeeping:
    - dispatch use case decides success/failure path
    - delivery result object lives at application/outbox boundary
    - outbox store persists the workflow result without pretending it is a domain policy artifact
- Flow 5:
  - Persistence scan:
    - adapter reads raw DB text
    - adapter-level normalizer maps legacy aliases or unknown historic text to canonical failure code
    - canonical VO enters domain after normalization

## Diagrams (optional)

- Mermaid sequence / flow:
  - `raw storage text -> adapter normalizer -> canonical VO -> domain entity/policy -> domain event or application workflow result`

## Data model

- Entities:
  - Keep:
    - `PaymentAddressAllocation`
    - `PaymentReceiptTracking`
  - Remove from entity bucket:
    - list-facing address policy metadata
    - issuance rule/config bundle if it has no entity lifecycle
- Schema changes or migrations:
  - 預期不需要 schema 變更。
  - 若實作時發現某個 misclassified type 目前被 schema shape 綁死，需另行評估是否拆成後續 spec，而不是在本 spec 中偷偷加 migration。
- Consistency and idempotency:
  - Canonical domain values must enter domain after adapter normalization, not before。
  - package 搬移不應改變 allocation / polling / dispatch / listing 行為。
  - `AddressPolicyID` 與 address `Scheme` 進入 domain 後應以 typed domain scalar 形式流動；真正的 transport/persistence string 只留在邊界上。
  - repo built-in policy IDs 應由 `valueobjects.AddressPolicyID` 的 sibling constants/helper 檔案集中提供，例如 bitcoin 固定 IDs 與 `EthereumCreate2AddressPolicyID(network)`。
  - Bitcoin-specific encoder/deriver 選擇若仍需要 narrower type，必須在 adapter 內由 `AddressScheme -> adapter-local bitcoin scheme` 轉換完成。
  - Bitcoin-specific chaincfg/esplora routing 若仍需要 narrower type，必須在 adapter 內由 `NetworkID -> adapter-local bitcoin network` 轉換完成。

## API or contracts

- Endpoints or events:
  - 對外 API 與 webhook payload 預期不變。
  - `PaymentReceiptStatusChanged` event shape 預期不變。
- Request/response examples:
  - 不新增外部 contract，本輪主要是內部 type ownership 調整。

## Backward compatibility (optional)

- API compatibility:
  - 維持既有 HTTP / webhook contract。
- Data migration compatibility:
  - 既有資料表欄位與 enum string 先保持不變。
  - legacy text 兼容若仍需要，由 adapter normalizer 承接。

## Failure modes and resiliency

- Retries/timeouts:
  - 不新增新的 retry 機制。
  - webhook delivery retry bookkeeping 若離開 domain，仍需保持既有行為與最大重試次數邏輯。
- Backpressure/limits:
  - 不適用，本輪不增加新的 runtime path。
- Degradation strategy:
  - 若重分類導致某個 type 目標位置不明確，優先選擇「更靠近消費者的邊界」，不要硬留在 domain。

## Observability

- Logs:
  - 不新增特別 log 要求。
- Metrics:
  - 無新增 metrics。
- Traces:
  - 無。
- Alerts:
  - 無。

## Security

- Authentication/authorization:
  - 不適用。
- Secrets:
  - issuance config / webhook dispatch 行為不可因 package 搬移而暴露更多 secret handling。
- Abuse cases:
  - 不得因「把 legacy alias 移到 adapter」而放寬未知字串的容忍方式；unknown fallback 必須是顯式 adapter decision。

## Alternatives considered

- Option A:
  - 維持目前 package 名不動，只在 review 時口頭約束。
- Option B:
  - 把所有型別都硬留在 domain，只調整註解與文件。
- Option C:
  - 以 exported type audit 為基準，真正移動錯位型別，並同步更新 `AGENTS.md`。
- Option D:
  - 直接把 `internal/domain` 改成以業務概念切的 package tree，並同步重組 `internal/application`。
- Why chosen:
  - 使用者已明確指出目前內容「不太對勁」；單靠命名解釋或口頭約束，無法防止下一輪再度漂移。
  - 但目前最需要的是把 type ownership 做準，不是同時推動更大的 package strategy migration。

## Risks

- Risk:
  - `AddressPolicy` / `AddressIssuancePolicy` 的最終命名與目標 package 可能牽動較多 call site。
- Mitigation:
  - 先固定責任，再用 targeted tests 與 compile checks 控制 churn。
- Risk:
  - 去掉 thin wrapper policy 後，use case 可能暫時暴露更多 domain orchestration 細節。
- Mitigation:
  - 只有在 entity 自己能自然承接時才折回 entity；若仍是獨立決策則保留為真正 policy。
- Risk:
  - 把 compatibility parsing 移到 adapter 後，兩個 persistence adapter 可能重複同樣 mapping。
- Mitigation:
  - 在各 adapter 內建立局部共用 normalizer；不要把 compatibility logic 拉回 domain 只為了去重。
- Risk:
  - `AGENTS.md` 規則若寫太抽象，之後仍不足以阻止錯位。
- Mitigation:
  - 寫入具體反例與 review trigger，而不只寫概念名詞。
