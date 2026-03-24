---
doc: 03_tasks
spec_date: 2026-03-24
slug: architecture-conformance-refactor
mode: Full
status: DONE
owners:
  - codex
depends_on: []
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
- Rationale: 本輪跨越多個 entrypoint、infrastructure 邊界、composition root 位置與 repo contract 文件；需要明確記錄搬移責任、風險與驗證。
- Upstream dependencies (`depends_on`): []
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: 不適用，本 spec 採 Full mode。
  - What would trigger switching to Full mode: 不適用。
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): 不適用，本 spec 產出 test plan。

## Milestones

- M1: 完成 spec 與高訊號違規點盤點。
- M2: 完成 `cmd/` 薄化與 fake webhook handler 搬移。
- M3: 完成 create2 assets 反向依賴修正。
- M4: 完成 composition root 自 `internal/infrastructure/di` 搬到 `internal/bootstrap/di`。
- M5: 合併當日 spec 並更新 AGENTS 契約。
- M6: 完成 `internal/bootstrap/di` 依 runtime responsibility 拆分。
- M7: 清除 infrastructure 剩餘 adapter/domain/env 耦合。
- M8: 收回 bootstrap policy catalog 對 adapter-specific config 的依賴。
- M9: 合併 bootstrap test-locality spec 並跑 lint/test，更新 spec 狀態。

## Tasks (ordered)

1. T-001 - Scaffold and lock refactor scope
   - Scope: 建立並維護單一 Full-mode spec，記錄本輪高訊號邊界修正與後續補充工作。
   - Output: `specs/2026-03-24-architecture-conformance-refactor/` 全套 spec 文件。
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-24-architecture-conformance-refactor" bash scripts/spec-lint.sh`
     - [ ] Expected result: spec-lint 通過。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-002 - Thin out command entrypoints
   - Scope: 將 fake webhook handler、poller/dispatcher env parsing、worker operation dispatch 從 `cmd/` 搬到 `internal/` 合適邊界。
   - Output: 新的 bootstrap/helper/adapter package 與更新後的 `cmd/*/main.go`。
   - Linked requirements: FR-001, FR-003, NFR-002, NFR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/bootstrap ./internal/adapters/inbound/http/fakewebhook`
     - [ ] Expected result: 測試通過，且 `cmd/` 對應 package 無需再持有可重用邏輯。
     - [ ] Logs/metrics to check (if applicable): 檢查 fake webhook/poller/dispatcher 舊有測試語意仍存在。
1. T-003 - Remove infrastructure reverse dependency
   - Scope: 在 `internal/infrastructure/ethereumcreate2assets` 內部完成 create2 source-ref / init-code / hash helper，不再 import adapter 層。
   - Output: 更新後的 assets package 與對應測試。
   - Linked requirements: FR-002, FR-003, NFR-003, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/infrastructure/ethereumcreate2assets`
     - [ ] Expected result: 測試通過，且 package import 不再指向 `internal/adapters/outbound/ethereum`。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-004 - Move composition root to bootstrap
   - Scope: 將 `internal/infrastructure/di` 整包搬到 `internal/bootstrap/di`，修正 `internal/bootstrap/*` import 與對應測試。
   - Output: 新的 `internal/bootstrap/di` package 與更新後的 bootstrap 呼叫點。
   - Linked requirements: FR-004, FR-003, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `rg -n 'internal/infrastructure/di' internal cmd`、`go test ./internal/bootstrap/...`
     - [ ] Expected result: `internal`/`cmd` 不再引用舊 DI 路徑，且 bootstrap 測試通過。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-005 - Merge same-day spec and align AGENTS
   - Scope: 將今日兩份架構 refactor spec 收斂成單一 spec，並更新 `AGENTS.md` 的 wiring / infrastructure 契約。
   - Output: 單一 `2026-03-24-architecture-conformance-refactor` spec 與更新後的 `AGENTS.md`。
   - Linked requirements: FR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `test ! -d specs/2026-03-24-bootstrap-di-relocation`、`rg -n 'internal/bootstrap|internal/infrastructure' AGENTS.md`
     - [ ] Expected result: 第二份 spec 不存在，AGENTS 契約與當前邊界一致。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-006 - Decompose bootstrap composition by runtime responsibility
   - Scope: 讓 bootstrap runtime wiring 依 `api`、`poller`、`webhookdispatcher` 責任拆分，並修正 bootstrap 呼叫點與測試。
   - Output: 新的 runtime-specific wiring files 與更新後 import/test。
   - Linked requirements: FR-006, FR-003, FR-004, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `find internal/bootstrap -maxdepth 1 -name '*runtime*.go' | sort`、`go test ./internal/bootstrap/...`
     - [ ] Expected result: composition code 依 runtime 分散在具體檔案群，bootstrap 測試通過。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-008 - Remove infrastructure technical contract back-dependencies
   - Scope: 將 Cloudflare postgres/webhook 的 bridge contract 與低層 query error 型別移回具體的 `internal/infrastructure/*` package，修正 adapter 與 worker wiring 的依賴方向。
   - Output: 更新後的 technical contract、adapter import 與對應測試。
   - Linked requirements: FR-007, FR-003, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `rg -n 'internal/adapters' internal/infrastructure --glob '*.go'`、`go test ./internal/adapters/outbound/persistence/cloudflarepostgres ./internal/adapters/outbound/webhook ./internal/infrastructure/cloudflarepostgres ./internal/infrastructure/cloudflarewebhook`
     - [ ] Expected result: infrastructure source code 不再依賴 adapter，相關 adapter/infrastructure 測試通過。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-009 - Move process postgres env ownership to bootstrap
   - Scope: 拔除 `OpenFromEnv()`，在 `internal/bootstrap` 增加 postgres env helper，讓 API/poller/webhook dispatcher 改由 bootstrap 讀取 `DATABASE_URL`。
   - Output: 新的 bootstrap DB helper、更新後的 runtime wiring 與測試。
   - Linked requirements: FR-008, FR-003, FR-004, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `rg -n 'OpenFromEnv' internal`、`go test ./internal/bootstrap/... ./internal/infrastructure/postgres`
     - [ ] Expected result: 程式碼不再使用 `OpenFromEnv()`，runtime 測試與 postgres integration 測試通過。
     - [ ] Logs/metrics to check (if applicable): 無。
1. T-010 - Remove domain coupling from create2 assets

- Scope: 將 `ethereumcreate2assets` 改為 string-based network lookup key，修正 bootstrap 呼叫點與測試。
- Output: 更新後的 assets API、bootstrap call site 與對應測試。
- Linked requirements: FR-009, FR-002, FR-003, NFR-003, NFR-006
- Validation:
  - [ ] How to verify (manual steps or command): `rg -n 'internal/domain' internal/infrastructure --glob '*.go'`、`go test ./internal/infrastructure/ethereumcreate2assets ./internal/bootstrap/...`
  - [ ] Expected result: infrastructure assets 不再依賴 domain，source-ref 與 metadata 測試通過。
  - [ ] Logs/metrics to check (if applicable): 無。

1. T-011 - Remove redundant infrastructure `drivers` naming

- Scope: 移除 `internal/infrastructure/drivers` 目錄，將 Cloudflare 與 Postgres integration 直接收斂到 `internal/infrastructure/<name>`，並同步修正 package 名與 import。
- Output: 更新後的 infrastructure 路徑、package 名與所有 call site。
- Linked requirements: FR-010, FR-007, FR-008, NFR-006
- Validation:
  - [ ] How to verify (manual steps or command): `test ! -d internal/infrastructure/drivers`、`go test ./internal/bootstrap/... ./internal/adapters/outbound/persistence/cloudflarepostgres ./internal/adapters/outbound/webhook ./internal/infrastructure/...`
  - [ ] Expected result: `drivers` 目錄不存在，所有相關 package 測試通過。
  - [ ] Logs/metrics to check (if applicable): 無。

1. T-012 - Co-locate runtime wiring back into bootstrap

- Scope: 移除 `internal/bootstrap/di`，將 API / poller / receipt webhook dispatcher / postgres env wiring 直接移回 `internal/bootstrap`，並修正呼叫點與測試。
- Output: 更新後的 bootstrap runtime files、移除的 `di` 目錄與對應測試。
- Linked requirements: FR-011, FR-006, FR-008, NFR-006
- Validation:
  - [ ] How to verify (manual steps or command): `test ! -d internal/bootstrap/di`、`go test ./internal/bootstrap/...`
  - [ ] Expected result: `bootstrap/di` 目錄不存在，bootstrap 測試通過。
  - [ ] Logs/metrics to check (if applicable): 無。

1. T-013 - Verify bootstrap/infrastructure consolidation tranche

- Scope: 執行當時已完成 tranche 的 targeted repo 驗證，確認 infrastructure purity、bootstrap locality 與 naming cleanup 已成立。
- Output: tranche 驗證結果。
- Linked requirements: FR-003, FR-004, FR-005, FR-006, FR-007, FR-008, FR-009, FR-010, FR-011, FR-012, NFR-001, NFR-002, NFR-003, NFR-006
- Validation:
  - [ ] How to verify (manual steps or command): `rg -n 'internal/adapters|internal/domain' internal/infrastructure --glob '*.go'`、`test ! -d internal/infrastructure/drivers`、`test ! -d internal/bootstrap/di`、`go test ./...`、`SPEC_DIR="specs/2026-03-24-architecture-conformance-refactor" bash scripts/spec-lint.sh`
  - [ ] Expected result: 指定範圍驗證通過，spec 狀態一致。
  - [ ] Logs/metrics to check (if applicable): 無。

1. T-014 - Drop redundant bootstrap `process` naming

- Scope: 將 API / poller / receipt webhook dispatcher runtime/container 命名與 postgres env helper 去掉沒有辨識價值的 `process` 前綴，並同步修正檔名、呼叫點與測試。
- Output: 更新後的 bootstrap runtime file names、helper names 與測試。
- Linked requirements: FR-011, FR-012, FR-008, NFR-006
- Validation:
  - [ ] How to verify (manual steps or command): `rg -n 'ProcessContainer|ProcessPostgres|process_runtime' internal/bootstrap --glob '*.go'`、`go test ./internal/bootstrap/...`
  - [ ] Expected result: bootstrap runtime naming 不再保留多餘的 `process` 前綴，且測試通過。
  - [ ] Logs/metrics to check (if applicable): 無。

1. T-015 - Consolidate bootstrap source files

- Scope: 將 API / poller / receipt webhook dispatcher 的 run/config/runtime source files 按 owning bootstrap 合併，並移除獨立的 `internal/bootstrap/postgres_env.go`。
- Output: 更新後的 bootstrap source file layout 與測試。
- Linked requirements: FR-013, FR-011, FR-008, NFR-006
- Validation:
  - [ ] How to verify (manual steps or command): `test ! -f internal/bootstrap/postgres_env.go`、`go test ./internal/bootstrap/...`
  - [ ] Expected result: bootstrap source files 收斂為較少的大檔，且 `postgres_env.go` 不存在。
  - [ ] Logs/metrics to check (if applicable): 無。

1. T-016 - Replace adapter-specific address policy config with domain-native catalog input

- Scope: 將 outbound policy adapter 的輸入改成 domain-native `entities.AddressIssuancePolicy`，並讓 API process/worker runtime 在 bootstrap 內組裝 address policy catalog。
- Output: 更新後的 `internal/adapters/outbound/policy`、`internal/bootstrap/api.go`、`internal/bootstrap/api_worker.go` 與相關測試。
- Linked requirements: FR-014, FR-003, NFR-002, NFR-006
- Validation:
  - [ ] How to verify (manual steps or command): `rg -n 'AddressPolicyConfig' internal --glob '*.go'`、`go test ./internal/adapters/outbound/policy ./internal/bootstrap/...`
  - [ ] Expected result: 程式碼不再保留 `AddressPolicyConfig`，policy adapter 與 bootstrap 測試通過。
  - [ ] Logs/metrics to check (if applicable): 無。

1. T-017 - Merge bootstrap test-locality spec back into the 2026-03-24 architecture spec

- Scope: 將 `2026-03-25-bootstrap-test-locality` 的 source-of-truth 內容合併回本 spec，並刪除多餘的 spec folder。
- Output: 更新後的 `2026-03-24-architecture-conformance-refactor` spec 與刪除的 `2026-03-25-bootstrap-test-locality` 目錄。
- Linked requirements: FR-015, FR-005, FR-013, NFR-006
- Validation:
  - [ ] How to verify (manual steps or command): `test ! -d specs/2026-03-25-bootstrap-test-locality`、`find internal/bootstrap -maxdepth 1 -type f | sort`、`go test ./internal/bootstrap/...`
  - [ ] Expected result: 次日 spec 目錄不存在，bootstrap test file locality 與 worker tests 只在本 spec 記錄，且測試通過。
  - [ ] Logs/metrics to check (if applicable): 無。

1. T-018 - Final verification and spec closeout

- Scope: 重新執行本 spec 涵蓋的 targeted/full 驗證，並將 spec frontmatter 更新為 DONE。
- Output: 最終驗證結果與一致的 DONE 狀態 frontmatter。
- Linked requirements: FR-003, FR-004, FR-005, FR-006, FR-007, FR-008, FR-009, FR-010, FR-011, FR-012, FR-013, FR-014, FR-015, NFR-001, NFR-002, NFR-003, NFR-006
- Validation:
  - [ ] How to verify (manual steps or command): `go test ./...`、`SPEC_DIR="specs/2026-03-24-architecture-conformance-refactor" bash scripts/spec-lint.sh`
  - [ ] Expected result: 指定範圍驗證通過，spec 狀態一致。
  - [ ] Logs/metrics to check (if applicable): 無。

## Traceability (optional)

- FR-001 -> T-001, T-002
- FR-002 -> T-001, T-003
- FR-003 -> T-001, T-002, T-003, T-004, T-006, T-008, T-009, T-010, T-011
- FR-004 -> T-001, T-004, T-006, T-009, T-011
- FR-005 -> T-001, T-005, T-011
- FR-006 -> T-001, T-006, T-011
- FR-007 -> T-001, T-008, T-011, T-012
- FR-008 -> T-001, T-009, T-011, T-012
- FR-009 -> T-001, T-010, T-012
- FR-010 -> T-001, T-011, T-012
- FR-011 -> T-001, T-012, T-013, T-014
- FR-012 -> T-001, T-013, T-014
- FR-013 -> T-001, T-013, T-015
- FR-014 -> T-001, T-016, T-018
- FR-015 -> T-001, T-017, T-018
- NFR-001 -> T-013
- NFR-002 -> T-002, T-004, T-008, T-009, T-011, T-012, T-013, T-016, T-018
- NFR-003 -> T-003, T-010, T-013
- NFR-005 -> T-002
- NFR-006 -> T-001, T-002, T-003, T-004, T-005, T-006, T-008, T-009, T-010, T-011, T-012, T-013, T-014, T-015, T-016, T-017, T-018

## Validation evidence

- `SPEC_DIR="specs/2026-03-24-architecture-conformance-refactor" bash scripts/spec-lint.sh`
- `go test ./internal/adapters/inbound/http/fakewebhook`
- `go test ./internal/bootstrap`
- `go test ./internal/bootstrap/...`
- `find internal/bootstrap -maxdepth 1 -name '*runtime*.go' | sort`
- `go test ./internal/infrastructure/ethereumcreate2assets`
- `rg -n 'internal/infrastructure/di' internal cmd`
- `rg -n 'internal/adapters|internal/domain' internal/infrastructure --glob '*.go'`
- `rg -n 'OpenFromEnv' internal`
- `test ! -d internal/infrastructure/drivers`
- `test ! -d internal/bootstrap/di`
- `rg -n 'ProcessContainer|ProcessPostgres|process_runtime' internal/bootstrap --glob '*.go'`
- `test ! -f internal/bootstrap/postgres_env.go`
- `rg -n 'AddressPolicyConfig' internal --glob '*.go'`
- `go test ./internal/adapters/outbound/policy`
- `test ! -d specs/2026-03-25-bootstrap-test-locality`
- `find internal/bootstrap -maxdepth 1 -type f | sort`
- `go test ./cmd/fake-webhook-receiver ./cmd/poller ./cmd/webhook-dispatcher ./cmd/payrune-worker`
- `go test ./...`

## Rollout and rollback

- Feature flag: 無。
- Migration sequencing: 先搬移程式碼與測試，再跑 targeted 驗證，最後更新 spec 狀態。
- Rollback steps: 還原本輪移動的 package、AGENTS 契約與 import；因無 schema 變更，rollback 只需 revert code patch。
