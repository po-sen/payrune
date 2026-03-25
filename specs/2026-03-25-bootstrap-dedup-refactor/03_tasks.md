---
doc: 03_tasks
spec_date: 2026-03-25
slug: bootstrap-dedup-refactor
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-24-architecture-conformance-refactor
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
- Rationale: 這輪只重構 `internal/bootstrap` 內既有 parsing / builder 重複，沒有新 integration、schema 或新的 failure-mode class，但仍需要 test plan 鎖住行為。
- Upstream dependencies (`depends_on`): `2026-03-24-architecture-conformance-refactor`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: 不改變架構邊界與 runtime flow，只在既有 bootstrap ownership 內做 dedup。
  - What would trigger switching to Full mode: 若實作中發現需要新增新的 bootstrap subpackage 或跨層 contract 變更。
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): 不適用，本 spec 產出 test plan。

## Milestones

- M1: 建立 dedup spec 並鎖定可安全共用的重複範圍。
- M2: 完成 API、poller、dispatcher 三組共通 parsing/builder 收斂。
- M3: 測試、lint、spec 收尾。

## Tasks (ordered)

1. T-001 - title
   - Scope: 建立並填寫 Quick-mode spec，限制本輪只在 `internal/bootstrap` 內做 dedup，不動 runtime ownership。
   - Output: `specs/2026-03-25-bootstrap-dedup-refactor/`。
   - Linked requirements: FR-004, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-25-bootstrap-dedup-refactor" bash scripts/spec-lint.sh`
     - [ ] Expected result: spec-lint 通過。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-002 - Refactor API receipt terms parsing
   - Scope: 將 API process / worker 兩側的 receipt confirmations 與 expires-after parsing 收成共用 helper，保留各自 default ownership。
   - Output: 更新後的 `internal/bootstrap/api.go`、`internal/bootstrap/api_worker.go` 與相關測試。
   - Linked requirements: FR-001, FR-004, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/bootstrap -run 'TestLoadReceipt|TestExecuteAPIWorkerRequest'`
     - [ ] Expected result: API process / worker 相關測試通過。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-003 - Refactor poller dispatch parsing
   - Scope: 將 poller process config 與 worker request 的共通 dispatch parsing 收成 lookup-based helper，保留 process-only `POLL_TICK_INTERVAL`。
   - Output: 更新後的 `internal/bootstrap/poller.go`、`internal/bootstrap/poller_worker.go` 與相關測試。
   - Linked requirements: FR-002, FR-004, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/bootstrap -run 'TestLoadPoller|TestBuildCloudflarePollerRequest|TestHandleCloudflarePollerRequestJSON'`
     - [ ] Expected result: poller process / worker 相關測試通過。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-004 - Refactor receipt webhook dispatcher parsing
   - Scope: 將 dispatcher process config、worker request 與 notifier 共通欄位 parsing 收成共享 helper，保留 runtime target 差異。
   - Output: 更新後的 `internal/bootstrap/receipt_webhook_dispatcher.go`、`internal/bootstrap/receipt_webhook_dispatcher_worker.go` 與相關測試。
   - Linked requirements: FR-003, FR-004, NFR-002, NFR-003, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/bootstrap -run 'TestLoadReceiptWebhookDispatcher|TestLoadPaymentReceiptWebhookNotifier|TestBuildCloudflareReceiptWebhookDispatcherRequest|TestHandleCloudflareReceiptWebhookDispatcherRequestJSON|TestLoadCloudflareReceiptWebhookNotifierConfig'`
     - [ ] Expected result: dispatcher process / worker 相關測試通過。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-005 - Final verification and spec closeout
   - Scope: 執行 bootstrap 與 repo 驗證，更新 spec 最終狀態。
   - Output: 測試結果與 DONE 狀態 frontmatter。
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, NFR-001, NFR-002, NFR-003, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/bootstrap/...`、`go test ./...`、`SPEC_DIR="specs/2026-03-25-bootstrap-dedup-refactor" bash scripts/spec-lint.sh`
     - [ ] Expected result: 所有驗證通過，spec 狀態一致。
     - [ ] Logs/metrics to check (if applicable): 無。

## Traceability (optional)

- FR-001 -> T-002, T-005
- FR-002 -> T-003, T-005
- FR-003 -> T-004, T-005
- FR-004 -> T-001, T-002, T-003, T-004, T-005
- NFR-001 -> T-005
- NFR-002 -> T-002, T-003, T-004, T-005
- NFR-003 -> T-004, T-005
- NFR-006 -> T-001, T-002, T-003, T-004, T-005

## Rollout and rollback

- Feature flag: 無。
- Migration sequencing: 先補 spec，接著逐組抽 shared helper 並修測試，最後跑 bootstrap/full test 與 spec-lint。
- Rollback steps: revert 本輪 bootstrap helper 抽取與測試調整；不涉及 schema 或 external contract 變更。
