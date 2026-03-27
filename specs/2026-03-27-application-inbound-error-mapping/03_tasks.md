---
doc: 03_tasks
spec_date: 2026-03-27
slug: application-inbound-error-mapping
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-application-error-boundaries
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
- Rationale: 這是 usecase contract cleanup，沒有新 integration、schema、或 persistent model；主要工作是 error mapping 與測試回歸。
- Upstream dependencies (`depends_on`):
  - `2026-03-27-application-error-boundaries`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: 沒有新增 async flow 或 schema；設計決策集中在 error ownership 與 usecase return contract。
  - What would trigger switching to Full mode: 若需要新增 cross-layer logging/telemetry protocol、worker error envelope redesign、或 broader adapter contract redesign。
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): 不適用；本 spec 會產出 `04_test_plan.md`。

## Milestones

- M1: 定義 generic inbound error 與 mapping 規則。
- M2: 使所有 usecase 只輸出 inbound contract error，並完成回歸測試。

## Tasks (ordered)

1. T-001 - Define generic inbound error contract
   - Scope: 在 `internal/application/ports/inbound/errors.go` 補齊 generic inbound error，區分 dependency failure 與 internal failure。
   - Output: 新的 inbound generic error 與對應命名規則。
   - Linked requirements: FR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `sed -n '1,220p' internal/application/ports/inbound/errors.go`
     - [ ] Expected result: 可看到清楚的 generic inbound error，且不與既有 business/config error 衝突。
     - [ ] Logs/metrics to check (if applicable): 不適用。
2. T-002 - Map unexpected usecase failures to inbound errors
   - Scope: 調整 `internal/application/usecases`，保留既有 known mapping，並把 unexpected outbound/private/internal error 收斂到 generic inbound error。
   - Output: updated usecases，不再讓 raw `outport`/adapter/private error 離開 application layer。
   - Linked requirements: FR-001, FR-002, NFR-003, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/application/usecases`
     - [ ] Expected result: usecase tests 通過，包含 generic inbound error regression。
     - [ ] Logs/metrics to check (if applicable): 不適用。
3. T-003 - Refresh inbound adapter regression coverage
   - Scope: 更新 controller / scheduler 相關測試，確認 application 改成 generic inbound error 後，transport 行為仍維持既有 contract。
   - Output: updated inbound adapter tests 與綠燈驗證。
   - Linked requirements: FR-003, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/adapters/inbound/... && go test ./...`
     - [ ] Expected result: inbound adapter tests 與全 repo 測試通過。
     - [ ] Logs/metrics to check (if applicable): 不適用。

## Traceability (optional)

- FR-001 -> T-002
- FR-002 -> T-001, T-002
- FR-003 -> T-003
- NFR-002 -> T-003
- NFR-003 -> T-002
- NFR-006 -> T-001, T-002

## Rollout and rollback

- Feature flag: 無。
- Migration sequencing: 先定義 inbound generic error，再改 usecase，最後更新測試。
- Rollback steps: 還原本次 commit，重新允許 raw error 穿過 usecase 邊界。

## Validation evidence

- `go test ./internal/application/usecases`
- `go test ./internal/adapters/inbound/...`
- `go test ./...`
- `SPEC_DIR="specs/2026-03-27-application-inbound-error-mapping" bash scripts/spec-lint.sh`
- `bash scripts/precommit-run.sh`
