---
doc: 03_tasks
spec_date: 2026-03-27
slug: application-error-boundaries
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-26-allocate-usecase-decomposition
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale: 這輪是 application error ownership cleanup，範圍集中在 `inport` / `outport` / `usecases` / 對應 tests；不新增 integration、schema、或 rollout behavior，但需要 test plan 鎖住 error taxonomy 和既有 mapping。
- Upstream dependencies (`depends_on`): `2026-03-26-allocate-usecase-decomposition`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: 不改 runtime architecture，只整理 application contract error ownership。
  - What would trigger switching to Full mode: 若實作中需要改 outward API contract、domain error model、或引入跨 package redesign。
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): 不適用，本 spec 產出 test plan。

## Milestones

- M1: 建立 application error ownership spec。
- M2: 集中 inbound application errors。
- M3: 改寫 usecases / tests / adapter mappings 使用 shared error。
- M4: 將多實作 outbound port contract error 升成 `outport.Err...`。
- M5: 將 error ownership 規則寫入 `AGENTS.md`。
- M6: 驗證與 spec closeout。

## Tasks (ordered)

1. T-001 - Define application error ownership
   - Scope: 建立 `inport` / `outport` error taxonomy，明確哪些 error 屬於 usecase 對外 contract，哪些維持 outbound adapter contract。
   - Output: 集中的 error definition 與對應 spec。
   - Linked requirements: FR-001, FR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-27-application-error-boundaries" bash scripts/spec-lint.sh`
     - [ ] Expected result: spec-lint 通過。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-002 - Replace ad-hoc usecase errors with inbound shared errors
   - Scope: 為 usecase 目前直接建立的 configuration / validation / consistency error 建立 shared inbound error，並更新 usecase source。
   - Output: `internal/application/usecases` 不再直接建立 shared application error。
   - Linked requirements: FR-001, FR-003, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/application/usecases`
     - [ ] Expected result: usecase package 測試通過。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-003 - Keep adapter mapping and outbound branching stable
   - Scope: 保留 controller 對 `inport.Err...` 的 mapping、以及 usecase 對 `outport.Err...` 的 branching；更新 tests 以 shared error 驗證。
   - Output: 穩定的 application / adapter error taxonomy。
   - Linked requirements: FR-002, FR-004, NFR-002, NFR-003, NFR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/adapters/inbound/http/controllers ./internal/application/usecases`
     - [ ] Expected result: controller mapping 與 usecase error 行為維持不變。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-004 - Promote shared outbound contract errors
   - Scope: 將多實作 persistence adapters 共享的 contract/validation/state error 收斂到 `internal/application/ports/outbound`，包含 idempotency、receipt tracking、notification outbox、payment address allocation、payment address status finder，並更新 postgres / cloudflarepostgres implementations 與 tests。
   - Output: 不再依賴重複字串的 outbound port contract errors。
   - Linked requirements: FR-002, FR-003, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres`
     - [ ] Expected result: 兩組 adapter 測試都通過，shared port errors 集中於 `outport`。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-005 - Document error ownership rules in AGENTS
   - Scope: 將 domain / inbound application / outbound port / adapter-private error 的 ownership 規則寫入 repo-specific architecture guidance。
   - Output: `AGENTS.md` 有明確的 error ownership contract。
   - Linked requirements: FR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `rg -n "Error ownership|adapter-private" AGENTS.md`
     - [ ] Expected result: `AGENTS.md` 可直接讀到 error ownership 規則。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-006 - Final verification and spec closeout
   - Scope: 跑全 repo 驗證並更新 spec 狀態。
   - Output: 測試結果與 DONE frontmatter。
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, NFR-001, NFR-002, NFR-003, NFR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./...`、`SPEC_DIR="specs/2026-03-27-application-error-boundaries" bash scripts/spec-lint.sh`
     - [ ] Expected result: 所有驗證通過，spec 狀態一致。
     - [ ] Logs/metrics to check (if applicable): 無。

## Traceability (optional)

- FR-001 -> T-001, T-002, T-005
- FR-002 -> T-001, T-003, T-004, T-005
- FR-003 -> T-004, T-005
- FR-004 -> T-002, T-003, T-005
- FR-005 -> T-005, T-006
- NFR-001 -> T-006
- NFR-002 -> T-002, T-003, T-004, T-006
- NFR-003 -> T-003, T-006
- NFR-005 -> T-003, T-006
- NFR-006 -> T-001, T-002, T-003, T-004, T-005, T-006

## Rollout and rollback

- Feature flag: 無。
- Migration sequencing: 先定義 `inport` / `outport` error taxonomy，再改 usecases/tests，最後收多實作 outbound port contract errors 並跑全量驗證。
- Rollback steps: revert 本輪 `inport` / `usecases` / tests / spec 變更；不涉及 schema 或 runtime config。
