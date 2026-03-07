---
doc: 04_test_plan
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

# Test Plan

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, FR-002, NFR-001, NFR-003
  - Steps:
    - Run the outbox-related application and postgres adapter tests after the rename.
  - Expected:
    - Notification claiming, enqueueing, and delivery state tests still pass with the outbox naming.
- TC-002:
  - Linked requirements: FR-004, NFR-001, NFR-003
  - Steps:
    - Run value-object and HTTP/controller tests covering the supported-chain rename.
  - Expected:
    - The code compiles and tests pass with the renamed supported-chain type.

### Integration

- TC-101:

  - Linked requirements: FR-003, NFR-001, NFR-003
  - Steps:
    - Run poller and webhook dispatcher wiring tests after renaming the multi-chain observer adapter.
  - Expected:
    - DI and runtime tests still pass with the new adapter naming.

- TC-102:

  - Linked requirements: FR-005, FR-006, FR-007, FR-009, NFR-001, NFR-003
  - Steps:
    - Run allocation, polling, webhook dispatch, outbox, and persistence tests after moving lifecycle rules into domain objects/policies.
  - Expected:
    - Domain objects and domain policies decide lifecycle rules and delivery transitions, and stores persist the resulting state without re-encoding those business rules.

- TC-103:

  - Linked requirements: FR-008, NFR-001, NFR-002, NFR-003
  - Steps:
    - Run application and postgres persistence tests after renaming allocation and receipt tracking persistence from repository terminology to store terminology.
  - Expected:
    - Production code and tests compile with consistent store naming and no remaining repository terminology for those workflow persistence contracts.

- TC-104:

  - Linked requirements: FR-010, FR-011, FR-012, NFR-001, NFR-003
  - Steps:
    - Run use case and bootstrap tests covering per-item webhook timestamps, shared issued-at expiry calculation, missing dependency guards, and split polling counters.
  - Expected:
    - Use cases record accurate times, fail fast on nil dependencies, and expose precise polling output counters.

- TC-105:

  - Linked requirements: FR-013, NFR-001, NFR-002, NFR-003
  - Steps:
    - Build and run allocate-payment-address use case tests after removing the `WithConfig` constructor.
  - Expected:
    - The allocate use case compiles through a single explicit constructor surface and no active code path references `NewAllocatePaymentAddressUseCaseWithConfig`.

- TC-106:

  - Linked requirements: FR-014, FR-015, NFR-001, NFR-002, NFR-003
  - Steps:
    - Run domain policy tests plus allocation and polling use case tests after moving issuance and lifecycle policy into domain code.
  - Expected:
    - Domain policies own reservation priority, issuance defaults, expiry reason, and paid-unconfirmed extension, while use cases keep orchestration only.

- TC-107:

  - Linked requirements: NFR-001, NFR-002, NFR-003
  - Steps:
    - Run allocate-payment-address use case tests after refactoring the transaction body into orchestration helpers.
  - Expected:
    - Allocation behavior stays unchanged, and the use case remains covered after the readability refactor.

- TC-108:

  - Linked requirements: FR-016, NFR-001, NFR-002, NFR-003
  - Steps:
    - Run application, persistence, DI, and use case tests after renaming `TxStores` to `TxScope`.
  - Expected:
    - Transaction callbacks compile against `TxScope` and unit-of-work behavior is unchanged.

- TC-109:

  - Linked requirements: FR-017, NFR-001, NFR-002, NFR-003
  - Steps:
    - Run allocate-payment-address use case and DI tests after replacing the local clock function dependency with `outport.Clock`.
  - Expected:
    - Allocation logic reads time through `Clock.NowUTC()`, and constructor call sites compile without raw time-function injection.

- TC-110:

  - Linked requirements: FR-018, NFR-001, NFR-002, NFR-003
  - Steps:
    - Run generate-address use case, allocate-payment-address use case, bitcoin adapter, and DI tests after introducing the generic chain address deriver port and generic address-policy derivation fields.
  - Expected:
    - The use cases compile against the generic derivation contract, bitcoin derivation stays behind an adapter, and runtime wiring still produces bitcoin addresses correctly.

- TC-111:

  - Linked requirements: FR-019, NFR-001, NFR-002, NFR-003
  - Steps:
    - Run blockchain adapter, generate-address use case, allocate-payment-address use case, and DI tests after introducing the multi-chain address-deriver adapter.
  - Expected:
    - Runtime wiring goes through the multi-chain adapter, bitcoin remains a chain-specific implementation, and adding another chain later requires registration rather than refactoring bitcoin-specific logic.

- TC-112:

  - Linked requirements: FR-020, NFR-001, NFR-002, NFR-003
  - Steps:
    - Run postgres adapter tests covering allocation store state transitions, receipt tracking create/claim/save behavior, and unit-of-work commit/rollback semantics.
  - Expected:
    - Cursor updates, reopened reservations, lease claims, save guards, and transaction commit/rollback behavior are all covered by direct adapter tests.

- TC-113:

  - Linked requirements: FR-021, NFR-001, NFR-002, NFR-003
  - Steps:
    - Run controller/helper, use case, and postgres adapter tests after adding a direct `json_response.go` test and moving the claimed outbox-row type out of `application/dto`.
  - Expected:
    - `writeJSON` has direct coverage, and the claimed outbox-row type compiles from an application message-oriented package with no behavior change.

- TC-114:

  - Linked requirements: FR-022, NFR-001, NFR-002, NFR-003
  - Steps:
    - Run use case and postgres adapter tests after moving the claimed receipt-notification outbox row from `application/messages` to `application/outbox`.
  - Expected:
    - The outbox workflow payload compiles from a dedicated outbox package, and no production code depends on the generic `application/messages` package for that type.

- TC-115:

  - Linked requirements: FR-023, NFR-001, NFR-002, NFR-003
  - Steps:
    - Run bitcoin adapter and DI tests after moving the bitcoin-only derivation collaborator contract out of application ports.
  - Expected:
    - The bitcoin adapter keeps its internal collaborator private to the bitcoin package, and application code no longer depends on a bitcoin-only port.

- TC-116:
  - Linked requirements: FR-024, NFR-001, NFR-002, NFR-003
  - Steps:
    - Run domain, policy-reader, generate-address, allocate-payment-address, and postgres persistence tests after splitting public address-policy metadata from issuance configuration.
  - Expected:
    - Public listing flows compile against public policy metadata, generate/allocate compile against operational issuance policy data, and runtime behavior stays unchanged.

## Validation commands

- TC-901:

  - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, FR-008, FR-009, FR-010, FR-011, FR-012, FR-013, FR-014, FR-015, FR-016, FR-017, FR-018, FR-019, FR-020, FR-021, FR-022, FR-023, FR-024, NFR-001, NFR-002, NFR-003
  - Steps:
    - Run `SPEC_DIR="specs/2026-03-07-architecture-naming-refactor" bash scripts/spec-lint.sh`.
  - Expected:
    - Spec docs pass lint.

- TC-902:
  - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, FR-008, FR-009, FR-010, FR-011, FR-012, FR-013, FR-014, FR-015, FR-016, FR-017, FR-018, FR-019, FR-020, FR-021, FR-022, FR-023, FR-024, NFR-001, NFR-002, NFR-003
  - Steps:
    - Run `bash scripts/precommit-run.sh`.
  - Expected:
    - Repo validations pass after the refactor.
