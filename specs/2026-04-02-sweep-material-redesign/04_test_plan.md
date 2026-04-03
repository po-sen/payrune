---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - domain/entity contract cleanup
  - deriver output contract cleanup
  - allocation store completion contract redesign
  - chain-specific sweep material JSON assembly
  - regression of bitcoin / ethereum create2 payload shape
- Not covered:
  - external operator tooling behavior outside repo tests
  - DB migration, because this spec intentionally avoids schema changes

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001 / NFR-006
  - Steps:
    - `go test ./internal/domain/entities`
  - Expected:
    - `PaymentAddressAllocation` tests 通過。
    - entity 不再要求或持有 `SweepMaterialJSON`。
- TC-002:
  - Linked requirements: FR-001 / FR-003 / NFR-006
  - Steps:
    - `go test ./internal/application/usecases ./internal/adapters/outbound/blockchain`
  - Expected:
    - deriver output 與 use case tests 通過。
    - deriver output 在既有 contract file 內以 plain `SweepMaterialJSON` 傳遞，不新增額外 model layer。
    - application ports 仍保持 boring contract 風格。

### Integration

- TC-101:
  - Linked requirements: FR-002 / FR-004 / FR-005 / NFR-001 / NFR-002
  - Steps:
    - `go test ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres`
  - Expected:
    - `Complete` 仍寫出既有 JSON payload。
    - `Complete` 失敗時不會寫出半套 issued 狀態。
    - 當 `SweepMaterialJSON` 缺失或無效時，store 不應發出任何 DB update。
    - generic `UnitOfWork` tests 不需要知道 sweep-material 專屬依賴。
- TC-102:
  - Linked requirements: FR-004 / FR-005 / NFR-005
  - Steps:
    - `go test ./internal/adapters/outbound/bitcoin ./internal/adapters/outbound/ethereum`
  - Expected:
    - bitcoin / ethereum chain-specific payload tests 通過。
    - deriver 所需 metadata 與 issuance ref flow 維持一致。

### E2E (if applicable)

- Scenario 1:
  - 不適用，本輪不新增外部行為。
- Scenario 2:
  - 不適用。

## Edge cases and failure modes

- Case:
  - deriver 無法產生 `sweep_material_json`
- Expected behavior:
  - use case 失敗，不寫入 issued allocation。

## NFR verification

- Performance:
  - `Complete` 不新增額外 DB/query hop；review 與 tests 確認 builder 是 in-process logic。
- Reliability:
  - `go test ./...` 通過，且 store tests 覆蓋 failure path。
- Security:
  - 重構後 payload 欄位集合與既有一致，未額外擴充 secret surface。
