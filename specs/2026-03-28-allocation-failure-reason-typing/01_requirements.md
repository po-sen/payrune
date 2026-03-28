---
doc: 01_requirements
spec_date: 2026-03-28
slug: allocation-failure-reason-typing
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-application-error-boundaries
  - 2026-03-27-domain-error-contracts
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Requirements

## Glossary (optional)

- Allocation derivation failure reason:
- A domain-owned value object representing why payment address derivation failed after reservation.

## Out-of-scope behaviors

- OOS1: No DB migration
- OOS2: No API surface rename or response shape change

## Functional requirements

### FR-001 - Allocation entity must own a typed derivation-failure reason

- Description: `PaymentAddressAllocation` must use a domain typed derivation-failure reason instead of a free-form string.
- Acceptance criteria:
  - [ ] `PaymentAddressAllocation` stores a typed derivation-failure reason value instead of raw text.
  - [ ] `MarkDerivationFailed(...)` accepts a typed domain reason code.
  - [ ] `MarkIssued(...)` clears the typed failure reason.
- Notes: persistence may still serialize the code as a string, but domain/application logic must use the typed value.

### FR-002 - Allocation issuance usecase must map lower-level derive errors to domain reason codes

- Description: `AllocatePaymentAddressUseCase` must not persist `deriveErr.Error()`; it should map derivation failure to a domain reason code before marking allocation failed.
- Acceptance criteria:
  - [ ] `persistDerivationFailure(...)` no longer passes raw `deriveErr.Error()` into the entity.
  - [ ] Unexpected derivation failures still preserve existing flow semantics: mark failed, save, release idempotency key.
  - [ ] Usecase tests assert typed reason behavior instead of raw error wording.
- Notes: this spec does not change retry/issuance semantics, only failure-reason representation.

### FR-003 - Allocation stores must serialize and parse typed derivation-failure reasons

- Description: postgres and cloudflarepostgres allocation stores must persist typed failure reasons while remaining compatible with the existing `failure_reason` text column.
- Acceptance criteria:
  - [ ] Allocation stores write the typed reason code string to `failure_reason`.
  - [ ] Allocation reads parse persisted reason text into the typed reason value.
  - [ ] Legacy persisted raw text is normalized to a safe typed reason when read.
- Notes: no migration is required.

## Non-functional requirements

- Performance (NFR-001): No extra IO, no new transaction boundary, and no schema change.
- Availability/Reliability (NFR-002): Allocation reservation, derivation failure handling, and issuance flows remain behaviorally unchanged aside from reason representation.
- Security/Privacy (NFR-003): Lower-level derivation error wording is no longer persisted into allocation process state.
- Compliance (NFR-004):
- Observability (NFR-005): Tests cover typed reason mapping, persistence serialization/parsing, and legacy compatibility.
- Maintainability (NFR-006): Allocation derivation failure reason has a single domain owner and is not duplicated as usecase-local strings.

## Dependencies and integrations

- External systems: None new
- Internal services: `internal/domain/valueobjects`, `internal/domain/entities`, `internal/application/usecases`, postgres/cloudflarepostgres allocation stores, and related tests
