---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - fake webhook handler 搬移後的 transport 行為。
  - poller config/env parsing。
  - receipt webhook dispatcher config/env parsing。
  - create2 assets helper 與 embedded metadata 行為。
  - bootstrap-owned container/runtime builder 測試。
  - bootstrap runtime wiring 依責任拆分後的編譯與測試。
  - bootstrap runtime naming 調整後的編譯與測試。
  - Cloudflare technical contract 反向依賴清除後的 adapter/infrastructure 測試。
  - postgres env ownership 移回 bootstrap 後的 DB helper 與 runtime 測試。
  - `internal/infrastructure` 不再依賴 adapter/domain 的邊界驗證。
  - `internal/infrastructure/drivers` 命名收斂後的 import path 與編譯驗證。
  - `internal/bootstrap/di` 收掉後的 bootstrap runtime 編譯與測試。
  - bootstrap runtime `process` naming cleanup 後的編譯與命名驗證。
  - bootstrap source file consolidation 與 `postgres_env.go` 移除後的編譯驗證。
  - bootstrap test file locality cleanup 與 worker direct tests。
  - address policy catalog 改成 domain-native bootstrap assembly 後的 adapter/bootstrap 測試。
- Not covered:
  - `cmd/ethereum-create2-tool` 全面重構。
  - 全 repo 所有 container 內部組裝去重。

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, FR-003, NFR-002, NFR-005
  - Steps: `go test ./internal/adapters/inbound/http/fakewebhook`
  - Expected: fake webhook handler 的 valid signature、invalid signature、missing secret、invalid JSON 測試通過。
- TC-002:
  - Linked requirements: FR-001, FR-003, NFR-002, NFR-006
  - Steps: `go test ./internal/bootstrap`
  - Expected: poller/dispatcher config parsing 與 worker dispatch 相關測試通過。
- TC-003:
  - Linked requirements: FR-002, FR-003, NFR-003, NFR-006
  - Steps: `go test ./internal/infrastructure/ethereumcreate2assets`
  - Expected: source-ref、init code hash、embedded metadata 測試通過。
- TC-004:
  - Linked requirements: FR-003, FR-004, NFR-002, NFR-006
  - Steps: `go test ./internal/bootstrap/...`
  - Expected: `internal/bootstrap` runtime wiring 測試通過。
- TC-005:
  - Linked requirements: FR-006, NFR-006
  - Steps: `find internal/bootstrap -maxdepth 1 -name '*runtime*.go' | sort`
  - Expected: composition code 位於 API、poller、webhook dispatcher 等具體 runtime files，而不是單一混合 package。
- TC-006:
  - Linked requirements: FR-007, FR-008, NFR-002, NFR-006
  - Steps: `go test ./internal/adapters/outbound/persistence/cloudflarepostgres ./internal/adapters/outbound/webhook ./internal/infrastructure/cloudflarepostgres ./internal/infrastructure/cloudflarewebhook ./internal/bootstrap/...`
  - Expected: Cloudflare bridge contract 變更與 postgres env wiring 相關測試通過。
- TC-007:
  - Linked requirements: FR-009, NFR-003, NFR-006
  - Steps: `go test ./internal/infrastructure/ethereumcreate2assets ./internal/bootstrap/...`
  - Expected: create2 assets 不再依賴 domain，metadata/source-ref 行為維持一致。
- TC-008:
  - Linked requirements: FR-014, NFR-002, NFR-006
  - Steps: `go test ./internal/adapters/outbound/policy ./internal/bootstrap/...`
  - Expected: outbound policy adapter 改為接受 domain-native policy input 後，reader 行為與 bootstrap catalog builder 測試通過。
- TC-009:
  - Linked requirements: FR-015, FR-013, NFR-002, NFR-006
  - Steps: `go test ./internal/bootstrap/...`
  - Expected: `poller_worker_test.go` 與 `receipt_webhook_dispatcher_worker_test.go` 存在並通過，bootstrap test file locality 維持一致。

### Integration

- TC-101:
  - Linked requirements: FR-001, FR-002, FR-003, FR-004, NFR-001, NFR-006
  - Steps: `go test ./...`
  - Expected: repo 全量編譯與測試通過。
- TC-102:
  - Linked requirements: FR-004, FR-005, NFR-006
  - Steps: `rg -n 'internal/infrastructure/di' internal cmd`
  - Expected: 無結果，表示程式碼已無舊 DI 路徑依賴。
- TC-103:
  - Linked requirements: FR-007, FR-009, NFR-006
  - Steps: `rg -n 'internal/adapters|internal/domain' internal/infrastructure --glob '*.go'`
  - Expected: 無結果，表示 infrastructure source code 不再依賴 adapter/domain。
- TC-104:
  - Linked requirements: FR-008, NFR-006
  - Steps: `rg -n 'OpenFromEnv' internal`
  - Expected: 無結果，表示 process env ownership 已不在 infrastructure driver。
- TC-105:
  - Linked requirements: FR-010, NFR-006
  - Steps: `test ! -d internal/infrastructure/drivers`
  - Expected: `drivers` 目錄不存在，表示命名已收斂到具體技術 package。
- TC-106:
  - Linked requirements: FR-011, NFR-006
  - Steps: `test ! -d internal/bootstrap/di`、`go test ./internal/bootstrap/...`
  - Expected: `bootstrap/di` 目錄不存在，且 bootstrap runtime wiring 測試通過。
- TC-107:
  - Linked requirements: FR-012, NFR-006
  - Steps: `rg -n 'ProcessContainer|ProcessPostgres|process_runtime' internal/bootstrap --glob '*.go'`
  - Expected: 無結果，表示 bootstrap runtime naming 不再保留多餘的 `process` 前綴。
- TC-108:
  - Linked requirements: FR-013, NFR-006
  - Steps: `test ! -f internal/bootstrap/postgres_env.go`、`go test ./internal/bootstrap/...`
  - Expected: 獨立的 `postgres_env.go` 不存在，且合併後的 bootstrap source files 仍可通過測試。
- TC-109:
  - Linked requirements: FR-014, NFR-006
  - Steps: `rg -n 'AddressPolicyConfig' internal --glob '*.go'`
  - Expected: 無結果，表示 bootstrap 不再依賴 adapter 專屬 address policy config。
- TC-110:
  - Linked requirements: FR-015, FR-005, NFR-006
  - Steps: `test ! -d specs/2026-03-25-bootstrap-test-locality`
  - Expected: 次日 bootstrap cleanup spec 已被合併刪除，只保留本 spec 作為 source of truth。

### E2E (if applicable)

- Scenario 1: 不適用，本輪不新增跨程序 e2e。
- Scenario 2: 不適用。

## Edge cases and failure modes

- Case: 無效 webhook JSON 或錯誤簽章。
- Expected behavior: 維持 `400` / `401` 行為與既有 log 訊息。
- Case: 缺少必要 dispatcher env。
- Expected behavior: config loader 回傳錯誤，不進入 runtime。
- Case: create2 metadata 或 hex input 不完整。
- Expected behavior: helper 拒絕產生 source ref 或回傳空字串/錯誤，與既有測試一致。
- Case: composition root relocation 後遺漏 import。
- Expected behavior: `go test ./...` 立即失敗，直到所有 import path 修正完成。
- Case: Cloudflare bridge contract 移動後，adapter 或 driver 仍保留舊 import 方向。
- Expected behavior: `rg` 邊界檢查會直接顯示殘留違規 import。

## NFR verification

- Performance: targeted `go test` 無顯著變慢；不新增外部依賴呼叫。
- Reliability: config loader、handler 與 bootstrap runtime relocation 測試涵蓋成功/失敗 path。
- Security: webhook HMAC 與 create2 hex validation 測試持續通過。
