---
doc: 01_requirements
spec_date: 2026-03-27
slug: issued-address-deriver-decoupling
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-application-error-boundaries
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null # set to 02_design.md in Full mode
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Requirements

## Glossary (optional)

- Concrete coupling:
  - One adapter depending on another adapter's concrete exported type instead of a consumer-shaped collaborator contract.
- Consumer-shaped collaborator:
  - The smallest interface or type contract needed by the consuming package, defined close to that consumer.
- Neutral dispatch:
  - A package that only routes by chain/provider and does not own chain-specific derivation rules.

## Out-of-scope behaviors

- OOS1: 不抽象化 `Create2SaltDeriver` 以外的 ethereum adapter。
- OOS2: 不把 `Create2SaltDeriver` 升成 application outbound port。

## Functional requirements

### FR-001 - Keep blockchain package neutral

- Description: `internal/adapters/outbound/blockchain` 必須只負責 issued address derivation 的 multi-chain dispatch，不再知道 `ethereum/create2` 等 chain-specific issued derivation 細節。
- Acceptance criteria:
  - [ ] `blockchain` package 不再分支 `ethereum/create2` issued derivation 流程。
  - [ ] `blockchain` package 改為只做 multi-chain deriver dispatch。
  - [ ] `blockchain` package 不再需要 `deriveCreate2SaltForIDFn` 這種 chain-specific hook。
- Notes: 這一層的責任應該是 dispatch，不是 chain-specific derivation。

### FR-002 - Move issued derivation logic back to chain-specific adapters

- Description: bitcoin 與 ethereum issued address derivation 必須各自實作於其 package；ethereum create2 專屬流程只留在 ethereum adapter。
- Acceptance criteria:
  - [ ] bitcoin issued derivation 由 `internal/adapters/outbound/bitcoin` owner。
  - [ ] ethereum issued derivation 由 `internal/adapters/outbound/ethereum` owner。
  - [ ] ethereum create2 salt / relative reference derivation 不再留在 `blockchain` package。
- Notes: 這輪重點是 ownership，不是功能變更。

### FR-003 - Preserve behavior and readable bootstrap wiring

- Description: 重構後行為不可變，且 bootstrap wiring 仍要直接、可讀，不引入 registry / factory indirection。
- Acceptance criteria:
  - [ ] 現有 bitcoin issuance 路徑不受影響。
  - [ ] 現有 ethereum create2 issuance 測試維持通過。
  - [ ] `bootstrap/api.go` 與 `bootstrap/api_worker.go` 仍能直接組裝 multi-chain issued deriver。
- Notes: 這輪重點是 readable composition，不是抽象化。

## Non-functional requirements

- Performance (NFR-001): 不新增額外 IO 或 runtime branching。
- Availability/Reliability (NFR-002): `go test ./internal/adapters/outbound/blockchain ./internal/bootstrap/...` 與 `go test ./...` 必須通過。
- Security/Privacy (NFR-003): 不改變 create2 secret / derivation material handling。
- Compliance (NFR-004):
- Observability (NFR-005): 不改變現有 logging / error propagation 行為。
- Maintainability (NFR-006): adapter ownership 更清楚，閱讀 `blockchain` 不需要理解 `ethereum/create2` 細節。

## Dependencies and integrations

- External systems: 無新增。
- Internal services:
  - `internal/adapters/outbound/blockchain`
  - `internal/adapters/outbound/ethereum`
  - `internal/bootstrap`
