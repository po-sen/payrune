---
doc: 01_requirements
spec_date: 2026-03-29
slug: allocation-issuance-naming
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-20-create2-eth-payment-receiving
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Allocation Issuance Naming - Requirements

## Glossary (optional)

- Address space reference:
  - A stable internal identifier for the deterministic source configuration that owns a family of
    issued addresses. Bitcoin HD uses xpub-like material; Ethereum CREATE2 uses factory /
    collector / init-code metadata.
- Slot index:
  - The internal allocation cursor reserved within one address space. It is not required to be a
    derivation-path segment.
- Issuance reference:
  - The canonical per-address payload needed to reconstruct or reconcile one issued address.
- Issuance reference kind:
  - The typed label that explains how to interpret `issuance_ref`, such as an absolute HD path or
    a CREATE2 salt.

## Out-of-scope behaviors

- OOS1:
  - Public API response changes.
- OOS2:
  - New issuance methods beyond the existing Bitcoin HD and Ethereum CREATE2 flows.

## Functional requirements

### FR-001 - Rename allocation persistence to explicit address-space and slot terminology

- Description:
  - Allocation persistence and internal models must use names that describe the actual role of each
    field across both Bitcoin HD and Ethereum CREATE2 issuance.
- Acceptance criteria:
  - [ ] `address_policy_allocations.address_source_ref` is renamed to `address_space_ref`.
  - [ ] `address_policy_allocations.derivation_index` is renamed to `slot_index`.
  - [ ] `address_policy_cursors.address_source_ref` is renamed to `address_space_ref`.
  - [ ] Unique and lookup indexes are updated to key by `(address_policy_id, address_space_ref,
slot_index)` or the equivalent cursor lookup columns.
  - [ ] Internal Go models, ports, and persistence adapters use the same address-space and slot
        terminology rather than retaining the old names in code.
- Notes:
  - The refactor should preserve existing uniqueness and reservation behavior.

### FR-002 - Persist typed issuance references instead of mixed-shape reference strings

- Description:
  - Allocation rows must store the reconstructible issuance reference payload separately from the
    kind that explains how that payload should be interpreted.
- Acceptance criteria:
  - [ ] `address_policy_allocations.address_reference` is renamed to `issuance_ref`.
  - [ ] A new nullable column `issuance_ref_kind` is added to `address_policy_allocations`.
  - [ ] Bitcoin-issued rows persist `issuance_ref_kind=hd_path_absolute` and keep their absolute HD
        path payload in `issuance_ref`.
  - [ ] Ethereum CREATE2-issued rows persist `issuance_ref_kind=create2_salt` and store only the
        salt payload in `issuance_ref`.
  - [ ] No newly issued Ethereum row stores policy-ID or prefix material concatenated into
        `issuance_ref`.
- Notes:
  - Reserved and derivation-failed rows may keep `issuance_ref_kind` and `issuance_ref` null until
    an address is actually issued.

### FR-003 - Migrate existing issued rows without changing externally visible behavior

- Description:
  - The schema rename must preserve current allocation identity, receipt tracking, and status lookup
    behavior for existing data and future issued rows.
- Acceptance criteria:
  - [ ] The migration backfills `issuance_ref_kind` for existing issued Bitcoin and Ethereum rows.
  - [ ] Existing Ethereum CREATE2 `address_reference` values are rewritten so `issuance_ref`
        contains the salt payload only.
  - [ ] Store queries and writes continue to support idempotent allocation replay, issued-allocation
        lookups, and receipt tracking creation without API contract changes.
  - [ ] Existing payment-address store tests are updated to assert the new column names and
        persisted values.
- Notes:
  - The migration may leave non-issued rows with null issuance reference metadata.

## Non-functional requirements

- Maintainability (NFR-001):
  - The final schema and internal type names must be understandable without requiring chain-specific
    tribal knowledge.
- Reliability (NFR-002):
  - Migration execution must preserve existing uniqueness guarantees and address replay behavior.
- Security/Privacy (NFR-003):
  - The refactor must not reintroduce public-index-like Ethereum issuance metadata or weaken the
    current CREATE2 salt privacy model.

## Dependencies and integrations

- External systems:
  - PostgreSQL / Cloudflare Postgres.
- Internal services:
  - Payment address allocation persistence adapters.
  - Bitcoin and Ethereum issued-address derivers.
  - Bootstrap policy wiring for issuance configuration.
