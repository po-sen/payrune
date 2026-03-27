---
doc: 03_tasks
spec_date: 2026-03-25
slug: usecase-boundary-cleanup
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-24-architecture-conformance-refactor
  - 2026-03-25-bootstrap-dedup-refactor
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
- Rationale: 這輪是 application / adapter / bootstrap 的 boundary refactor，沒有新增 integration、schema 或 rollout class，但需要 test plan 鎖住既有行為與新邊界。
- Upstream dependencies (`depends_on`): `2026-03-24-architecture-conformance-refactor`, `2026-03-25-bootstrap-dedup-refactor`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: 不改外部 contract 或資料模型，只在既有層內移動責任與調整 port。
  - What would trigger switching to Full mode: 若實作中需要新增獨立 runtime flow、持久化模型、或第二種 issued-address derivation backend。
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): 不適用，本 spec 產出 test plan。

## Milestones

- M1: 建立 usecase boundary cleanup spec 並鎖定 in-scope / out-of-scope。
- M2: 完成 allocation derivation 與 preview capability 的邊界搬移。
- M3: 完成 poller typed scope 遷移、測試與 spec 收尾。

## Tasks (ordered)

1. T-001 - Scaffold and validate refactor spec
   - Scope: 建立 Quick-mode spec，明確限制本輪聚焦在 `internal/application/usecases` 的 purity cleanup。
   - Output: `specs/2026-03-25-usecase-boundary-cleanup/`。
   - Linked requirements: FR-004, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-25-usecase-boundary-cleanup" bash scripts/spec-lint.sh`
     - [ ] Expected result: spec-lint 通過。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-002 - Refactor allocation address derivation ownership
   - Scope: 將 `allocate` usecase 的 issued-address derivation 細節收斂到單一 outbound port / adapter，移除 usecase 內的 `create2` 技術分支。
   - Output: 更新後的 application port、adapter implementation、bootstrap wiring、usecase 與測試。
   - Linked requirements: FR-001, FR-004, NFR-001, NFR-002, NFR-003, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/application/usecases -run 'TestAllocatePaymentAddressUseCase'`
     - [ ] Expected result: allocate usecase 相關測試通過，含 Ethereum create2 success / failure path。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-003 - Move preview capability into domain policy
   - Scope: 將 address preview support rule 收回 domain entity / policy，讓 `generate` usecase 只做 orchestration 與 error mapping。
   - Output: 更新後的 domain entity、generate usecase 與測試。
   - Linked requirements: FR-002, FR-004, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/application/usecases -run 'TestGenerateAddressUseCase'`
     - [ ] Expected result: generate usecase 相關測試通過，preview unsupported 行為維持不變。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-004 - Normalize poller scope before usecase
   - Scope: 將 poller `chain/network` filter 提前正規化為 typed value，移除 usecase 中的 raw parsing helper，並同步更新 bootstrap / scheduler call path。
   - Output: 更新後的 DTO、scheduler adapter、bootstrap poller path、poller usecase 與測試。
   - Linked requirements: FR-003, FR-004, NFR-002, NFR-003, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/application/usecases -run 'TestRunReceiptPollingCycleUseCase' && go test ./internal/bootstrap -run 'TestLoadPoller|TestBuildCloudflarePollerRequest|TestHandleCloudflarePollerRequestJSON'`
     - [ ] Expected result: poller usecase 與 bootstrap path 測試通過，typed filter 行為與既有 validation 一致。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-005 - Final verification and spec closeout
   - Scope: 執行 package / repo 驗證，更新 spec 最終狀態。
   - Output: 測試結果與 DONE 狀態 frontmatter。
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, NFR-001, NFR-002, NFR-003, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/application/...`、`go test ./internal/bootstrap/...`、`go test ./...`、`SPEC_DIR="specs/2026-03-25-usecase-boundary-cleanup" bash scripts/spec-lint.sh`
     - [ ] Expected result: 所有驗證通過，spec 狀態一致。
     - [ ] Logs/metrics to check (if applicable): 無。

## Traceability (optional)

- FR-001 -> T-002, T-005
- FR-002 -> T-003, T-005
- FR-003 -> T-004, T-005
- FR-004 -> T-001, T-002, T-003, T-004, T-005
- NFR-001 -> T-002, T-005
- NFR-002 -> T-002, T-003, T-004, T-005
- NFR-003 -> T-002, T-004, T-005
- NFR-006 -> T-001, T-002, T-003, T-004, T-005

## Rollout and rollback

- Feature flag: 無。
- Migration sequencing: 先補 spec，再調整 allocation derivation port、preview rule、poller typed filter，最後跑全量測試與 spec-lint。
- Rollback steps: revert 本輪 application port / adapter / bootstrap / usecase 變更；不涉及 schema 或 data migration。
