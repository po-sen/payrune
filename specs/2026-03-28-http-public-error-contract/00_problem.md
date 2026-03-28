---
doc: 00_problem
spec_date: 2026-03-28
slug: http-public-error-contract
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-application-inbound-error-mapping
  - 2026-03-28-http-controller-api-locality
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Problem & Goals

## Context

- Background: chain-address HTTP controllers 已經拆成 per-API files，但目前 error response text 仍大量直接使用 `err.Error()`。
- Users or stakeholders: 維護 API contract 的後端開發者，以及需要穩定 error wording 的 API consumer。
- Why now: 使用者明確指出 controller 層現在仍難 review，尤其 public error text 其實還是和 application error string 綁在一起。

## Constraints (optional)

- Technical constraints: 不改 usecase/inport error contract，不改 route shape，不改 status code 行為。
- Timeline/cost constraints: Quick mode，只整理 HTTP inbound adapter 的 public message ownership 與對應測試。
- Compliance/security constraints: 不得讓 lower-layer raw error wording 成為 public HTTP contract。

## Problem statement

- Current pain: controller 目前雖然會用 `errors.Is(err, inport.Err...)` 分類，但最後回給 client 的 `error` 文字很多仍直接取自 `err.Error()`。
- Evidence or examples:
  - `generate_address_controller.go`、`allocate_payment_address_controller.go`、`get_payment_address_status_controller.go`、`list_address_policies_controller.go` 都有直接把 `err.Error()` 回進 `dto.ErrorResponse`。
  - 這讓 `internal/application/ports/inbound/errors.go` 的字串同時成為 application contract 和 HTTP public contract。

## Goals

- G1: 讓每個 HTTP controller 自己擁有 public error text，而不是直接轉發 `inport.Err...` 的字串。
- G2: 保持現有 status code 與整體錯誤分類不變。
- G3: 讓 controller tests 明確鎖住 public status/message contract。

## Non-goals (out of scope)

- NG1: 不改 usecase return 的 `inport.Err...` 型別與語義。
- NG2: 不做 generic/global error mapper framework。

## Assumptions

- A1: 少量 endpoint-local message duplication 比抽象化更符合這個 repo 的可讀性目標。
- A2: invalid chain path 的 public message 也應由 HTTP adapter 擁有，不再依賴 `inport.ErrChainNotSupported.Error()`。

## Open questions

- Q1: 無
- Q2:

## Success metrics

- Metric: HTTP controller production code 是否不再使用 `err.Error()` 生成 public error response
- Target: `internal/adapters/inbound/http/controllers/*.go` 中，error response 不再從 `inport.Err...` 文字直接輸出
- Metric: controller tests 是否同時鎖住 status 與 public error text
- Target: per-controller tests 覆蓋 public error mapping 並通過
