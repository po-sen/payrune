---
doc: 01_requirements
spec_date: 2026-03-27
slug: domain-error-contracts
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-application-inbound-error-mapping
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Requirements

## Glossary (optional)

- Domain error contract:
  - A stable sentinel error exported from a domain package for invariant/rule violations with cross-layer meaning.

## Out-of-scope behaviors

- OOS1: 不為 application/outbound error 再重新分層。
- OOS2: 不為單一 caller 的 purely local helper error 建立多餘抽象。

## Functional requirements

### FR-001 - Extract stable domain sentinel errors

- Description: `internal/domain` 既有匿名且具跨層判斷價值的 invariant errors，必須改為 stable sentinel errors。
- Acceptance criteria:
  - [ ] AC1 `entities`, `valueobjects`, `events`, `policies` 中既有匿名 invariant error 會改成 exported sentinel error。
  - [ ] AC2 既有已具名的 domain errors 保持命名與語義穩定。
  - [ ] AC3 implementation 不引入單一全域 error dumping ground。
- Notes: package-local `errors.go` 或同檔 error block 都可以，只要 ownership 清楚。

### FR-002 - Domain callers can use errors.Is

- Description: application/usecase 與 domain tests 必須能透過 `errors.Is(...)` 判斷新的 domain errors。
- Acceptance criteria:
  - [ ] AC1 受影響的 domain tests 改為驗證 stable sentinel，而不是只看字串。
  - [ ] AC2 若 usecase 已對某些 domain rule 做 mapping，對應路徑仍然成立。
- Notes: 本 requirement 聚焦 contract 可判斷性，不要求所有 usecase 立即新增更多 mapping。

### FR-003 - Behavior remains unchanged

- Description: 抽取 domain errors 後，既有 domain validation 行為與 usecase business behavior 不得回歸。
- Acceptance criteria:
  - [ ] AC1 既有 domain validation 仍在相同情境失敗。
  - [ ] AC2 既有 application/usecase tests 仍通過。
- Notes: 這輪是 contract cleanup，不是 rule redesign。

## Non-functional requirements

- Performance (NFR-001): 不新增任何 runtime IO 或額外 allocation-heavy path；變更僅限 error definitions 與 assertions。
- Availability/Reliability (NFR-002): 既有 domain/application tests 必須維持綠燈，避免 rule regressions。
- Security/Privacy (NFR-003): 不適用。
- Compliance (NFR-004): 不適用。
- Observability (NFR-005): 不新增 logging；以測試作為主要證據。
- Maintainability (NFR-006): error 命名必須 concrete、可讀、package-owned；不可引入抽象 router / registry / hierarchy。

## Dependencies and integrations

- External systems: 無。
- Internal services: `internal/domain/...` 與依賴其 contract 的 `internal/application/usecases`。
