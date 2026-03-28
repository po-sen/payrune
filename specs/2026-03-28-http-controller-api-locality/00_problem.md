---
doc: 00_problem
spec_date: 2026-03-28
slug: http-controller-api-locality
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-24-architecture-conformance-refactor
  - 2026-03-27-application-inbound-error-mapping
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Problem & Goals

## Context

- Background: chain-address HTTP handling 原本集中在單一 shared controller shape，讓多個 API 的 request parsing、path parsing、error/status mapping 混在一起。
- Users or stakeholders: 維護 HTTP inbound adapter 與 review API contract 的開發者。
- Why now: 使用者明確指出目前很難一眼看出每個 API 會回哪些 status code，希望改善 review readability。

## Constraints (optional)

- Technical constraints: 不更動 route shape、不改 HTTP status/message 行為、不引入新的 generic error mapper。
- Timeline/cost constraints: Quick mode，只做檔案與責任 locality 重構。
- Compliance/security constraints: 對外 API contract 必須保持相容。

## Problem statement

- Current pain: 單一 controller 檔同時承載 4 個 API，status mapping 藏在各段流程裡，reviewer 很難快速看出單一 API 的契約。
- Evidence or examples:
  - bootstrap 端無法一眼看出四個 API controller 是怎麼組裝進 router。
  - test 端曾透過 shared helper 重建 aggregate-like wiring，和 production 結構不一致。

## Goals

- G1: 讓每個 API 的 handler、request parsing、error/status mapping 能集中在自己的檔案。
- G2: 保留現有 HTTP status、response body、header 行為不變。
- G3: 讓 reviewer 打開單一檔案就能看懂單一 API 契約。
- G4: endpoint-local controller 檔名要一眼可辨識，統一用 `*_controller.go`。
- G5: shared controller test scaffolding 也要用 owner-aligned 命名，不保留突兀的 helper-style 檔名。
- G6: 讓 `internal/bootstrap/api.go` 和 HTTP router 能直接看出每個 API controller 的組裝關係，不再透過 aggregate controller 隱藏。
- G7: controller / router / test 都遵守同一套規律，避免混用半舊半新的 pattern。

## Non-goals (out of scope)

- NG1: 不重設 HTTP error wording contract。
- NG2: 不改動 usecase 或 route registration。

## Assumptions

- A1: 少量 helper duplication 比隱性的 generic mapping 更符合這個 repo 的可讀性目標。
- A2: 以 transport-native `http.Handler` 統一 controller / router / test pattern，比保留 per-controller special method name 更容易 review。

## Open questions

- Q1: 無
- Q2:

## Success metrics

- Metric: 每個 API 是否有自己的 endpoint-local controller file
- Target: list/generate/allocate/get-status 各自有獨立檔案，且 tests/full suite 維持通過
- Metric: controller / router / test 是否遵守同一套組裝規律
- Target: 每個 controller 只有 `ServeHTTP` 入口，router 掛 `http.Handler`，tests 直接掛 route 到 controller
