---
doc: 03_tasks
spec_date: 2026-04-03
slug: notification-delivery-boundary
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-02-domain-model-boundary-cleanup
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
- Rationale:
  - 這是小範圍的 type ownership 重分類，沒有 schema 變更、沒有新外部整合，也不需要額外 design doc 才能實作。
- Upstream dependencies (`depends_on`):
  - `2026-04-02-domain-model-boundary-cleanup`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - 變更只涉及 package ownership 與 call site 更新，沒有新的 async flow、schema、或 integration design。
  - What would trigger switching to Full mode:
    - 若實作過程發現必須改 schema、改 outbox payload、或重寫 webhook dispatch flow。
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - 不適用，本 spec 保留 test plan。

## Milestones

- M1:
  - 完成 spec 並確認 ownership 決策。
- M2:
  - 完成 type 搬移與 call site 更新。
- M3:
  - 完成 regression 驗證與 spec 收尾。

## Tasks (ordered)

1. T-001 - Move notification delivery workflow types to application ownership
   - Scope:
     - 將 delivery status / failure reason type 與 parse/helper API 從 `internal/domain/valueobjects` 搬到 application/outbox 邊界。
   - Output:
     - 更新後的 type 定義與 package 位置。
   - Linked requirements: FR-001, FR-003, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/application/outbox ./internal/domain/...`
     - [ ] Expected result: domain 不再匯出 notification delivery workflow type，application/outbox tests 通過。
     - [ ] Logs/metrics to check (if applicable): 無。
2. T-002 - Update use cases, persistence adapters, and tests
   - Scope:
     - 更新 webhook dispatch use case、postgres/cloudflarepostgres outbox stores、以及相關測試的 import 與型別引用。
   - Output:
     - 更新後的 call sites 與 regression tests。
   - Linked requirements: FR-002, FR-003, NFR-002, NFR-005
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/application/... ./internal/adapters/outbound/persistence/...`
     - [ ] Expected result: sent / pending / failed workflow 行為與 persisted values 維持不變。
     - [ ] Logs/metrics to check (if applicable): 無。
3. T-003 - Run full verification and close the spec
   - Scope:
     - 跑 compile/test/spec lint，確認這輪 boundary cleanup 沒有 regression。
   - Output:
     - 驗證結果與最終 spec 狀態更新。
   - Linked requirements: FR-002, NFR-001, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go list ./... && go test ./... && SPEC_DIR="specs/2026-04-03-notification-delivery-boundary" bash scripts/spec-lint.sh`
     - [ ] Expected result: 全部通過。
     - [ ] Logs/metrics to check (if applicable): 無。

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-002, T-003
- FR-003 -> T-001, T-002
- NFR-001 -> T-003
- NFR-002 -> T-002, T-003
- NFR-005 -> T-002
- NFR-006 -> T-001, T-003

## Rollout and rollback

- Feature flag:
  - 不適用，這是 internal refactor。
- Migration sequencing:
  - 無 migration。
- Rollback steps:
  - 若實作產生不必要 churn，可回退這輪 type 搬移；schema 與 persisted values 不受影響。

## Completion

- Completed on:
  - 2026-04-03
- Outcome:
  - notification delivery workflow status / failure reason 已從 `internal/domain/valueobjects` 移到 `internal/application/outbox`，相關 use case、outbox store、與測試皆已對齊。
- Validation evidence:
  - `go list ./...`
  - `go test ./internal/application/outbox ./internal/application/usecases ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres ./internal/domain/...`
  - `go test ./...`
  - `SPEC_DIR="specs/2026-04-03-notification-delivery-boundary" bash scripts/spec-lint.sh`
  - `bash scripts/precommit-run.sh`
