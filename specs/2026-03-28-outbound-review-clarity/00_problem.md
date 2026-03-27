---
doc: 00_problem
spec_date: 2026-03-28
slug: outbound-review-clarity
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-outbound-port-error-conformance
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Problem & Goals

## Context

- Background: `outbound port error contract` 那輪已經把真正的 `outport` 邊界收斂好，但使用者指出 code review 仍然不夠直覺，因為 `internal/adapters/outbound` 裡還有一些 exported interface 其實只是 adapter-private collaborator，例如 `CloudflareEsploraBridge`、`Executor`、`Rows`、`Result`、`AddressEncoder`。這會讓 reviewer 難判斷「這個 interface 是不是 application port，error 要不要集中到 `outport`」。
- Users or stakeholders: 做 architecture review、adapter refactor、與 error contract cleanup 的維護者。
- Why now: 使用者明確要求把 review 規則做得更容易看懂，避免再發生「這個錯誤到底該不該收斂」的模糊地帶。

## Constraints (optional)

- Technical constraints: 不改變 runtime behavior；只改善 naming/locality 與 repo guidance。
- Timeline/cost constraints: Quick mode，小範圍整理。
- Compliance/security constraints: N/A.

## Problem statement

- Current pain: 即使 `outport` contract 已集中，reviewer 仍可能把 adapter-private bridge/helper 誤判成 application port，因為一些 interface 仍是 exported 名稱。
- Evidence or examples:
  - `internal/adapters/outbound/bitcoin/cloudflare_esplora_bridge.go`
  - `internal/adapters/outbound/bitcoin/hd_xpub_address_deriver.go`
  - `internal/adapters/outbound/persistence/postgres/executor.go`
  - `internal/adapters/outbound/persistence/cloudflarepostgres/executor.go`
  - `AGENTS.md` 雖然已有 error ownership 規則，但還沒有把「怎麼快速判斷 review 範圍」寫得夠具體。

## Goals

- G1: 讓 adapter-private collaborator 一眼可辨識，不再和 `outport` contract 混淆。
- G2: 把 review 規則寫進 `AGENTS.md`，明確區分 `outport` boundary、constructor、adapter-private collaborator。

## Non-goals (out of scope)

- NG1: 不再做一輪新的 outbound error contract 重構。
- NG2: 不移動 package 邊界或改寫 usecase。

## Assumptions

- A1: exported interface 不是錯，但在這個 repo 會增加 review 負擔，因此值得收成 unexported naming。
- A2: 若 exported constructor 需要接受 unexported interface type，現有 Go 用法仍可正常呼叫。

## Open questions

- Q1: 無。
- Q2:

## Success metrics

- Metric: reviewer 能否快速分辨哪些 method 必須只回 `outport.Err...`。
- Target: adapter-private interface 以 unexported naming 呈現，且 `AGENTS.md` 補上明確 review heuristics。
