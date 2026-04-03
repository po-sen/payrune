---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Sweep material:
  - 寫入 `address_policy_allocations.sweep_material_json` 的 operator-facing recovery document。
- Issuance metadata:
  - 完成 allocation issued transition 時可取得的 policy、issuance ref kind/ref、issued allocation 等資料。

## Out-of-scope behaviors

- OOS1:
  - 不將 `sweep_material_json` 變更為另一種 DB representation。
- OOS2:
  - 不新增新的公開 API 或 external operator payload format。

## Functional requirements

### FR-001 - Remove serialized sweep material from the domain entity

- Description:
  - JSON document 不得再由 domain entity 直接持有。
- Acceptance criteria:
  - [ ] `PaymentAddressAllocation` 不再有 `SweepMaterialJSON` runtime field。
  - [ ] `allocation.MarkIssued(...)` 不再接受 raw JSON string。
- Notes:
  - 這裡移除的是 domain 內的 serialized representation，不是移除 operator payload 這個需求本身。

### FR-002 - Keep sweep material JSON as a plain deriver-to-store contract

- Description:
  - `sweep_material_json` 由 chain-specific issued deriver adapter 直接產出，供 allocation completion 寫入 DB。
- Acceptance criteria:
  - [ ] `DeriveIssuedPaymentAddressOutput` 直接包含 plain `SweepMaterialJSON string`。
  - [ ] `PaymentAddressAllocationStore.Complete` 只接收 `Allocation`、`SweepMaterialJSON`、`IssuedAt` 或等價的最小寫入資訊。
  - [ ] `bitcoin` 與 `ethereum` adapters 各自擁有自己的 chain-specific document assembly。
  - [ ] postgres 與 cloudflarepostgres stores 不再自己理解 chain-specific payload 細節，只寫入傳進來的 JSON。
  - [ ] generic `NewUnitOfWork(...)` 保持原本乾淨形狀，不新增 allocation-specific constructor。
  - [ ] DB 寫入的 JSON shape 與既有 payload 相容。
- Notes:
  - deriver output 與 store input 都是 application port contract；這裡接受 plain string，因為 application 的確需要把它帶到 persistence。

### FR-003 - Keep application ports boring and local to real interfaces

- Description:
  - 不得再為 `SweepMaterial` 問題在 `internal/application/ports/outbound` 發明新的 mini-model layer。
- Acceptance criteria:
  - [ ] 不新增 standalone `internal/application/ports/outbound/sweep_material.go`。
  - [ ] 不新增 `type SweepMaterial struct`、`NewBitcoinHDSweepMaterial(...)`、`ValidateSweepMaterial(...)` 這類與 real interface 脫節的 port API。
  - [ ] 若需要新的 plain input/output contract，必須 co-locate 在實際 owning port file。
- Notes:
  - 這份 requirement 不是反對 struct contract，而是反對「沒有 interface 也沒有明確 boundary ownership 的自創模型檔」。

### FR-004 - Keep technical assembly local and simple

- Description:
  - chain-specific document assembly 必須留在對應 adapter，且不新增多餘 collaborator、dispatcher、或 bootstrap helper。
- Acceptance criteria:
  - [ ] 不新增第二套 `UnitOfWork` constructor。
  - [ ] 不新增只為組裝 sweep-material 存在的 `internal/bootstrap/*.go` helper。
  - [ ] 若 deriver 無法產生 document，use case 仍沿既有 dependency failure 路徑失敗，不寫入半套 issued update。
- Notes:
  - 重點是 ownership 與 wiring 簡潔度，不是再發明一套新的抽象層。

### FR-005 - Preserve current storage and operator compatibility

- Description:
  - 重構後不得要求 schema migration，也不得破壞既有 operator JSON consumer。
- Acceptance criteria:
  - [ ] `address_policy_allocations.sweep_material_json` column 保持存在，schema 不變。
  - [ ] bitcoin 與 ethereum create2 的 JSON field names、material type、material version 與既有格式保持相容。
  - [ ] 既有與 `sweep_material_json` 相關的 persistence tests 能以新 ownership 繼續驗證相同 payload。
- Notes:
  - 行為目標是 ownership 改正，不是 payload redesign。

## Non-functional requirements

- Performance (NFR-001):
  - `Complete` 新增的 document assembly 不得引入額外 network/DB round trip；persistence builder 應在本地記憶體完成。
- Availability/Reliability (NFR-002):
  - allocation `Complete` 與 `sweep_material_json` 寫入仍須保持單一 transaction/update path，不得產生 issued state 與 payload 脫鉤。
- Security/Privacy (NFR-003):
  - 不新增額外 secret exposure；重構後寫入的 payload 內容不得超出現有欄位集合。
- Compliance (NFR-004):
  - 不適用，無新增法規需求。
- Observability (NFR-005):
  - 若 persistence builder 失敗，錯誤必須沿既有 `outport.Err...` 路徑回傳，不得 silent fallback 成空 JSON 或部分 payload。
- Maintainability (NFR-006):
  - 不新增新的 application model bucket 或 generic serialization framework；chain-specific operator document assembly 必須留在對應 chain adapter。
- Maintainability (NFR-007):
  - 不新增第二套 UnitOfWork constructor 或只為集中組裝 builder 而存在的 `internal/bootstrap/*.go` helper。

## Dependencies and integrations

- External systems:
  - PostgreSQL / Cloudflare Postgres schema `address_policy_allocations.sweep_material_json`
  - `internal/infrastructure/ethereumcreate2assets`
- Internal services:
  - `IssuedPaymentAddressDeriver`
  - `PaymentAddressAllocationStore`
  - `AllocatePaymentAddressUseCase`
