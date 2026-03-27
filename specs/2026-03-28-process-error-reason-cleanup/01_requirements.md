---
doc: 01_requirements
spec_date: 2026-03-28
slug: process-error-reason-cleanup
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

- Process error reason code:
- A domain-owned value object representing a persisted operational failure reason for receipt tracking or webhook delivery.

## Out-of-scope behaviors

- OOS1: No DB migration
- OOS2: No cleanup of allocation derivation failure reasons in this spec

## Functional requirements

### FR-001 - Receipt tracking must own a typed failure-reason model

- Description: `PaymentReceiptTracking` and related receipt-tracking flows must use a domain typed failure-reason value object instead of a free-form string.
- Acceptance criteria:
  - [ ] `PaymentReceiptTracking` stores a typed failure reason value instead of a raw string.
  - [ ] Polling usecase maps lower-level errors to domain reason codes, not to ad-hoc strings.
  - [ ] Expiration continues to use a domain-owned reason code.
- Notes: DB persistence may still serialize the code as a string, but application/domain logic must use the typed value.

### FR-002 - Webhook delivery results must own a typed failure-reason model

- Description: webhook delivery policy/results and claimed outbox models must use a domain typed failure reason instead of a free-form string.
- Acceptance criteria:
  - [ ] `ResolvePaymentReceiptStatusNotificationDeliveryFailure(...)` accepts a typed reason code.
  - [ ] Retry and terminal-failure paths persist the typed reason through stores/outbox models.
  - [ ] Webhook dispatch tests verify the typed reason contract.
- Notes: This spec does not change delivery status semantics, only the reason representation.

### FR-003 - Read-side models must expose public text, not internal code/raw text

- Description: status/query surfaces should convert typed failure reasons to public text instead of leaking raw persisted strings.
- Acceptance criteria:
  - [ ] `PaymentAddressStatusRecord` carries a typed tracking failure reason value.
  - [ ] `GetPaymentAddressStatusUseCase` maps the typed reason to public text in the DTO.
  - [ ] Legacy persisted raw strings are normalized to a safe typed/public representation when read.
- Notes: This keeps internal reason codes stable while preserving a readable API.

## Non-functional requirements

- Performance (NFR-001): No extra IO or new transaction boundary.
- Availability/Reliability (NFR-002): Existing polling, webhook delivery, and status lookup behavior remains functionally unchanged aside from reason representation.
- Security/Privacy (NFR-003): Lower-level runtime or adapter wording is no longer persisted or re-exposed through these receipt process reason paths.
- Compliance (NFR-004):
- Observability (NFR-005): Tests cover typed reason serialization/parsing and public-text mapping on polling, webhook dispatch, and status read paths.
- Maintainability (NFR-006): Receipt process reasons have a single domain owner and are not duplicated as usecase-local string catalogs.

## Dependencies and integrations

- External systems: None new
- Internal services: `internal/domain/valueobjects`, `internal/domain/entities`, `internal/domain/policies`, `internal/application/usecases`, persistence adapters, and related tests
