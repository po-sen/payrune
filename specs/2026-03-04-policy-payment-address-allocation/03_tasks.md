---
doc: 03_tasks
spec_date: 2026-03-04
slug: policy-payment-address-allocation
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-03-btc-xpub-address-api
  - 2026-03-03-postgresql18-migration-runner-container
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Policy-Based Payment Address Allocation - Task Plan

## Mode decision

- Selected mode: Full
- Rationale:
  - Includes schema changes, non-trivial reserve/derive/finalize flow, and architecture boundary refactors.
- Upstream dependencies (`depends_on`):
  - `2026-03-03-btc-xpub-address-api`
  - `2026-03-03-postgresql18-migration-runner-container`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`.

## Milestones

- M1: Core allocation feature and schema delivered.
- M2: Repository/UoW boundaries stabilized and cleaned.
- M3: Policy provider and read path simplified.
- M4: Consolidated single-source spec finalized.

## Tasks (ordered)

1. T-001 - Deliver policy-based unique allocation flow

   - Scope:
     - Implement customer allocation API without index input and return reconciliation fields.
   - Output:
     - Controller/use case/API contract updates for allocation endpoint.
   - Linked requirements: FR-001, FR-004
   - Validation:
     - [x] `go test ./internal/adapters/inbound/http/controllers -count=1`
     - [x] `go test ./internal/application/use_cases -count=1`

2. T-002 - Implement xpub-partitioned cursor and lifecycle persistence

   - Scope:
     - Add migration-backed cursor/allocation schema and lifecycle transitions including failure reopen.
   - Output:
     - `000002` migration and postgres reservation/finalization logic.
   - Linked requirements: FR-002, FR-003, NFR-002
   - Validation:
     - [x] `go test ./... -short -count=1`

3. T-003 - Stabilize repository + UnitOfWork architecture

   - Scope:
     - Enforce UoW-owned transaction lifecycle with repository-agnostic UoW contract.
     - Keep repository outputs as entities/aggregates only.
     - Keep repository/transaction composition in DI via tx-repository builder and remove context-based implicit tx dependency from use case.
     - Normalize naming to `UnitOfWork`, `AddressPolicyReader`, `PaymentAddressAllocationRepository`.
     - Refactor allocation domain model to single aggregate state transitions (reserved/issued/derivation_failed).
   - Output:
     - Updated outbound ports/use cases/postgres adapter/DI wiring.
   - Linked requirements: FR-003, FR-006, NFR-006
   - Validation:
     - [x] `go test ./internal/application/use_cases -count=1`
     - [x] `go test ./... -short -count=1`

4. T-004 - Simplify policy provider path and remove config adapter coupling

   - Scope:
     - Move policy source to DI in-memory provider and unify list/find under one repository port.
     - Remove runtime file `internal/adapters/outbound/config/address_policy_repository.go`.
   - Output:
     - DI policy provider + updated list/generate/allocate dependencies.
   - Linked requirements: FR-005, FR-006, NFR-006
   - Validation:
     - [x] `rg -n "AddressPolicyQueryService|adapters/outbound/config" internal`
     - [x] `go test ./... -short -count=1`

5. T-005 - Consolidate split micro-specs into this canonical spec

   - Scope:
     - Merge architecture micro-adjustment context into this folder and remove superseded micro-spec folders.
   - Output:
     - Single feature-level spec source of truth.
   - Linked requirements: FR-006, NFR-006
   - Validation:
     - [x] `SPEC_DIR="specs/2026-03-04-policy-payment-address-allocation" bash scripts/spec-lint.sh`

6. T-006 - Split use case tests by responsibility
   - Scope:
     - Refactor application use case tests into separate files for `ListAddressPoliciesUseCase`, `GenerateAddressUseCase`, and `AllocatePaymentAddressUseCase`.
     - Move policy reader behavior tests to adapter-layer test file.
   - Output:
     - Remove monolithic `internal/application/use_cases/address_policy_use_cases_test.go`.
     - Add per-use-case test files and adapter policy reader test file.
   - Linked requirements: FR-006, NFR-006
   - Validation:
     - [x] `go test ./internal/application/use_cases -count=1`
     - [x] `go test ./internal/adapters/outbound/policy -count=1`

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-002
- FR-003 -> T-002, T-003
- FR-004 -> T-001
- FR-005 -> T-004
- FR-006 -> T-003, T-004, T-005, T-006
- NFR-002 -> T-002
- NFR-006 -> T-003, T-004, T-005, T-006

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - Apply migration before app deploy.
- Rollback steps:
  - Roll back app first, then revert migration only if data retention policy allows.

## Merged micro-spec history

- Merged into this spec and superseded:
  - `2026-03-04-uow-tx-repositories-bundle`
  - `2026-03-04-unit-of-work-naming-cleanup`
  - `2026-03-04-address-policy-repository-layering`
  - `2026-03-04-repository-naming-consistency`
  - `2026-03-04-address-policy-repository-name-fix`
  - `2026-03-04-xpub-fingerprint-di-strategy`
  - `2026-03-04-payment-allocation-repository-file-split`
  - `2026-03-04-adapter-directory-rollback`
  - `2026-03-05-remove-config-address-policy-adapter`
