---
doc: 03_tasks
spec_date: 2026-03-27
slug: domain-error-contracts
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-application-inbound-error-mapping
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
- Rationale: 這輪只抽 domain error contract 與調整測試，沒有新 integration、schema、或 runtime design 變更。
- Upstream dependencies (`depends_on`):
  - `2026-03-27-application-inbound-error-mapping`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: 這是 package-local contract cleanup，不涉及新流程或部署設計。
  - What would trigger switching to Full mode: 若需要全面重設 domain model hierarchy 或引入新的 shared error framework。
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): 不適用；本 spec 會產出 `04_test_plan.md`。

## Milestones

- M1: 盤點並定義 domain sentinel errors。
- M2: 完成 domain/application 回歸測試與驗證。

## Tasks (ordered)

1. T-001 - Define package-owned domain sentinel errors
   - Scope: 在 `internal/domain/entities`, `internal/domain/valueobjects`, `internal/domain/events`, `internal/domain/policies` 中，將匿名且具跨層價值的 invariant error 改成 stable sentinel。
   - Output: package-owned domain error definitions 與對應 implementation。
   - Linked requirements: FR-001, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `rg -n "errors\\.New\\(" internal/domain --glob '*.go'`
     - [ ] Expected result: 原本的匿名 invariant error 顯著減少，並可看到具名 sentinel 定義。
     - [ ] Logs/metrics to check (if applicable): 不適用。
2. T-002 - Update domain and usecase tests for stable errors
   - Scope: 調整受影響的 domain tests，必要時補 application/usecase regression，讓 callers 可以用 `errors.Is(...)` 判斷新的 contract。
   - Output: updated tests with stable error assertions。
   - Linked requirements: FR-002, FR-003, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/domain/... ./internal/application/usecases`
     - [ ] Expected result: domain 與 usecase tests 通過。
     - [ ] Logs/metrics to check (if applicable): 不適用。
3. T-003 - Run repo validation
   - Scope: 執行 spec lint 與全 repo Go tests，確認這輪是純 contract cleanup。
   - Output: validation evidence。
   - Linked requirements: FR-003, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-27-domain-error-contracts" bash scripts/spec-lint.sh && go test ./...`
     - [ ] Expected result: spec lint 與全 repo 測試通過。
     - [ ] Logs/metrics to check (if applicable): 不適用。

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-002
- FR-003 -> T-002, T-003
- NFR-002 -> T-002, T-003
- NFR-006 -> T-001, T-003

## Rollout and rollback

- Feature flag: 無。
- Migration sequencing: 先定義 sentinel errors，再改呼叫點與測試，最後跑 validation。
- Rollback steps: 還原本次 commit，恢復匿名 `errors.New(...)` contract。

## Validation evidence

- `SPEC_DIR="specs/2026-03-27-domain-error-contracts" bash scripts/spec-lint.sh`
- `go test ./internal/domain/... ./internal/application/usecases`
- `go test ./...`
