---
doc: 03_tasks
spec_date: 2026-03-09
slug: bitcoin-compose-defaults
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-09-poller-interval-separation
  - 2026-03-09-sticky-paid-unconfirmed-status
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale:
  - This merged change only updates Compose defaults and consolidates rollout sizing notes into one spec; it does not add schema or runtime flow changes.
- Upstream dependencies (`depends_on`):
  - `2026-03-09-poller-interval-separation`
  - `2026-03-09-sticky-paid-unconfirmed-status`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - No code-path or schema redesign is involved.
  - What would trigger switching to Full mode:
    - Adding runtime jitter, new config fields, or a real confirmation-grace expiry rule.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Validation steps are listed under each task below.

## Milestones

- M1:
  - Consolidate Bitcoin Compose default tuning into one spec with rollout numbers.
- M2:
  - Apply the merged defaults to both Bitcoin Compose files and verify they remain valid.

## Tasks (ordered)

1. T-001 - Merge Compose default specs and sizing data
   - Scope:
     - Replace the split cadence/expiry draft specs with one merged Compose-default spec that includes the operational API/CU numbers and Validation Cloud cost notes.
   - Output:
     - `specs/2026-03-09-bitcoin-compose-defaults/*.md`
   - Linked requirements: FR-004, FR-005, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-09-bitcoin-compose-defaults" bash scripts/spec-lint.sh`
     - [x] Expected result: spec lint passes and one spec now documents cadence, unpaid expiry, sizing results, and Validation Cloud rollout cost notes together.
     - [x] Logs/metrics to check (if applicable): none
2. T-002 - Apply merged Bitcoin Compose defaults
   - Scope:
     - Keep the smoother poller defaults, raise required confirmations to `2`, and keep the shorter unpaid receipt defaults in the two Bitcoin Compose files.
   - Output:
     - `deployments/compose/compose.bitcoin.mainnet.yaml`
     - `deployments/compose/compose.bitcoin.testnet4.yaml`
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, NFR-001, NFR-002, NFR-003, NFR-005, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `ruby --disable-gems -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f) }' deployments/compose/compose.bitcoin.mainnet.yaml deployments/compose/compose.bitcoin.testnet4.yaml` and `SPEC_DIR="specs/2026-03-09-bitcoin-compose-defaults" bash scripts/spec-lint.sh`
     - [x] Expected result: both Compose files parse successfully and expose the merged defaults.
     - [x] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-002
- FR-002 -> T-002
- FR-003 -> T-002
- FR-004 -> T-002
- FR-005 -> T-001
- FR-006 -> T-001
- NFR-001 -> T-002
- NFR-002 -> T-002
- NFR-003 -> T-002
- NFR-005 -> T-002
- NFR-006 -> T-001, T-002

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Restore the previous Compose default env values if the lower-cost profile is too stale or the `24h` unpaid payment window is too aggressive.
