---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Shared parsing helper:
  - 同一 package 內供 process / worker 共用的 lookup-based validator/builder，不承擔 runtime-specific wiring。
- Runtime-specific builder:
  - 仍需保留 process 或 worker 專屬依賴的 container / HTTP handler / scheduler runtime 建構邏輯。

## Out-of-scope behaviors

- OOS1: 不新增獨立的 bootstrap utility package 或 generic config framework。
- OOS2: 不改變 API / poller / webhook dispatcher 的外部 contract 或排程行為。

## Functional requirements

### FR-001 - Share API receipt terms parsing across process and worker

- Description: API process runtime 與 API worker runtime 對 payment receipt terms 的 confirmations / expires-after parsing 必須共用同一套 lookup-based helper，避免兩側維護兩份等價 map-building 邏輯。
- Acceptance criteria:
  - [ ] `internal/bootstrap/api.go` 與 `internal/bootstrap/api_worker.go` 不再各自維護獨立的 receipt terms map builder。
  - [ ] process 與 worker 仍可保留不同 default 值，但 map shape 與 scope mapping 由同一組 helper 產生。
  - [ ] 既有 API bootstrap 測試持續通過。
- Notes: 共用點是 parsing/builder，不是 runtime handler。

### FR-002 - Share poller dispatch parsing across process and worker

- Description: poller process config 與 worker request builder 的共通 dispatch 欄位 parsing 必須由同一組 helper 驅動，避免 `POLL_*` validation 重複維護。
- Acceptance criteria:
  - [ ] `POLL_BATCH_SIZE`、`POLL_RESCHEDULE_INTERVAL`、`POLL_CLAIM_TTL`、`POLL_CHAIN`、`POLL_NETWORK` 的 parsing / validation 不再在 process 與 worker 各寫一份。
  - [ ] `POLL_CHAIN is required when POLL_NETWORK is set` 的 validation 仍維持原本語意。
  - [ ] poller process 與 worker 測試持續通過。
- Notes: `POLL_TICK_INTERVAL` 只屬於 process runtime，可留在 process side。

### FR-003 - Share receipt webhook dispatcher parsing across process and worker

- Description: receipt webhook dispatcher 的 dispatch config parsing 與 notifier 共通欄位 parsing 必須盡量以 lookup-based helper 共用，降低 process / worker drift。
- Acceptance criteria:
  - [ ] dispatch batch / claim ttl / max attempts / retry delay 的 parsing 不再分散維護兩套相似邏輯。
  - [ ] notifier 的 `secret`、`timeout`、`insecureSkipVerify` parsing 由共通 helper 支援；runtime target 差異仍由 process / worker 各自決定。
  - [ ] receipt webhook dispatcher process 與 worker 測試持續通過。
- Notes: process 的 `URL` 與 worker 的 `CloudflareBinding/Path` 仍應各自保留。

### FR-004 - Keep dedup local to bootstrap ownership

- Description: 這輪 dedup 必須維持在 `internal/bootstrap` 本地 ownership，不得為了共享而抽出新的 generic package 或模糊命名層。
- Acceptance criteria:
  - [ ] 不新增 `internal/bootstrap/shared`、`internal/bootstrap/common`、`internal/bootstrap/framework` 之類目錄。
  - [ ] process / worker 的 runtime builder 仍留在各自 owning file。
  - [ ] 共享 helper 命名直接描述目前用途，而不是假設未來會支援更多 runtime。
  - [ ] `api.go` 不得反向依賴只定義在 `api_worker.go` 的 shared env key constants。
- Notes: 允許在現有檔案內新增小型 helper，或在 `internal/bootstrap` 新增極少量 concrete helper file。

## Non-functional requirements

- Performance (NFR-001): dedup 後不得新增 runtime 外部 IO；只允許本地 helper 抽取。
- Availability/Reliability (NFR-002): `go test ./internal/bootstrap/...` 與 `go test ./...` 必須持續通過。
- Security/Privacy (NFR-003): 不得放寬現有 env validation、duration/int parsing、bool parsing、webhook secret handling。
- Compliance (NFR-004): 不適用。
- Observability (NFR-005): 不改動現有 log/response behavior。
- Maintainability (NFR-006): process / worker 共通 parsing 應只保留單一 source of truth；新 helper 不得讓閱讀路徑比現在更抽象。

## Dependencies and integrations

- External systems: 無新增。
- Internal services: `internal/bootstrap`、`scripts/spec-lint.sh`。
