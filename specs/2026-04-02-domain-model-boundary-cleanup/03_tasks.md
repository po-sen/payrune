---
doc: 03_tasks
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

# Task Plan

## Mode decision

- Selected mode: Full
- Rationale:
  - 本輪不是單一 type rename，而是要重整 domain taxonomy、跨 application/adapters/bootstrap 調整 call site、並同步更新 repo contract `AGENTS.md`。
  - 同時需要明確鎖住 scope，避免 cleanup 失控成 full DDD package migration。
- Upstream dependencies (`depends_on`):
  - `2026-03-07-architecture-naming-refactor`
  - `2026-03-24-architecture-conformance-refactor`
  - `2026-03-28-allocation-failure-reason-typing`
  - `2026-03-29-allocation-issuance-naming`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: 不適用，本 spec 採 Full mode。
  - What would trigger switching to Full mode: 不適用。
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): 不適用，本 spec 產出 test plan。

## Milestones

- M1:
  - 完成新 spec 與 domain audit 決策。
- M2:
  - 完成 address-policy / issuance-policy reclassification。
- M3:
  - 完成 pseudo-policy / workflow result / health enum 清理。
- M4:
  - 完成 compatibility normalizer 下沉到 adapters。
- M5:
  - 完成 `AGENTS.md` domain modeling 契約更新。
- M6:
  - 完成 compile/test/spec-lint 驗證並收斂 spec 狀態。

## Tasks (ordered)

1. T-001 - Scaffold and lock the new domain cleanup spec
   - Scope:
     - 建立 `2026-04-02-domain-model-boundary-cleanup` Full-mode spec，記錄 domain type audit、分類原則與驗證策略。
   - Output:
     - `specs/2026-04-02-domain-model-boundary-cleanup/` 五份 spec 文件。
   - Linked requirements: FR-001, FR-007, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `SPEC_DIR="specs/2026-04-02-domain-model-boundary-cleanup" bash scripts/spec-lint.sh`
     - [ ] Expected result: spec-lint 通過。
     - [ ] Logs/metrics to check (if applicable): 無。
2. T-002 - Reclassify address policy modeling

   - Scope:
     - 將 `AddressPolicy` 從 `entities` bucket 拿開，並將 issuance capability/rule 重新放到正確 category。
     - 同步修正 `AddressPolicyReader`、list/allocate/generate use cases、bootstrap catalog 組裝與相關 adapters。
     - entity 若需要 issuance rule 的結果，只能吃 plain values 或 value-object snapshot，不可直接依賴 policy type。
     - 將 `AddressPolicyID` 與 address `Scheme` 提升為 typed domain scalar / value object，並更新相關 domain 與 application 型別。
     - 補齊 `AddressPolicyID` 的 repo built-in constants/helper，並將 bootstrap/runtime code 的裸字串 ID 收斂到集中定義；若常數集仍小，保持與 `AddressPolicyID` 同檔高 locality，不保留多餘的 constants sibling file。
     - 將 malformed `AddressPolicyID` 與 unknown policy 的語義拆開：application boundary 要先顯式驗證，再做 lookup。
     - 將 `AddressPolicyID` 的無實際 consumer 的 exported parser surface 收斂成明確 constructor/validator API。
     - 將 persisted malformed `AddressPolicyID` 的 read/scan path 改成 explicit contract error，而不是靜默 normalize 成 zero value。
     - 將 `BitcoinAddressScheme` 從 domain valueobjects 移除；若 bitcoin adapter 仍需 narrower routing type，改為 adapter-local 型別與 helper。
     - 將 `BitcoinNetwork` 從 domain valueobjects 移除；若 bitcoin adapter 仍需 narrower routing type，改為 adapter-local 型別與 helper。
     - 不新增 `domain/services` / `domain/enums`，也不順手重組整個 `internal/application` package strategy。
   - Output:
     - 更新後的 address-policy related types、ports、use cases、adapters、bootstrap wiring。
   - Linked requirements: FR-001, FR-002, FR-003, FR-008, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/application/usecases ./internal/adapters/outbound/policy ./internal/bootstrap/...`
     - [ ] Expected result: list/generate/allocate 相關測試通過，且 `ListByChain` 不再回傳 `entities.AddressPolicy`，`internal/domain/entities` 也不再 import `internal/domain/policies`，`AddressPolicyID` / `Scheme` 已在核心模型中 typed 化，runtime code 不再散落 built-in policy ID 裸字串，malformed `AddressPolicyID` 不再與 not-found 共用結果，persisted malformed `AddressPolicyID` 也有 explicit contract error，domain 不再保留重複的 Bitcoin-only scheme/network VO。
     - [ ] Logs/metrics to check (if applicable): 無。

3. T-003 - Remove pseudo-policies and misplaced workflow carriers from domain
   - Scope:
     - 清掉 `PaymentReceiptTrackingLifecyclePolicy` 這類 thin wrapper。
     - 將 receipt webhook delivery result/helper 移到 application/outbox 或其他正確 workflow boundary。
     - 將 `ServiceStatus` 移出 domain。
   - Output:
     - 更新後的 receipt polling / webhook dispatch / health-check 邊界型別。
   - Linked requirements: FR-001, FR-004, FR-005, FR-008, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/domain/... ./internal/application/usecases/...`
     - [ ] Expected result: domain tests與受影響 use case tests 通過，domain policy package 不再承載 workflow result carrier。
     - [ ] Logs/metrics to check (if applicable): 無。
4. T-004 - Move compatibility parsing out of canonical value objects
   - Scope:
     - 將 failure reason 類 VO 的 legacy alias / fallback mapping 移到 persistence adapter local normalizer。
     - 保留 canonical VO code 與 validation。
   - Output:
     - 更新後的 VO、adapter normalizer、persistence scanner 與測試。
   - Linked requirements: FR-001, FR-005, FR-008, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/domain/valueobjects ./internal/adapters/outbound/persistence/...`
     - [ ] Expected result: VO tests 與 persistence adapter tests 通過，VO 不再直接吸收 legacy alias。
     - [ ] Logs/metrics to check (if applicable): 無。
5. T-005 - Tighten AGENTS.md domain modeling rules
   - Scope:
     - 在 `AGENTS.md` 寫清楚 entity / aggregate root / event / value object / policy 的判準，並加入本 repo 的錯位反例與 review trigger。
     - 補上 `Repository` / `Store` / `Reader` / `Finder` / `DAO` 的命名規則，說明哪些屬於 application port，哪些只屬於 adapter 內部實作。
     - 清掉 `AGENTS.md` 後段 embedded generic guidance 與 repo-specific stance 的衝突，例如預設 `domain/services`、過度偏向 `Repository` 的 naming guidance 等。
   - Output:
     - 更新後的 `AGENTS.md`。
   - Linked requirements: FR-007, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `rg -n 'aggregate root|Repository|Store|Reader|Finder|DAO|legacy alias|workflow result|health|domain/services' AGENTS.md`
     - [ ] Expected result: `AGENTS.md` 明確包含新的 domain category rules 與錯位警訊。
     - [ ] Logs/metrics to check (if applicable): 無。
6. T-006 - Final compile and regression verification
   - Scope:
     - 跑 compile/test/spec lint，確認 domain cleanup 沒有改壞既有行為。
   - Output:
     - 最終驗證結果與可實作/可合併的狀態。
   - Linked requirements: FR-006, FR-008, NFR-001, NFR-002, NFR-003, NFR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go list ./... && go test ./... && SPEC_DIR="specs/2026-04-02-domain-model-boundary-cleanup" bash scripts/spec-lint.sh`
     - [ ] Expected result: 全部通過。
     - [ ] Logs/metrics to check (if applicable): 無。

## Traceability (optional)

- FR-001 -> T-001, T-002, T-003, T-004
- FR-002 -> T-002
- FR-003 -> T-002
- FR-004 -> T-003
- FR-005 -> T-004
- FR-006 -> T-003, T-006
- FR-007 -> T-001, T-005
- FR-008 -> T-002, T-003, T-004, T-006
- NFR-001 -> T-006
- NFR-002 -> T-002, T-003, T-004, T-006
- NFR-003 -> T-006
- NFR-005 -> T-006
- NFR-006 -> T-001, T-002, T-003, T-004, T-005, T-006

## Rollout and rollback

- Feature flag:
  - 不適用，本輪為 internal refactor。
- Migration sequencing:
  - 預期無 migration。
- Rollback steps:
  - 若某個 reclassification 導致 call site churn 過大，先回退該型別搬移，保留 spec 與 `AGENTS.md` 作為下一輪 source of truth。

## Completion

- Completed on:
  - 2026-04-02
  - Outcome:
    - `AddressPolicyID` 的 repo built-in constants 與 `EthereumCreate2AddressPolicyID(network)` helper 已保留，但已收回 [`address_policy_id.go`](/Users/posen/Desktop/payrune/internal/domain/valueobjects/address_policy_id.go) 同檔，避免多餘的 `address_policy_id_constants.go` sibling file。
    - bootstrap policy catalog 與 bitcoin xpub env-key mapping 持續使用集中定義的 typed ID，而不是散落裸字串。
    - `AddressPolicyID` 仍維持 open identifier，沒有被收斂成 closed enum。
    - `AddressPolicyID` 現在以明確 constructor/validator surface 建模；malformed input 會回 explicit invalid error，不再與 unknown policy 共用 `not found` 結果。
    - persisted malformed `AddressPolicyID` 讀取時會回 explicit `outport.Err...PersistedAddressPolicyIDInvalid` contract error，而不是靜默 normalize 成 zero value。
    - `AddressPolicyReader` 與 in-memory test reader 對 typed `AddressPolicyID` 改成 direct lookup，不再在 port 內部偷偷 normalize 來掩蓋 boundary validation 漏洞。
- Validation evidence:
  - `go test ./internal/domain/valueobjects ./internal/application/usecases ./internal/adapters/inbound/http/controllers ./internal/adapters/outbound/policy ./internal/adapters/outbound/persistence/postgres/...`
  - `go test ./internal/adapters/outbound/persistence/cloudflarepostgres/... ./internal/bootstrap`
  - `go test ./internal/domain/valueobjects ./internal/bootstrap`
  - `go list ./...`
  - `go test ./...`
  - `SPEC_DIR="specs/2026-04-02-domain-model-boundary-cleanup" bash scripts/spec-lint.sh`
  - `bash scripts/precommit-run.sh`
