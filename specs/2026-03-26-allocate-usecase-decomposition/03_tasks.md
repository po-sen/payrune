---
doc: 03_tasks
spec_date: 2026-03-26
slug: allocate-usecase-decomposition
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-25-usecase-boundary-cleanup
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
- Rationale: 這輪持續重整 `internal/application/usecases` 的 readability，且使用者明確要求以「更好讀」而不是「拆更多檔案」為優先；最後一輪還要把 usecase 內殘留的 transport normalization 與 runtime default ownership 收回正確層。沒有新增 integration、schema、或 rollout class，但需要 test plan 鎖住 flow semantics。
- Upstream dependencies (`depends_on`): `2026-03-25-usecase-boundary-cleanup`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: 不改 architecture boundary 或外部 contract，只調整同 package internal collaborator ownership。
  - What would trigger switching to Full mode: 若實作中需要新增新的 package boundary、repository port、或 rollout behavior。
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): 不適用，本 spec 產出 test plan。

## Milestones

- M1: 建立並維護同一份 readability-first spec。
- M2: 完成 allocation 與 generate flow 的可讀性重整，減少不必要的 top-level helper 與跳轉。
- M3: 完成 receipt polling cycle 的重複 error/save path 收斂。
- M4: 完成 receipt webhook dispatch cycle 的重複 delivery-result path 收斂。
- M5: 完成 usecase 最後一輪 boundary ownership cleanup。
- M6: 測試、spec lint、spec closeout。

## Tasks (ordered)

1. T-001 - Maintain one readability spec
   - Scope: 維護單一 Quick-mode spec，集中記錄本輪 usecase readability cleanup，不再新增碎片 spec。
   - Output: `specs/2026-03-26-allocate-usecase-decomposition/`。
   - Linked requirements: FR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-26-allocate-usecase-decomposition" bash scripts/spec-lint.sh`
     - [ ] Expected result: spec-lint 通過。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-002 - Simplify replay handling
   - Scope: 以可讀性為優先整理 replay lookup / conflict consistency / replay transaction scope，必要時將 replay path 收回 `Execute` 附近。
   - Output: 較少跳轉、較易理解的 replay path。
   - Linked requirements: FR-001, FR-002, FR-005, NFR-002, NFR-003, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/application/usecases -run 'TestAllocatePaymentAddressUseCase(ReturnsExistingIssuedAllocation|RejectsIdempotencyKeyConflict|AllowsRetryWhenClaimConflictHasNoCompletedRecord)'`
     - [ ] Expected result: replay / idempotency 相關測試通過。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-003 - Simplify issuance and response flow
   - Scope: 以線性可讀為優先整理 reservation / derivation / failure persistence / tracking creation / idempotency completion、response mapping、以及 `issueAllocation` 後的 fallback/error 分支；避免保留單次使用的薄 helper。
   - Output: 較少概念切換、較好順讀的 allocation flow。
   - Linked requirements: FR-001, FR-003, FR-004, FR-005, NFR-001, NFR-002, NFR-003, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/application/usecases -run 'TestAllocatePaymentAddressUseCase'`
     - [ ] Expected result: allocate usecase 全量測試通過。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-004 - Inline generate preview validation helper
   - Scope: 移除 `GenerateAddressUseCase` 的 `validateGenerateAddressPolicy`，將 preview validation 與 error mapping 收回 `Execute`。
   - Output: 更順讀的 `generate_address_use_case.go`。
   - Linked requirements: FR-006, FR-007, NFR-001, NFR-002, NFR-003, NFR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/application/usecases -run 'TestGenerateAddressUseCase'`
     - [ ] Expected result: generate usecase 相關測試通過。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-005 - Simplify receipt polling cycle flow
   - Scope: 收斂 `RunReceiptPollingCycleUseCase` 內重複的 polling-error save path 與 save+enqueue path，讓 main loop 保留高階流程。
   - Output: 更容易順讀的 `run_receipt_polling_cycle_use_case.go`。
   - Linked requirements: FR-008, NFR-001, NFR-002, NFR-003, NFR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/application/usecases -run 'TestRunReceiptPollingCycleUseCase'`
     - [ ] Expected result: receipt polling usecase 相關測試通過。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-006 - Simplify receipt webhook dispatch cycle flow
   - Scope: 收斂 `RunReceiptWebhookDispatchCycleUseCase` 內單筆 notification dispatch 流程與重複的 `SaveDeliveryResult` transaction。
   - Output: 更容易順讀的 `run_receipt_webhook_dispatch_cycle_use_case.go`。
   - Linked requirements: FR-009, NFR-001, NFR-002, NFR-003, NFR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/application/usecases -run 'TestRunReceiptWebhookDispatchCycleUseCase'`
     - [ ] Expected result: receipt webhook dispatch usecase 相關測試通過。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-007 - Remove remaining transport and runtime ownership from usecases
   - Scope: 把 `AllocatePaymentAddressUseCase` 的 input trimming 移回 inbound controller ownership，並把 polling / webhook dispatch usecase 的 runtime default ownership 移回 bootstrap / scheduler path。
   - Output: `internal/application/usecases` 不再持有 transport normalization 或 scheduler runtime default。
   - Linked requirements: FR-010, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/application/usecases -run 'TestAllocatePaymentAddressUseCase|TestRunReceiptPollingCycleUseCase|TestRunReceiptWebhookDispatchCycleUseCase'`、`go test ./internal/adapters/inbound/http/controllers ./internal/bootstrap/...`
     - [ ] Expected result: usecase semantics 維持，且 controller / bootstrap 測試證明 ownership 已回到正確層。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-008 - Final verification and spec closeout
   - Scope: 跑 repo 驗證並更新 spec 狀態。
   - Output: 測試結果與 DONE frontmatter。
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, FR-008, FR-009, FR-010, NFR-001, NFR-002, NFR-003, NFR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/application/usecases`、`go test ./...`、`SPEC_DIR="specs/2026-03-26-allocate-usecase-decomposition" bash scripts/spec-lint.sh`
     - [ ] Expected result: 所有驗證通過，spec 狀態一致。
     - [ ] Logs/metrics to check (if applicable): 無。

## Traceability (optional)

- FR-001 -> T-002, T-003, T-004
- FR-002 -> T-002, T-004
- FR-003 -> T-003, T-006
- FR-004 -> T-003, T-006
- FR-005 -> T-001, T-002, T-003, T-004, T-005, T-006
- FR-006 -> T-004, T-006
- FR-007 -> T-004, T-006
- FR-008 -> T-005, T-008
- FR-009 -> T-006, T-008
- FR-010 -> T-007, T-008
- NFR-001 -> T-003, T-004, T-005, T-006, T-008
- NFR-002 -> T-002, T-003, T-004, T-005, T-006, T-007, T-008
- NFR-003 -> T-002, T-003, T-004, T-005, T-006, T-008
- NFR-005 -> T-004, T-005, T-006, T-008
- NFR-006 -> T-001, T-002, T-003, T-004, T-005, T-006, T-007, T-008

## Rollout and rollback

- Feature flag: 無。
- Migration sequencing: 維持同一份 spec，依序收斂 allocation、generate、receipt polling 的 readability 問題，最後跑 tests / spec-lint。
- Rollback steps: revert 本輪 `internal/application/usecases` 與 spec 變更；不涉及 schema、config、或 rollout 變更。
