---
doc: 03_tasks
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

# Task Plan

## Mode decision

- Selected mode: Full
- Rationale:
  - 這次不是單一 rename，而是要重設 domain/application/adapters 之間的 ownership，並調整 persistence write contract 與 failure handling。
- Upstream dependencies (`depends_on`):
  - `2026-04-02-domain-model-boundary-cleanup`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: 不適用，本 spec 採 Full mode。
  - What would trigger switching to Full mode: 不適用。
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): 不適用，本 spec 產出 test plan。

## Milestones

- M1:
  - 完成新 spec，明確定義 ownership 與 chosen design。
- M2:
  - 完成 domain / port contract 清理。
- M3:
  - 完成簡化後的 deriver/store 改寫，不新增額外 UoW 或 wiring。
- M4:
  - 完成 regression tests 與 spec 收尾。

## Tasks (ordered)

1. T-001 - Lock the new sweep-material redesign spec
   - Scope:
     - 建立獨立 Full-mode spec，將問題收斂在 `sweep_material_json` 的 ownership、contract 與 persistence assembly。
   - Output:
     - `specs/2026-04-02-sweep-material-redesign/` 五份文件。
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `SPEC_DIR="specs/2026-04-02-sweep-material-redesign" bash scripts/spec-lint.sh`
     - [ ] Expected result: spec-lint 通過。
     - [ ] Logs/metrics to check (if applicable): 無。
2. T-002 - Remove sweep-material JSON from domain and deriver port
   - Scope:
     - 從 `PaymentAddressAllocation` 拿掉 `SweepMaterialJSON`。
     - 調整 `MarkIssued(...)`、use case、測試與相鄰 call sites。
   - Output:
     - 更新後的 entity、deriver port、use case 與測試。
   - Linked requirements: FR-001, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/domain/entities ./internal/application/usecases ./internal/adapters/outbound/blockchain`
     - [ ] Expected result: domain 與 use case tests 通過，runtime code 不再在 domain/entity 持有 `SweepMaterialJSON`。
     - [ ] Logs/metrics to check (if applicable): 無。
3. T-003 - Redesign allocation completion contract around issued metadata
   - Scope:
     - 將 `PaymentAddressAllocationStore.Complete` 改成 plain input struct。
     - input 只攜帶 issued allocation、`SweepMaterialJSON`、issuedAt。
     - contract co-locate 在 store interface file，不新增 standalone mini-model 檔。
   - Output:
     - 更新後的 outbound store port、use case call site、相關 fake/test helpers。
   - Linked requirements: FR-002, FR-003, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/application/usecases ./internal/application/ports/outbound`
     - [ ] Expected result: `Complete` 走新的 plain input contract，`internal/application/ports/outbound` 不新增 `SweepMaterial` standalone model API。
     - [ ] Logs/metrics to check (if applicable): 無。
4. T-004 - Move document assembly into persistence adapters
   - Scope:
     - 在 `bitcoin` / `ethereum` issued deriver 內建立 chain-specific JSON 組裝。
     - use case 將 `SweepMaterialJSON` 原樣傳給 store。
     - postgres / cloudflarepostgres 只寫入既有 JSON payload，不新增特殊 UoW/wiring。
   - Output:
     - 更新後的 chain deriver、store contract、stores 與測試。
   - Linked requirements: FR-002, FR-004, FR-005, NFR-001, NFR-002, NFR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/adapters/outbound/bitcoin ./internal/adapters/outbound/ethereum ./internal/adapters/outbound/blockchain ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres`
     - [ ] Expected result: stores 仍寫出相同 JSON payload；deriver 失敗時 issued flow 失敗；不新增第二套 `UnitOfWork` 或 bootstrap helper。
     - [ ] Logs/metrics to check (if applicable): 無。
5. T-005 - Final regression verification and spec closeout
   - Scope:
     - 跑 compile/test/spec lint，確認新的 ownership 沒有造成 regression。
   - Output:
     - 完整驗證結果與更新後的 spec 狀態。
   - Linked requirements: FR-005, NFR-001, NFR-002, NFR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go list ./... && go test ./... && SPEC_DIR="specs/2026-04-02-sweep-material-redesign" bash scripts/spec-lint.sh`
     - [ ] Expected result: 全部通過。
     - [ ] Logs/metrics to check (if applicable): 無。

## Traceability (optional)

- FR-001 -> T-001, T-002
- FR-002 -> T-001, T-003, T-004
- FR-003 -> T-001, T-003
- FR-004 -> T-001, T-004
- FR-005 -> T-001, T-004, T-005
- NFR-001 -> T-004, T-005
- NFR-002 -> T-003, T-004, T-005
- NFR-005 -> T-004, T-005
- NFR-006 -> T-001, T-002, T-003, T-004, T-005

## Rollout and rollback

- Feature flag:
  - 不適用，本輪為 internal refactor。
- Migration sequencing:
  - 無 DB migration。
- Rollback steps:
  - 若新 contract 導致 churn 過大，可先只保留 spec，不提交 code；已改動的 runtime contract 可完整回退，因 schema 與 external API 不變。

## Completion

- Completed on:
  - 2026-04-03
- Outcome:
  - `sweep_material_json` 已從 domain entity 移除，改由 bitcoin / ethereum issued deriver 直接產出，use case 原樣傳給既有 store 寫入。
- Validation evidence:
  - `go list ./...`
  - `go test ./...`
  - `SPEC_DIR="specs/2026-04-02-sweep-material-redesign" bash scripts/spec-lint.sh`
  - `bash scripts/precommit-run.sh`
