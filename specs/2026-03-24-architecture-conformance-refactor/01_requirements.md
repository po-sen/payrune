---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Thin entrypoint:
  - `cmd/<app>` 只保留程序啟動、signal handling、單次 bootstrap 呼叫與必要 build tag glue。
- High-signal violation:
  - 明顯把可重用邏輯放在錯誤層級，或讓依賴方向直接反轉的實作。
- Composition root:
  - 建立 concrete adapter、driver、policy、use case 與 runtime handler 的 wiring 入口。

## Out-of-scope behaviors

- OOS1: 不整理所有歷史命名與 container 重複組裝問題。
- OOS2: 不把所有 CLI/tooling 套件一次搬進新的 shared framework。

## Functional requirements

### FR-001 - Thin command entrypoints

- Description: `cmd/fake-webhook-receiver`、`cmd/poller`、`cmd/webhook-dispatcher`、`cmd/payrune-worker` 必須把可重用邏輯搬到 `internal/` 內合適層級，`cmd/` 本身只保留 entrypoint glue。
- Acceptance criteria:
  - [ ] `cmd/fake-webhook-receiver` 不再持有 webhook handler 驗證邏輯。
  - [ ] `cmd/poller` 不再持有 poller env parsing 與 validation 邏輯。
  - [ ] `cmd/webhook-dispatcher` 不再持有 dispatcher env parsing 與 validation 邏輯。
  - [ ] `cmd/payrune-worker` 不再持有 worker operation routing switch。
- Notes: 搬移後的邏輯位置需符合現有分層責任，例如 transport handler 放 inbound adapter，runtime/env startup 放 bootstrap。

### FR-002 - Remove infrastructure reverse dependency

- Description: `internal/infrastructure/ethereumcreate2assets` 必須獨立完成 metadata/source-ref 所需的本地 helper 計算，不能再直接依賴 outbound adapter 套件。
- Acceptance criteria:
  - [ ] `internal/infrastructure/ethereumcreate2assets` 不再 import `internal/adapters/outbound/ethereum`。
  - [ ] create2 metadata 推導與 init code hash 行為保持既有測試可驗證的一致性。
- Notes: 優先選擇具體、局部 helper，而不是新增抽象共享層。

### FR-003 - Preserve behavior through targeted tests

- Description: 重構後必須以測試證明 fake webhook 驗證、poller config parsing、dispatcher config parsing 與 assets helper 行為未回歸。
- Acceptance criteria:
  - [ ] fake webhook handler 測試在新 package 位置仍覆蓋成功、失敗、略過驗證、invalid JSON 場景。
  - [ ] poller 與 dispatcher config parsing 測試覆蓋成功與驗證失敗場景。
  - [ ] ethereum create2 assets 相關單元測試通過。
- Notes: 測試可搬移到更符合邊界的新 package，但不可直接刪除而不補。

### FR-004 - Keep composition root out of infrastructure

- Description: dependency wiring / runtime builder 必須由 `internal/bootstrap` 持有，而不是 `internal/infrastructure`。
- Acceptance criteria:
  - [ ] `internal/infrastructure/di` 目錄不再存在。
  - [ ] `internal/bootstrap` 提供原本的 container/runtime builder 能力。
  - [ ] `internal/bootstrap/*` 呼叫點使用同 package local wiring，不再 import 額外 DI package。
- Notes: 這一輪是 package boundary 修正，不要求同步重寫所有 container 內部重複組裝。

### FR-005 - Align repo contract docs with the new boundary

- Description: 今日重構必須只保留一份 2026-03-24 架構重構 spec，並讓 `AGENTS.md` 與當前 wiring/infrastructure 邊界一致。
- Acceptance criteria:
  - [ ] 今日兩份架構 refactor spec 合併為單一 spec folder。
  - [ ] `AGENTS.md` 將 wiring 指向 `internal/bootstrap`。
  - [ ] `AGENTS.md` 明確說明 `internal/infrastructure` 只承擔外部技術資源，不承擔 composition root。
- Notes: 較早的歷史 spec 可保留舊路徑作為當時狀態紀錄，不要求回寫所有歷史文件。

### FR-006 - Split bootstrap composition by runtime responsibility

- Description: `internal/bootstrap` 內的 runtime wiring 必須按具體 runtime responsibility 分散在清楚的檔案群，避免 API、poller、webhook dispatcher 與 worker runtime 全部混在單一 package 或語義重複的子樹。
- Acceptance criteria:
  - [ ] API wiring 位於 `internal/bootstrap` 內專屬的 runtime files。
  - [ ] poller wiring 位於 `internal/bootstrap` 內專屬的 runtime files。
  - [ ] webhook dispatcher wiring 位於 `internal/bootstrap` 內專屬的 runtime files。
  - [ ] `internal/bootstrap/*` 呼叫點改為依賴同 package local wiring。
- Notes: 可以保留必要的共用 helper，但應優先以具體 runtime file group 為中心，不新增過度泛化的 wiring framework。

### FR-007 - Remove infrastructure-to-adapter technical contract coupling

- Description: Cloudflare runtime bridge contract 與低層錯誤型別必須由具體的 infrastructure package 擁有，adapter 只能依賴它們，不能反過來讓 infrastructure import adapter。
- Acceptance criteria:
  - [ ] `internal/infrastructure/cloudflarepostgres` 不再 import adapter package。
  - [ ] `internal/infrastructure/cloudflarewebhook` 不再 import adapter package。
  - [ ] 對應 adapter 仍能使用相同 bridge 能力與錯誤判斷行為。
- Notes: 這裡的 contract 是技術邊界，不是 application port；定義在 infrastructure 技術邊界比留在 adapter 更符合依賴方向。

### FR-008 - Move process-hosted postgres env ownership to bootstrap

- Description: `DATABASE_URL` 的 env parsing 必須由 `internal/bootstrap` 承擔，postgres infrastructure package 只保留 open/ping 能力，不直接讀 process env。
- Acceptance criteria:
  - [ ] `internal/infrastructure/postgres` 不再提供 `OpenFromEnv()`。
  - [ ] process-hosted API、poller、webhook dispatcher runtime 仍能透過 bootstrap wiring 開啟 DB。
  - [ ] 缺少 `DATABASE_URL` 時仍回傳既有驗證錯誤。
- Notes: 可以抽出小型 bootstrap helper，但 ownership 必須留在 bootstrap。

### FR-009 - Remove infrastructure-to-domain coupling from create2 assets

- Description: `internal/infrastructure/ethereumcreate2assets` 必須改用技術層自己的 network key 表示，不直接依賴 domain value object。
- Acceptance criteria:
  - [ ] `internal/infrastructure/ethereumcreate2assets` 不再 import `internal/domain/valueobjects`。
  - [ ] 現有 embedded metadata lookup 與 source-ref 生成行為保持一致。
  - [ ] 呼叫端在 adapter/bootstrap 層完成 domain network 到 asset lookup key 的轉換。
- Notes: 這是技術資產索引，不是 domain invariant。

### FR-010 - Remove redundant `drivers` naming from infrastructure

- Description: `internal/infrastructure` 下的具體技術整合應直接以技術名稱命名，不再保留多餘的 `drivers/` 目錄與 `*driver` package 名。
- Acceptance criteria:
  - [ ] `internal/infrastructure/drivers` 目錄不再存在。
  - [ ] Cloudflare 與 Postgres integration import path 改為直接位於 `internal/infrastructure/<name>`。
  - [ ] 呼叫端與測試全部更新且可通過編譯。
- Notes: 這是命名收斂，不是新增抽象層。

### FR-011 - Co-locate wiring with bootstrap runtimes

- Description: `internal/bootstrap/di` 必須收掉，runtime wiring 直接放回對應的 `internal/bootstrap` 檔案群，避免 bootstrap 之下再多一層語義重複的 composition package。
- Acceptance criteria:
  - [ ] `internal/bootstrap/di` 目錄不再存在。
  - [ ] API runtime / worker wiring 直接位於 `internal/bootstrap`。
  - [ ] poller runtime / worker wiring 直接位於 `internal/bootstrap`。
  - [ ] receipt webhook dispatcher runtime / worker wiring 與 postgres env helper 直接位於 `internal/bootstrap`。
  - [ ] 呼叫端改為使用同 package local wiring，而不是 import `internal/bootstrap/di/*`。
- Notes: 允許少量重複以換取 locality，但不要退化成單一超大檔案。

### FR-012 - Remove redundant `process` naming from bootstrap runtime code

- Description: `internal/bootstrap` 內的 runtime/container/helper 命名應保留真正有辨識價值的差異；對於只是在 Go process 內運行、但沒有額外資訊量的 `process` 前綴，應予以移除。
- Acceptance criteria:
  - [ ] API runtime/container 命名不再包含 `process`。
  - [ ] poller runtime/container 命名不再包含 `process`。
  - [ ] receipt webhook dispatcher runtime/container 命名不再包含 `process`。
  - [ ] postgres env helper 以 env ownership 命名，而不是 `process postgres` 命名。
- Notes: `worker` 仍可保留，因為它代表 Cloudflare worker runtime 的實際區別。

### FR-013 - Consolidate bootstrap source files by owning runtime

- Description: `internal/bootstrap` 內的 source files 應優先按 owning bootstrap/runtime 合併，避免為了責任拆分而留下過多小檔；只有真正跨 area 且有明確獨立價值的檔案才應保留。
- Acceptance criteria:
  - [ ] API process runtime wiring 合併回 `api.go`。
  - [ ] API Cloudflare runtime wiring 合併回 `api_worker.go`。
  - [ ] poller config 與 runtime wiring 合併回 `poller.go`；Cloudflare runtime wiring 合併回 `poller_worker.go`。
  - [ ] receipt webhook dispatcher config 與 runtime wiring 合併回 `receipt_webhook_dispatcher.go`；Cloudflare runtime wiring 合併回 `receipt_webhook_dispatcher_worker.go`。
  - [ ] `internal/bootstrap/postgres_env.go` 不再存在。
- Notes: 減少檔案數優先於共用小 helper；允許少量重複以換取 locality。

### FR-014 - Keep address policy catalog assembly domain-native in bootstrap

- Description: `internal/bootstrap` 應負責決定當前部署有哪些 address issuance policy，但不應直接組 adapter 專屬 config 型別；outbound policy adapter 應改為接受 domain-native `entities.AddressIssuancePolicy`。
- Acceptance criteria:
  - [ ] `internal/adapters/outbound/policy` 不再公開 `AddressPolicyConfig` 作為 bootstrap 輸入型別。
  - [ ] `internal/bootstrap/api.go` 與 `internal/bootstrap/api_worker.go` 改為組 `[]entities.AddressIssuancePolicy` 或等價的 domain-native policy input。
  - [ ] API process runtime 與 API worker runtime 的 address policy catalog 行為保持一致。
  - [ ] 既有 policy reader 行為與測試維持通過。
- Notes: 這裡的 catalog 仍屬於 bootstrap/deployment decision，不應搬到 infrastructure；重點是把 bootstrap 對 adapter 私有 config 的依賴拿掉。

### FR-015 - Merge bootstrap test locality back into the architecture refactor spec

- Description: bootstrap source/test locality cleanup 與 worker test 補強屬於同一波 architecture conformance refactor，必須回收進本 spec，而不是保留獨立的次日 spec。
- Acceptance criteria:
  - [ ] `specs/2026-03-25-bootstrap-test-locality` 目錄不再存在。
  - [ ] 本 spec 明確記錄 bootstrap test file 命名收斂與 worker direct tests。
  - [ ] `internal/bootstrap/poller_worker_test.go` 與 `internal/bootstrap/receipt_webhook_dispatcher_worker_test.go` 持續存在並通過。
- Notes: 這是 source-of-truth 合併，不是回滾既有 cleanup。

## Non-functional requirements

- Performance (NFR-001): 啟動與 handler 執行路徑不得新增額外外部 IO；單元測試時間維持在既有本地可接受範圍，驗證以 targeted `go test` 完成。
- Availability/Reliability (NFR-002): process startup、worker request dispatch、fake webhook 回應狀態碼與既有一致，不得因重構變更 happy-path/failure-path 行為。
- Security/Privacy (NFR-003): webhook HMAC 驗證 header 與 fake receiver TLS 最低版本要求維持不變；create2 helper 對 hex/address 驗證不可放寬。
- Compliance (NFR-004): 不適用，無新增法規/資料處理變更。
- Observability (NFR-005): fake webhook log 內容與 poller/dispatcher runtime log 輸出語意維持既有可診斷性。
- Maintainability (NFR-006): 本輪完成後，至少四個 `cmd/` entrypoint 的非 wiring 邏輯被移到 `internal/`；`internal/infrastructure` 不再擁有 composition root，且 source code 不再依賴 adapter/domain；`AGENTS.md` 與實際目錄責任保持一致；bootstrap wiring 依具體 runtime 拆分並持有 process env ownership，且不保留語義重複的 `di` 子樹。
  Bootstrap runtime naming 也應去除沒有辨識價值的 `process` 前綴。
  Bootstrap source files 也應避免因過度拆分而降低閱讀 locality。
  Bootstrap 組裝不應再理解 adapter 專屬 config shape，且同一波 bootstrap cleanup 不應分裂成第二份 spec。

## Dependencies and integrations

- External systems: PostgreSQL、Cloudflare Worker JS bridge、Ethereum JSON-RPC、fake webhook HTTPS。
- Internal services: `internal/bootstrap`、`internal/adapters/inbound/http`、`internal/infrastructure/*`。
