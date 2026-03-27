---
doc: 03_tasks
spec_date: 2026-03-28
slug: process-error-reason-cleanup
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-application-error-boundaries
  - 2026-03-27-domain-error-contracts
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale: 只做 receipt process reason 的 model refactor，不新增 integration 或 schema migration。
- Upstream dependencies (`depends_on`): `2026-03-27-application-error-boundaries`, `2026-03-27-domain-error-contracts`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: 沒有新 integration 或 migration，只有既有 receipt process reason 的型別化與 mapping。
  - What would trigger switching to Full mode: 若要改 DB schema、重做 aggregate 邊界、或調整 outward API contract。
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): 每個 task 都附具體 `go test` / grep / spec-lint 驗證。

## Milestones

- M1: Introduce domain typed reason codes for tracking and webhook delivery.
- M2: Adapt persistence/read-side mapping and close with full validation.

## Tasks (ordered)

1. T-001 - Add domain typed reason codes for receipt tracking and webhook delivery
   - Scope: 在 `internal/domain/valueobjects` 定義 reason code，並調整 entity / policy 改吃 typed reason。
   - Output: receipt tracking and webhook delivery domain model use typed reason values instead of free-form strings.
   - Linked requirements: FR-001 / FR-002 / NFR-002 / NFR-005 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./internal/domain/... ./internal/application/usecases -run 'TestRunReceiptPollingCycleUseCase|TestRunReceiptWebhookDispatchCycleUseCase'`
     - [x] Expected result: domain and usecase tests pass with typed reasons.
     - [x] Logs/metrics to check (if applicable): N/A
2. T-002 - Adapt persistence and read-side mapping to typed reasons
   - Scope: 調整 receipt tracking store, status finder, webhook outbox store, and status read usecase to serialize/parse typed reasons and expose public text.
   - Output: internal persistence/query models use typed reason values while outward DTO keeps readable text.
   - Linked requirements: FR-003 / NFR-002 / NFR-003 / NFR-005 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres ./internal/application/usecases`
     - [x] Expected result: stores and read-side tests pass with typed reason serialization/parsing.
     - [x] Logs/metrics to check (if applicable): N/A
3. T-003 - Run full validation and close the spec
   - Scope: 跑 full test/spec lint，並把 spec 收回 `DONE`。
   - Output: verified cleanup with final spec evidence.
   - Linked requirements: FR-001 / FR-002 / FR-003 / NFR-001 / NFR-002 / NFR-003 / NFR-005 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./...`, `SPEC_DIR="specs/2026-03-28-process-error-reason-cleanup" bash scripts/spec-lint.sh`, `bash scripts/precommit-run.sh`
     - [x] Expected result: full suite passes and spec evidence reflects the final typed-reason model.
     - [x] Logs/metrics to check (if applicable): N/A

## Traceability (optional)

- FR-001 -> T-001, T-003
- FR-002 -> T-001, T-003
- FR-003 -> T-002, T-003
- NFR-001 -> T-003
- NFR-002 -> T-001, T-002, T-003
- NFR-003 -> T-001, T-002, T-003
- NFR-005 -> T-001, T-002, T-003
- NFR-006 -> T-001, T-002, T-003

## Rollout and rollback

- Feature flag: None
- Migration sequencing: define domain reasons first, adapt usecases next, then persistence/read-side, then validation
- Rollback steps: revert the typed-reason refactor if any persistence/read-side compatibility issue appears

## Validation evidence

- `go test ./internal/domain/... ./internal/application/usecases -run 'TestRunReceiptPollingCycleUseCase|TestRunReceiptWebhookDispatchCycleUseCase'` passed.
- `go test ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres ./internal/application/usecases` passed.
- `go test ./...` passed.
- `SPEC_DIR="specs/2026-03-28-process-error-reason-cleanup" bash scripts/spec-lint.sh` passed.
- `bash scripts/precommit-run.sh` passed.
