---
doc: 03_tasks
spec_date: 2026-03-07
slug: architecture-naming-refactor
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on: []
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale:
  - This is a behavior-preserving naming refactor with no new schema, no new external integration, and no runtime design expansion.

## Tasks (ordered)

1. T-001 - Rename receipt notification outbox terminology

   - Scope:
     - Rename the receipt notification persistence port, transaction field, postgres adapter, and tests from repository naming to outbox naming.
     - Move the claimed outbox row type from domain entities to an application-level DTO/message type.
   - Output:
     - Consistent outbox-oriented naming across application, adapters, and tests.
   - Linked requirements: FR-001, FR-002, NFR-001, NFR-002, NFR-003
   - Validation:
     - [x] `go test ./... -short -count=1`
     - [x] The renamed API no longer uses `PaymentReceiptStatusNotificationRepository`.

2. T-002 - Rename multi-chain observer and supported-chain types

   - Scope:
     - Rename the multi-chain observer adapter away from router terminology.
     - Rename the supported-chain value object and propagate the new name through HTTP, DTO, policy, and use case code.
   - Output:
     - Multi-chain adapter and supported-chain value object use clearer names.
   - Linked requirements: FR-003, FR-004, NFR-001, NFR-002, NFR-003
   - Validation:
     - [x] `go test ./... -short -count=1`
     - [x] Old misleading names no longer remain in production code.

3. T-003 - Final validation and spec sync

   - Scope:
     - Run repo validation and update the spec to final state.
   - Output:
     - Done spec and clean validation evidence.
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, FR-008, FR-009, FR-010, FR-011, FR-012, FR-013, FR-014, FR-015, FR-016, FR-017, FR-018, FR-019, FR-020, FR-021, FR-022, FR-023, FR-024, NFR-001, NFR-002, NFR-003
   - Validation:
     - [x] `SPEC_DIR="specs/2026-03-07-architecture-naming-refactor" bash scripts/spec-lint.sh`
     - [x] `bash scripts/precommit-run.sh`

4. T-004 - Move persistence-embedded business rules into domain

   - Scope:
     - Add a domain receipt status-change object and move pollable receipt-status policy out of stores.
     - Create new tracking state in domain before persistence writes it.
     - Simplify tracking persistence so stores save domain state rather than branching on business paths.
     - Move webhook delivery retry/fail/sent transition rules into a domain policy.
   - Output:
     - Store methods only persist or claim already-decided domain state.
   - Linked requirements: FR-005, FR-006, FR-007, FR-009, NFR-001, NFR-002, NFR-003
   - Validation:
     - [x] `go test ./... -short -count=1`
     - [x] No store hardcodes receipt lifecycle rules, status-change validity, or webhook retry-vs-failed policy.

5. T-005 - Rename remaining workflow persistence to store terminology

   - Scope:
     - Rename payment address allocation and payment receipt tracking persistence ports, implementations, transaction bundle fields, and test doubles from repository naming to store naming.
   - Output:
     - Workflow persistence naming matches actual store semantics across production code and tests.
   - Linked requirements: FR-008, NFR-001, NFR-002, NFR-003
   - Validation:
     - [x] `go test ./... -short -count=1`
     - [x] Production code no longer uses `Repository` naming for allocation or receipt tracking persistence.

6. T-006 - Clean up use case orchestration semantics

   - Scope:
     - Use per-item time for webhook delivery results.
     - Derive receipt expiry from the same issued-at base time in allocation flow.
     - Add explicit missing-dependency guards to thin use cases.
     - Split receipt polling output counters into terminal-failure and processing-error semantics.
   - Output:
     - Use case runtime behavior and telemetry semantics match the actual processing model.
   - Linked requirements: FR-010, FR-011, FR-012, NFR-001, NFR-002, NFR-003
   - Validation:
     - [x] `go test ./... -short -count=1`
     - [x] Poller bootstrap logs terminal failures and processing errors separately.

7. T-007 - Remove redundant allocate use case constructor naming

   - Scope:
     - Delete `NewAllocatePaymentAddressUseCaseWithConfig`.
     - Keep a single `NewAllocatePaymentAddressUseCase` entry point with one explicit constructor surface.
     - Update DI and tests to call the single constructor consistently.
   - Output:
     - Allocate-payment-address constructor API is explicit and has no duplicate naming surface.
   - Linked requirements: FR-013, NFR-001, NFR-002, NFR-003
   - Validation:
     - [x] `go test ./... -short -count=1`
     - [x] No production code references `NewAllocatePaymentAddressUseCaseWithConfig`.

8. T-008 - Move remaining use-case business policy into domain

   - Scope:
     - Introduce a domain allocation-issuance policy for reservation priority, allocation validation, and receipt issuance terms.
     - Introduce a domain receipt-lifecycle policy for expiry reason and paid-unconfirmed expiry extension.
     - Refactor use cases, DI, and tests to depend on these domain policies instead of use-case-local policy fields and constants.
   - Output:
     - Application use cases orchestrate domain policy decisions instead of encoding the remaining business rules directly.
   - Linked requirements: FR-014, FR-015, NFR-001, NFR-002, NFR-003
   - Validation:
     - [x] `go test ./... -short -count=1`
     - [x] Allocate and polling use cases no longer hold issuance defaults, reservation priority, expiry reason, or paid-unconfirmed extension policy directly.

9. T-009 - Refactor allocate use case orchestration for readability

   - Scope:
     - Keep allocation reservation, issuance, and tracking creation in one transaction.
     - Extract the transaction body into focused orchestration helpers.
     - Add only minimal comments where the committed-failure flow would otherwise be hard to follow.
   - Output:
     - `AllocatePaymentAddressUseCase` reads as plan -> transact -> respond, with the transaction body split into understandable steps.
   - Linked requirements: NFR-001, NFR-002, NFR-003
   - Validation:
     - [x] `go test ./... -short -count=1`
     - [x] The allocate use case transaction no longer inlines reservation, derivation-failure persistence, and tracking creation in one long function body.

10. T-010 - Rename transaction bundle to `TxScope`

- Scope:
  - Rename the unit-of-work callback bundle from `TxStores` to `TxScope`.
  - Update postgres unit-of-work builder naming, use cases, and tests to match.
- Output:
  - Transaction-scoped collaborators use a scope-oriented name that stays valid even when the bundle mixes stores and outboxes.
- Linked requirements: FR-016, NFR-001, NFR-002, NFR-003
- Validation:
  - [x] `go test ./... -short -count=1`
  - [x] No active production code references `TxStores`.

1. T-011 - Align allocate use case with the shared clock port

- Scope:
  - Replace the allocate use case's local `func() time.Time` dependency with `outport.Clock`.
  - Update DI and tests to inject the shared clock port implementation/fakes.
  - Keep the existing behavior and validation semantics unchanged.
- Output:
  - Time-aware application use cases use one consistent clock abstraction.
- Linked requirements: FR-017, NFR-001, NFR-002, NFR-003
- Validation:
  - [x] `go test ./... -short -count=1`
  - [x] No active constructor call site passes a raw time function into `NewAllocatePaymentAddressUseCase`.

1. T-012 - Refactor address generation to a chain-generic derivation contract

- Scope:
  - Introduce a chain-generic address deriver outbound port shaped by the core use case.
  - Replace bitcoin-specific derivation field names in `AddressPolicy` with chain-generic ones at the core boundary.
  - Wrap bitcoin derivation behind a bitcoin adapter that implements the generic port.
  - Update generate-address use case, allocation issuance flow, DI, and tests to depend on the generic port instead of `BitcoinAddressDeriver`.
- Output:
  - Address generation and allocation keep a chain-generic application boundary while bitcoin derivation stays in adapters.
- Linked requirements: FR-018, NFR-001, NFR-002, NFR-003
- Validation:
  - [x] `go test ./... -short -count=1`
  - [x] No active `NewGenerateAddressUseCase` call site injects `BitcoinAddressDeriver` directly.

1. T-013 - Compose address derivation through a multi-chain adapter

- Scope:
  - Introduce a multi-chain address-deriver adapter that routes by chain and implements the application `ChainAddressDeriver` port.
  - Make bitcoin's address deriver self-describe its chain binding so the multi-chain adapter can register it without embedding bitcoin-specific branching in use cases or DI call sites.
  - Update DI and tests so generate/allocate always depend on the multi-chain adapter, even when only bitcoin is registered today.
- Output:
  - Future chain support becomes an additive adapter-registration change rather than a refactor of existing chain-specific wiring.
- Linked requirements: FR-019, NFR-001, NFR-002, NFR-003
- Validation:
  - [x] `go test ./... -short -count=1`
  - [x] No active use case constructor call site injects a single chain-specific deriver directly.

1. T-014 - Add direct tests for postgres persistence and transaction flows

- Scope:
  - Add focused tests for `PaymentAddressAllocationStore` cursor/reservation/state-transition behavior.
  - Add focused tests for `PaymentReceiptTrackingStore` create/claim/save SQL paths and scan edge cases.
  - Add direct tests for `UnitOfWork` commit/rollback behavior.
- Output:
  - High-risk postgres persistence code paths are covered by direct tests instead of only indirect use case coverage.
- Linked requirements: FR-020, NFR-001, NFR-002, NFR-003
- Validation:
  - [x] `go test ./internal/adapters/outbound/persistence/postgres -count=1`
  - [x] `go test ./... -short -count=1`

1. T-015 - Add direct helper coverage and move claimed outbox row out of DTO

- Scope:
  - Add a direct test for `json_response.go`.
  - Move `PaymentReceiptStatusNotificationOutboxMessage` from `internal/application/dto` into a package whose name reflects an application message rather than a transport DTO.
  - Update use cases, ports, adapters, and tests to compile against the renamed package.
- Output:
  - Helper coverage becomes explicit, and the claimed outbox-row type lives in a package whose semantics match its responsibility.
- Linked requirements: FR-021, NFR-001, NFR-002, NFR-003
- Validation:
  - [x] `go test ./internal/adapters/inbound/http/controllers ./internal/application/use_cases ./internal/adapters/outbound/persistence/postgres -count=1`
  - [x] `go test ./... -short -count=1`

1. T-016 - Move claimed outbox workflow payload into `application/outbox`

- Scope:
  - Move `PaymentReceiptStatusNotificationOutboxMessage` from the generic `internal/application/messages` package into `internal/application/outbox`.
  - Update ports, use cases, adapters, and tests to import the dedicated outbox package.
  - Keep `application/dto` reserved for use case input/output contracts.
- Output:
  - Claimed outbox workflow payloads live in a package whose ownership is explicitly outbox-specific.
- Linked requirements: FR-022, NFR-001, NFR-002, NFR-003
- Validation:
  - [x] `go test ./internal/application/use_cases ./internal/adapters/outbound/persistence/postgres -count=1`
  - [x] `go test ./... -short -count=1`

1. T-017 - Move bitcoin-only derivation collaborator out of application ports

- Scope:
  - Remove the bitcoin-only derivation collaborator interface from `internal/application/ports/out`.
  - Make the bitcoin chain-address adapter depend on a bitcoin-package-local interface instead.
  - Update DI and tests without changing runtime behavior.
- Output:
  - Application ports stay shaped by use-case needs, while bitcoin adapter internals remain private to the bitcoin adapter package.
- Linked requirements: FR-023, NFR-001, NFR-002, NFR-003
- Validation:
  - [x] `go test ./internal/adapters/outbound/bitcoin ./internal/infrastructure/di -count=1`
  - [x] `go test ./... -short -count=1`

1. T-018 - Split public address policy metadata from issuance configuration

- Scope:
  - Keep `AddressPolicy` as public metadata plus explicit enabled state.
  - Introduce separate domain types for derivation/issuance configuration and operational issuance policy lookup.
  - Update address-policy reader, generate/allocate use cases, stores, and tests to use the split contracts.
- Output:
  - Listing flows use public metadata; generate/allocate flows use operational issuance policy data.
- Linked requirements: FR-024, NFR-001, NFR-002, NFR-003
- Validation:
  - [x] `go test ./internal/domain/... ./internal/application/use_cases ./internal/adapters/outbound/policy ./internal/adapters/outbound/persistence/postgres -count=1`
  - [x] `go test ./... -short -count=1`

## Traceability

- FR-001 -> T-001, T-003
- FR-002 -> T-001, T-003
- FR-003 -> T-002, T-003
- FR-004 -> T-002, T-003
- FR-005 -> T-004, T-003
- FR-006 -> T-004, T-003
- FR-007 -> T-004, T-003
- FR-008 -> T-005, T-003
- FR-009 -> T-004, T-003
- FR-010 -> T-006, T-003
- FR-011 -> T-006, T-003
- FR-012 -> T-006, T-003
- FR-013 -> T-007, T-003
- FR-014 -> T-008, T-003
- FR-015 -> T-008, T-003
- FR-016 -> T-010, T-003
- FR-017 -> T-011, T-003
- FR-018 -> T-012, T-003
- FR-019 -> T-013, T-003
- FR-020 -> T-014, T-003
- FR-021 -> T-015, T-003
- FR-022 -> T-016, T-003
- FR-023 -> T-017, T-003
- FR-024 -> T-018, T-003
- NFR-001 -> T-001, T-002, T-003, T-004, T-005, T-006, T-007, T-008, T-009, T-010, T-011, T-012, T-013, T-014, T-015, T-016, T-017, T-018
- NFR-002 -> T-001, T-002, T-003, T-004, T-005, T-006, T-007, T-008, T-009, T-010, T-011, T-012, T-013, T-014, T-015, T-016, T-017, T-018
- NFR-003 -> T-001, T-002, T-003, T-004, T-005, T-006, T-007, T-008, T-009, T-010, T-011, T-012, T-013, T-014, T-015, T-016, T-017, T-018
