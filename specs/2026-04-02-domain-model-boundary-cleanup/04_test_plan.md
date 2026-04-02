---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - domain package boundary cleanup 對 `entities`、`events`、`policies`、`valueobjects` 的影響。
  - address policy / issuance policy reclassification 對 list/generate/allocate flows 的影響。
  - receipt polling 與 webhook dispatch 對 pseudo-policy 移除後的影響。
  - persistence adapter 對 failure reason compatibility normalizer 下沉後的影響。
  - `AGENTS.md` domain modeling contract 更新。
- Not covered:
  - 新 payment feature。
  - schema migration。
  - unrelated bootstrap/infrastructure refactor。

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-002 / FR-006 / NFR-002
  - Steps:
    - `go test ./internal/domain/entities ./internal/domain/events`
  - Expected:
    - 真正保留的 entities 與 event tests 通過。
    - `internal/domain/entities` runtime code 不再依賴 `internal/domain/policies`。
- TC-002:
  - Linked requirements: FR-004 / FR-006 / NFR-002
  - Steps:
    - `go test ./internal/domain/policies`
  - Expected:
    - domain policies tests 通過，且 package 中不再只有 pass-through wrapper。
- TC-003:
  - Linked requirements: FR-005 / NFR-002
  - Steps:
    - `go test ./internal/domain/valueobjects`
  - Expected:
    - canonical VO tests 通過，legacy alias compatibility 測試若存在則已移到 adapters，且 `AddressPolicyID` / `Scheme` 的 canonicalization 有對應測試。
    - repo built-in `AddressPolicyID` constants/helper 有對應測試，且 bootstrap/runtime code 可用集中定義取代散落裸字串。
    - malformed `AddressPolicyID` 有 explicit constructor/validator 測試，不再只靠 exported parser + `Normalize()` 隱式驅動。
    - `internal/domain/valueobjects` 不再保留 `BitcoinAddressScheme` 與 `AddressScheme` 兩套重疊模型。
    - `internal/domain/valueobjects` 不再保留 `BitcoinNetwork` 與 `NetworkID` 兩套重疊模型。
- TC-004:
  - Linked requirements: FR-003 / FR-008 / NFR-002
  - Steps:
    - `go test ./internal/application/usecases`
  - Expected:
    - list/generate/allocate/polling/webhook dispatch 相關 use case tests 通過。
    - malformed `AddressPolicyID` 在 allocate/generate use case 會回 explicit invalid-input error，而不是 `ErrAddressPolicyNotFound`。
- TC-005:
  - Linked requirements: FR-007 / NFR-006
  - Steps:
    - `rg -n 'Entity|Aggregate|aggregate root|Repository|Store|DAO|legacy alias|workflow result|health' AGENTS.md`
  - Expected:
    - `AGENTS.md` 內可檢索到新的 domain classification rules 與錯位反例。

### Integration

- TC-101:
  - Linked requirements: FR-001 / FR-008 / NFR-006
  - Steps:
    - `go list ./...`
  - Expected:
    - 全 repo compile graph 正常，沒有因 package 搬移造成 import cycle 或缺漏。
- TC-102:
  - Linked requirements: FR-003 / FR-008 / NFR-002
  - Steps:
    - `go test ./internal/adapters/outbound/policy ./internal/bootstrap/...`
  - Expected:
    - address policy catalog 與 call sites 更新後測試通過。
- TC-103:
  - Linked requirements: FR-005 / FR-008 / NFR-002
  - Steps:
    - `go test ./internal/adapters/outbound/persistence/...`
  - Expected:
    - persistence scanner / normalizer 更新後測試通過。
    - malformed persisted `AddressPolicyID` 會回 explicit persisted-invalid contract error。
- TC-104:
  - Linked requirements: FR-001 / FR-008 / NFR-001 / NFR-002
  - Steps:
    - `go test ./...`
  - Expected:
    - 全量測試通過。
- TC-105:
  - Linked requirements: FR-001 / FR-007 / NFR-006
  - Steps:
    - `SPEC_DIR="specs/2026-04-02-domain-model-boundary-cleanup" bash scripts/spec-lint.sh`
  - Expected:
    - spec-lint 通過。

### E2E (if applicable)

- Scenario 1:
  - 不適用，本輪沒有新增跨程序產品流程。
- Scenario 2:
  - 不適用。

## Edge cases and failure modes

- Case:
  - `AddressPolicy` 搬離 `entities` 後，某些 read path 仍依賴舊型別。
- Expected behavior:
  - 編譯或 use case tests 直接失敗，直到 call site 全部更新完成。
- Case:
  - compatibility parsing 從 VO 移走後，某個 persistence adapter 漏掉 legacy alias mapping。
- Expected behavior:
  - adapter tests 失敗，顯示 raw storage text 不能正確轉成 canonical code。
- Case:
  - thin wrapper policy 被刪掉後，真正的 lifecycle rule 沒有被 entity 或新 policy 接住。
- Expected behavior:
  - polling / webhook dispatch 測試失敗，暴露遺漏的 domain decision。

## NFR verification

- Performance:
  - `go test ./...` 無顯著新增的 runtime IO 或外部依賴。
- Reliability:
  - domain、application、adapter 層既有回歸測試持續通過。
- Security:
  - payment/notification/issuance 相關邏輯只搬責任，不放寬驗證。

## Execution result

- Result:
  - PASS
- Executed commands:
  - `go test ./internal/domain/valueobjects ./internal/bootstrap`
  - `go list ./...`
  - `go test ./...`
  - `SPEC_DIR="specs/2026-04-02-domain-model-boundary-cleanup" bash scripts/spec-lint.sh`
  - `bash scripts/precommit-run.sh`
- Notes:
  - `AddressPolicyID` constants/helper 仍覆蓋到 bootstrap runtime path，但 file locality 已收斂回單一 VO file。
  - 這一輪也補齊了 explicit invalid-ID semantics: malformed `AddressPolicyID` 不再與 `not found` 混淆，persisted malformed IDs 也有獨立 contract error。
