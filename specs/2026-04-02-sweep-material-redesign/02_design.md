---
doc: 02_design
spec_date: 2026-04-02
slug: sweep-material-redesign
mode: Full
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-02-domain-model-boundary-cleanup
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Technical Design

## High-level approach

- Summary:
  - 將 `sweep_material_json` 明確收斂為 operator/persistence document。
  - domain entity 不再持有 serialized JSON。
  - `bitcoin` / `ethereum` issued deriver adapters 各自產出 `SweepMaterialJSON`。
  - `PaymentAddressAllocationStore.Complete` 只接收 allocation、`SweepMaterialJSON`、`IssuedAt` 後寫 DB。
  - generic `UnitOfWork` 與 bootstrap wiring 保持原本簡單形狀。
- Key decisions:
  - 不引入新的 `SweepMaterial` application model。
  - 不在 `internal/application/ports/outbound` 增加 constructor / validation helper。
  - 不新增新的 builder dispatcher / collaborator / 專用 UnitOfWork。
  - `SweepMaterialJSON` 雖然是 technical representation，但它作為 plain contract 經過 application 層是可接受的，因為 application 確實需要把它從 deriver 搬到 store。

## System context

- Components:
  - `AllocatePaymentAddressUseCase`
  - `IssuedPaymentAddressDeriver`
  - `PaymentAddressAllocation` entity
  - `PaymentAddressAllocationStore`
  - bitcoin issued deriver + local sweep material helper
  - ethereum issued deriver + local sweep material helper
  - postgres / cloudflarepostgres persistence adapters
- Interfaces:
  - `IssuedPaymentAddressDeriver.DeriveIssuedAddress(...)`
  - `PaymentAddressAllocationStore.Complete(...)`

## Key flows

- Flow 1:
  - use case reserve allocation
  - deriver returns `Address`, `IssuanceRefKind`, `IssuanceRef`, `SweepMaterialJSON`
  - entity `MarkIssued(...)` sets issued business state only
  - use case calls `allocationStore.Complete(input)`
  - persistence store writes `sweep_material_json` as received
  - adapter writes issued row + JSON in one update path
- Flow 2:
  - `FindIssuedByID` and other read paths no longer need to hydrate `SweepMaterialJSON` into domain entity
  - `sweep_material_json` remains persisted for operators, but does not re-enter the domain model

## Data model

- Entities:
  - `PaymentAddressAllocation`
    - removes `SweepMaterialJSON`
    - keeps only business state: policy ID, slot, amount, status, chain, network, scheme, address, failure reason
- Schema changes or migrations:
  - none
  - DB column `address_policy_allocations.sweep_material_json` remains unchanged
- Consistency and idempotency:
  - `Complete` remains the single write boundary for issued state + persisted operator payload
  - builder failure aborts `Complete`; no partial issued row should be committed

## API or contracts

- Endpoints or events:
  - no external API changes
- Request/response examples:
  - new `CompletePaymentAddressAllocationInput` shape:
    - `Allocation entities.PaymentAddressAllocation`
    - `SweepMaterialJSON string`
    - `IssuedAt time.Time`
  - updated `DeriveIssuedPaymentAddressOutput` shape:
    - `Address string`
    - `IssuanceRefKind valueobjects.IssuanceRefKind`
    - `IssuanceRef string`
    - `SweepMaterialJSON string`

## Backward compatibility (optional)

- API compatibility:
  - internal-only contract churn; no HTTP/API payload changes expected
- Data migration compatibility:
  - existing JSON format preserved; no migration

## Failure modes and resiliency

- Retries/timeouts:
  - no new retry policy
  - builder is pure in-process logic and should fail fast
- Backpressure/limits:
  - none beyond existing `Complete` path
- Degradation strategy:
  - if builder cannot construct a valid document, fail `Complete` and surface existing dependency/store failure path instead of writing incomplete JSON

## Observability

- Logs:
  - existing store-level failure logging remains sufficient
- Metrics:
  - none新增；依既有 `Complete` failure monitoring
- Traces:
  - no new trace boundaries
- Alerts:
  - none新增

## Security

- Authentication/authorization:
  - not applicable
- Secrets:
  - builder may reuse existing xpub / create2 metadata inputs only
- Abuse cases:
  - do not silently accept malformed issuance metadata and serialize garbage payload

## Alternatives considered

- Option A:
  - keep `SweepMaterialJSON` in entity
  - rejected because JSON representation leaks through the domain boundary
- Option B:
  - replace JSON with typed `SweepMaterial` in `internal/application/ports/outbound`
  - rejected because this repo does not benefit from a separate application mini-model here; it pushes port package toward overdesign
- Why chosen:
  - 這是最小且誠實的設計: chain-specific adapter 自己知道怎麼組 payload，application 只搬運，persistence 只寫入，generic wiring 不需要知道更多

## Risks

- Risk:
  - `SweepMaterialJSON` 仍會經過 application contract
- Mitigation:
  - domain entity 不持有它，且 application contract 只維持 plain string，不再額外抽象化
