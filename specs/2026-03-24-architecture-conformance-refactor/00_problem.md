---
doc: 00_problem
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

# Problem & Goals

## Context

- Background: repo 已經採用 `internal/domain`、`internal/application`、`internal/adapters`、`internal/bootstrap`、`internal/infrastructure` 的基礎結構，但原本仍有高流量 entrypoint 與 helper 落在錯層，且 composition root 一度放在 `internal/infrastructure/di`，與 infrastructure 應代表外部技術資源的語義衝突。完成 relocation、runtime 拆分與 infrastructure 純化後，後續又暴露兩個收尾問題: bootstrap 仍直接組 adapter 專屬的 `policy.AddressPolicyConfig`，以及 bootstrap test locality 被額外拆成另一份次日 spec，沒有留在同一份 architecture refactor source of truth。
- Users or stakeholders: 維護 payrune 服務與 worker 的開發者、未來接手重構的人。
- Why now: 使用者準備進行更認真的架構重整；若不先移除明顯的邊界違規，後續 feature/refactor 會持續把不該在 `cmd/` 與不該跨層的依賴繼續擴大。

## Constraints (optional)

- Technical constraints: 不改變既有對外行為；維持單一 Go module、既有 `cmd/` + `internal/` 佈局；避免引入新的抽象層來掩蓋當前只有單一實作的責任。
- Timeline/cost constraints: 這一輪延伸為把 `internal/infrastructure` 剩餘污染點一併清乾淨，但仍不做全 repo 大搬家。
- Compliance/security constraints: webhook 驗證、TLS 啟動、create2 metadata 推導行為不得被弱化。

## Problem statement

- Current pain: 前一輪已把 `cmd/` 薄化、移走 composition root、拆開 runtime wiring，並清掉 infrastructure 的 adapter/domain/env 耦合，但最後收尾仍有兩個邊界不夠乾淨。第一，`internal/bootstrap/api.go` 與 `api_worker.go` 直接組 adapter 專屬的 `policy.AddressPolicyConfig`，讓 composition root 需要理解 adapter 私有輸入型別。第二，bootstrap test locality cleanup 被拆到 `2026-03-25-bootstrap-test-locality`，使同一波 architecture refactor 的 source of truth 又分裂成兩份 spec。
- Evidence or examples:
  - `cmd/fake-webhook-receiver` 內含 HTTP handler、簽章驗證、TLS 憑證產生與 env 解析。
  - `cmd/poller` 與 `cmd/webhook-dispatcher` 內含 runtime config/env validation。
  - `cmd/payrune-worker` 內含 operation routing。
  - `internal/infrastructure/ethereumcreate2assets/assets.go` 直接呼叫 adapter 層的 create2 helper。
  - `internal/infrastructure/di/*` 曾經組裝 use case、controller、runtime builder，不屬於外部服務或 driver。
  - `AGENTS.md` 一度把 wiring 位置描述成 `internal/bootstrap/di`。
  - `internal/bootstrap/di/*` 曾同時承載 API process/worker、poller process/worker、webhook dispatcher process/worker 的 wiring，閱讀與測試定位成本偏高。
  - `internal/infrastructure/cloudflarepostgres/*` 應直接表達 Cloudflare Postgres runtime bridge，而不是再包一層 `drivers` generic 名詞。
  - `internal/infrastructure/cloudflarewebhook/*` 應直接表達 Cloudflare webhook runtime bridge，而不是再包一層 `drivers` generic 名詞。
  - `internal/infrastructure/postgres/connection.go` 應直接表達 process-hosted postgres integration，而不是再透過 `drivers/postgres` 間接命名。
  - `internal/infrastructure/ethereumcreate2assets/assets.go` 仍依賴 `internal/domain/valueobjects.NetworkID`。
  - `internal/bootstrap/api.go` 與 `internal/bootstrap/api_worker.go` 仍直接組 `policyadapter.AddressPolicyConfig`，表示 bootstrap 需要知道 adapter 的私有 config shape。
  - `specs/2026-03-25-bootstrap-test-locality/` 把同一波 bootstrap cleanup 分裂到第二份 spec。

## Goals

- G1: 讓 `cmd/` entrypoint 回到薄層，只保留 process startup 與單一 bootstrap 呼叫。
- G2: 清除 infrastructure 反向依賴與 misplaced composition root，讓 helper 和 wiring 各自回到正確層級。
- G3: 在不改變既有對外行為的前提下，補上對重構後 package 邊界的測試覆蓋。
- G4: 讓 AGENTS 與實際 repo 邊界一致，避免後續 agent 再把 DI 放回 infrastructure。
- G5: 讓 bootstrap runtime wiring 依具體責任分散在清楚的檔案群，而不是混成單一 package 或多餘的 `di` 子樹。
- G6: 讓 `internal/infrastructure` 不再 import adapter/domain 套件，technical contract 與 env ownership 都回到正確邊界。
- G7: 讓 `internal/infrastructure` 直接以具體技術名稱命名，不保留多餘的 `drivers` 目錄與 `*driver` package 名。
- G8: 收掉 `internal/bootstrap/di`，讓 wiring ownership 直接回到各自的 `internal/bootstrap` runtime。
- G9: 移除 bootstrap runtime 中沒有辨識價值的 `process` 命名，保留真正有資訊量的 runtime/env distinction。
- G10: 將過度拆分的 bootstrap source files 重新按 owning runtime 合併，避免一個 bootstrap area 被切成過多小檔。
- G11: 讓 bootstrap 組裝 address policy catalog 時只使用 domain-native policy 資料，不直接依賴 adapter 專屬 config 型別。
- G12: 將 bootstrap test locality 與 worker test 補強收回同一份 2026-03-24 spec，維持單一 architecture refactor source of truth。

## Non-goals (out of scope)

- NG1: 不在本輪全面重組所有 DI container 內部實作或進一步去重所有 runtime builder。
- NG2: 不在本輪完整重寫 `cmd/ethereum-create2-tool` 成獨立內部套件。
- NG3: 不新增新的 domain 模型或改動持久化 schema。

## Assumptions

- A1: 使用者要的是「先把明顯不符合架構的 code 收乾淨」，可以先處理最高訊號的違規點，而不是一次做全 repo rewrite。
- A2: `codex` 可作為 spec owner 以便本輪把 spec 狀態推進到可驗證狀態。

## Open questions

- Q1: `cmd/ethereum-create2-tool` 是否要在下一輪視為 CLI inbound adapter 來完整搬離 `cmd/`？

## Success metrics

- Metric: `cmd/` 內被保留的非測試 `.go` 檔只包含 `main`/stub 與必要 build tag glue。
- Target: `cmd/fake-webhook-receiver`、`cmd/poller`、`cmd/webhook-dispatcher`、`cmd/payrune-worker` 不再持有可重用 handler / env parsing / operation dispatch 邏輯。
- Metric: infrastructure 反向依賴。
- Target: `internal/infrastructure/ethereumcreate2assets` 不再 import `internal/adapters/outbound/ethereum`。
- Metric: infrastructure 目錄純度。
- Target: `internal/infrastructure` 不再承擔 composition root；DI/wiring 改由 `internal/bootstrap` 承擔。
- Metric: repo contract alignment。
- Target: `AGENTS.md` 明確把 wiring 指向 `internal/bootstrap`，並將 `internal/infrastructure` 定義為外部技術資源層。
- Metric: composition package clarity。
- Target: `internal/bootstrap` 底下的 runtime wiring 以具體檔案群就近放置，而不是繼續保留 `di` 子樹或混合式 package。
- Metric: infrastructure purity。
- Target: `internal/infrastructure` source code 不再 import `internal/adapters` 或 `internal/domain`。
- Metric: env ownership。
- Target: `internal/infrastructure/postgres` 不再提供 `OpenFromEnv()`；`DATABASE_URL` 讀取責任改由 `internal/bootstrap` 承擔。
- Metric: infrastructure naming clarity。
- Target: `internal/infrastructure` 下的技術整合以具體技術名稱直接命名，不再保留泛化的 `drivers/` 目錄。
- Metric: bootstrap wiring locality。
- Target: `internal/bootstrap/di` 目錄不再存在；API、poller、webhook dispatcher 與 process postgres wiring 直接位於 `internal/bootstrap`。
- Metric: address policy composition boundary。
- Target: `internal/bootstrap` 不再直接組 adapter 專屬的 `AddressPolicyConfig`；address policy reader 改為接受 domain-native policy input。
- Metric: spec/source-of-truth locality。
- Target: `specs/2026-03-25-bootstrap-test-locality` 被合併回 `specs/2026-03-24-architecture-conformance-refactor`，bootstrap test locality 與 worker test worklog 只保留在同一份 architecture spec。
- Metric: 回歸驗證。
- Target: `SPEC_DIR="specs/2026-03-24-architecture-conformance-refactor" bash scripts/spec-lint.sh` 與相關 `go test` 通過。
