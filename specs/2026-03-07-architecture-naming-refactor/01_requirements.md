---
doc: 01_requirements
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

# Requirements

## Functional requirements

### FR-001 - Rename receipt notification persistence to outbox terminology

- Description:
  - Receipt notification persistence contracts and implementations must use outbox-oriented naming rather than repository naming.
- Acceptance criteria:
  - [x] The application outbound port is renamed away from `*Repository` to `*Outbox`.
  - [x] The transaction bundle field and all use cases use the new outbox name consistently.
  - [x] The postgres implementation name matches the outbox terminology.

### FR-002 - Move claimed outbox row type out of domain entities

- Description:
  - The claimed receipt notification row used by dispatch logic should not be modeled as a domain entity.
- Acceptance criteria:
  - [x] The claimed row type is moved to an application-level DTO/message type.
  - [x] The outbox port returns the new application-level type instead of a domain entity.

### FR-003 - Rename the multi-chain observer adapter

- Description:
  - The adapter that selects a chain-specific observer must use a name that reflects multi-chain observation rather than routing terminology.
- Acceptance criteria:
  - [x] `ChainRouterReceiptObserver` is renamed to `MultiChainReceiptObserver`.
  - [x] DI wiring, tests, and constructors use the new name consistently.

### FR-004 - Rename supported-chain value object for clarity

- Description:
  - The supported-chain value object used by public API and address-policy code must be clearly distinct from the generic `ChainID` type.
- Acceptance criteria:
  - [x] `Chain` is renamed to `SupportedChain`.
  - [x] The parse helper and constants use the new naming consistently.
  - [x] API, DTO, and policy code compile against the renamed type.

### FR-005 - Move receipt status-change validation into domain

- Description:
  - Validation and construction of receipt status-change data must happen in domain code before the outbox store persists it.
- Acceptance criteria:
  - [x] A domain event or domain object represents a payment receipt status change.
  - [x] Polling code creates the status-change object from domain state transitions.
  - [x] The outbox store no longer decides whether a status change is valid.

### FR-006 - Move receipt polling claim policy into domain

- Description:
  - The set of receipt statuses that are eligible for polling must be defined in domain code, not hardcoded in SQL adapters.
- Acceptance criteria:
  - [x] Pollable receipt statuses are provided by a domain policy/function.
  - [x] The tracking store receives the claimable statuses from the application layer.
  - [x] The tracking store no longer hardcodes receipt lifecycle states in SQL text.

### FR-007 - Create new receipt tracking state in domain before persistence

- Description:
  - A newly issued allocation must be turned into a receipt tracking domain entity before persistence writes it.
- Acceptance criteria:
  - [x] Domain code constructs the initial `PaymentReceiptTracking`.
  - [x] The tracking store persists the provided tracking state instead of inventing the initial status internally.
  - [x] Polling save paths persist a full tracking state rather than splitting business-specific save methods.

### FR-008 - Rename workflow persistence contracts to store terminology

- Description:
  - Payment address allocation and payment receipt tracking persistence contracts must use store terminology rather than repository terminology because they coordinate workflow state, claims, and leases instead of acting as classic aggregate repositories.
- Acceptance criteria:
  - [x] Application outbound ports are renamed from `*Repository` to `*Store`.
  - [x] Transaction bundle fields and use cases use the store terminology consistently.
  - [x] Postgres implementations and tests compile against the renamed store contracts.

### FR-009 - Move webhook delivery transition policy into domain

- Description:
  - Webhook delivery retry/fail/sent transition rules must be decided in domain code before the outbox store persists the result.
- Acceptance criteria:
  - [x] A domain policy produces delivery results for sent, retry, and failed outcomes.
  - [x] The webhook dispatch use case persists the domain-produced delivery result rather than constructing persistence-specific retry/fail inputs.
  - [x] The outbox store persists delivery results without deciding retry-vs-failed policy.

### FR-010 - Use case timestamps must reflect per-item processing time

- Description:
  - Use cases that claim batches and then process individual items must record outcome timestamps from the individual item processing moment, not from the batch claim moment.
- Acceptance criteria:
  - [x] Webhook dispatch uses per-notification time when marking sent or scheduling retry.
  - [x] Receipt allocation derives receipt expiry from the same issued-at base time used for the allocation.

### FR-011 - Use cases must fail fast on missing core dependencies

- Description:
  - Every application use case must validate its required collaborators before executing business flow.
- Acceptance criteria:
  - [x] Allocate, generate-address, list-address-policies, and check-health use cases return explicit configuration errors instead of panicking on nil dependencies.
  - [x] Existing polling and webhook dispatch use cases keep consistent dependency validation behavior.

### FR-012 - Receipt polling output counters must use precise semantics

- Description:
  - Receipt polling cycle output fields must distinguish terminal business failure from processing errors so logs and metrics are not misleading.
- Acceptance criteria:
  - [x] Polling output no longer overloads one `FailedCount` field for both terminal receipt failure and processing errors.
  - [x] Bootstrap logging uses the renamed output fields consistently.

### FR-013 - Allocate payment address use case must expose a single explicit constructor

- Description:
  - The allocate-payment-address use case must use one explicit constructor instead of a redundant default constructor plus `WithConfig` variant.
- Acceptance criteria:
  - [x] `NewAllocatePaymentAddressUseCaseWithConfig` is removed.
  - [x] `NewAllocatePaymentAddressUseCase` exposes a single explicit collaborator set with no duplicate constructor naming.
  - [x] DI and tests compile against the single constructor name consistently.

### FR-014 - Allocation issuance policy must live in domain

- Description:
  - Allocation issuance rules must be decided by domain code rather than being assembled inside the allocate use case.
- Acceptance criteria:
  - [x] Domain policy decides reservation attempt priority before fresh allocation.
  - [x] Domain code validates allocation readiness for address policy, chain, amount, and fingerprint requirements.
  - [x] Domain policy produces receipt issuance terms such as required confirmations and expiry.
  - [x] The allocate use case depends on the domain policy instead of holding network-specific issuance rules itself.

### FR-015 - Receipt lifecycle policy must live in domain

- Description:
  - Receipt expiry reason and paid-unconfirmed expiry-extension policy must be encapsulated in domain code rather than the polling use case.
- Acceptance criteria:
  - [x] Domain policy applies expiry decisions and the paid-unconfirmed extension rule.
  - [x] The polling use case no longer defines receipt-expiry reason constants or lifecycle-extension policy itself.

### FR-016 - Transaction-scoped collaborator bundles must use scope-oriented naming

- Description:
  - The unit-of-work callback bundle must use a name that reflects transaction scope rather than assuming every collaborator is a store.
- Acceptance criteria:
  - [x] `TxStores` is renamed to `TxScope`.
  - [x] `WithinTransaction` callbacks, builders, and tests compile against `TxScope`.
  - [x] The rename does not change unit-of-work behavior.

### FR-017 - Allocate use case must depend on the shared clock port

- Description:
  - The allocate-payment-address use case must use the shared outbound `Clock` port instead of a local function-shaped time hook so application-layer time handling stays consistent across use cases.
- Acceptance criteria:
  - [x] `NewAllocatePaymentAddressUseCase` accepts `outport.Clock`.
  - [x] Allocation code reads time through `Clock.NowUTC()`.
  - [x] DI and tests compile against the shared clock port instead of `func() time.Time`.

### FR-018 - Address generation must use a chain-generic derivation contract

- Description:
  - Address-policy-backed generation paths must depend on a chain-generic derivation contract rather than `BitcoinAddressDeriver` directly so the use case boundary stays aligned with its chain-scoped API contract instead of exposing bitcoin-only fields.
- Acceptance criteria:
  - [x] `GenerateAddressUseCase` depends on a generic chain address deriver port.
  - [x] `AddressPolicy` stores chain-generic derivation fields rather than bitcoin-specific field names.
  - [x] A bitcoin adapter implements the generic deriver port without moving bitcoin-specific derivation logic into the use case.
  - [x] DI and tests compile without injecting `BitcoinAddressDeriver` directly into `NewGenerateAddressUseCase`.

### FR-019 - Chain-specific address derivation must compose through a multi-chain adapter

- Description:
  - Runtime wiring for address derivation must use a multi-chain adapter/registry so future chain support can be added by registering a new chain-specific deriver rather than refactoring existing chain adapters or use cases.
- Acceptance criteria:
  - [x] A multi-chain address-deriver adapter implements the application `ChainAddressDeriver` port.
  - [x] DI injects the multi-chain adapter into generate/allocate use cases even when only bitcoin is configured today.
  - [x] The bitcoin-specific adapter exposes only its own chain binding and derivation logic; future chain registration should not require changing bitcoin-specific behavior.

### FR-020 - High-risk postgres persistence paths must have direct tests

- Description:
  - The postgres allocation store, receipt tracking store, and unit-of-work code paths must have direct tests that cover their SQL control flow, guard conditions, and transaction behavior rather than relying only on higher-level use case tests.
- Acceptance criteria:
  - [x] `PaymentAddressAllocationStore` has direct tests covering success and failure paths for `Complete`, `MarkDerivationFailed`, `ReopenFailedReservation`, and `ReserveFresh`.
  - [x] `PaymentReceiptTrackingStore` has direct tests covering `Create`, `ClaimDue`, `Save`, and scan/validation edge cases.
  - [x] `UnitOfWork` has direct tests covering missing builder guard, commit-on-success, and rollback-on-error.

### FR-021 - Helper tests and application message placement must match responsibility

- Description:
  - Small helper code should have direct tests when it is reused across handlers, and claimed outbox-row types should live in an application message-oriented package rather than `dto` when they are not external transport DTOs.
- Acceptance criteria:
  - [x] `json_response.go` has a direct test covering status code, content type, and JSON body behavior.
  - [x] `PaymentReceiptStatusNotificationOutboxMessage` is moved out of `internal/application/dto` to a package whose name reflects application messages rather than external DTOs.
  - [x] All imports and tests compile against the new package without behavior changes.

### FR-022 - Workflow outbox payloads must live in a dedicated application outbox package

- Description:
  - Application-layer workflow payloads that represent claimed outbox rows must live in a package whose ownership is explicitly outbox-specific, so `application/dto` stays reserved for use case boundary data and generic `application/messages` does not become a catch-all bucket.
- Acceptance criteria:
  - [x] `PaymentReceiptStatusNotificationOutboxMessage` lives under `internal/application/outbox`.
  - [x] Production code and tests no longer import `internal/application/messages` for the claimed receipt-notification outbox row.
  - [x] `internal/application/dto` remains focused on use case input/output DTOs rather than claimed outbox workflow payloads.

### FR-023 - Bitcoin-only derivation internals must stay inside the bitcoin adapter

- Description:
  - The bitcoin adapter's internal derivation collaborator should not be modeled as an application outbound port once application use cases no longer depend on it directly.
- Acceptance criteria:
  - [x] `BitcoinAddressDeriver` no longer lives under `internal/application/ports/out`.
  - [x] The bitcoin chain-address adapter depends on a bitcoin-package-local collaborator contract instead of an application port.
  - [x] DI and tests compile without reintroducing a bitcoin-only application port.

### FR-024 - Address policy metadata and issuance configuration must use separate domain types

- Description:
  - Public address-policy metadata used by listing flows and operational derivation/issuance configuration used by generate/allocate flows must no longer share one over-broad domain type.
- Acceptance criteria:
  - [x] `AddressPolicy` keeps only public/listing metadata plus explicit enabled state.
  - [x] A separate domain type represents derivation/issuance configuration used by generate and allocate flows.
  - [x] The address-policy reader returns public metadata for list flows and operational issuance policy data for generate/allocate flows.
  - [x] Generate and allocate use cases compile against the split types without behavior changes.

## Non-functional requirements

- Maintainability (NFR-001):
  - The refactor must reduce architecture ambiguity without introducing broader renames than necessary.
- Clarity (NFR-002):
  - New names must describe actual responsibility directly and should not require explanatory comments to decode.
- Safety (NFR-003):
  - The refactor must be behavior-preserving; tests and validations must pass unchanged.

## Dependencies and integrations

- Internal services:
  - Receipt polling, webhook dispatch, policy reading, and HTTP controller layers that currently reference the old names.
