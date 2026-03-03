---
doc: 00_problem
spec_date: 2026-03-03
slug: cmd-app-compose-prefix
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-03-postgresql18-migration-runner-container
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background: Current project entrypoint folder is `cmd/payrune`, and compose currently lacks explicit project prefix naming.
- Users or stakeholders: Developers running local compose stack and maintaining command layout consistency.
- Why now: Required naming alignment is `cmd/app` and compose resource prefix `payrune`.

## Constraints (optional)

- Technical constraints:
  - Move app command folder from `cmd/payrune` to `cmd/app`.
  - Keep build/runtime behavior equivalent after rename.
  - Add compose project prefix `payrune` without breaking existing services.
- Timeline/cost constraints:
  - Keep changes minimal and non-disruptive.
- Compliance/security constraints:
  - Follow spec-first workflow and pass spec lint.

## Problem statement

- Current pain:
  - Command path naming does not match requested `app` naming.
  - Compose resources (container/network/volume names) use default directory-based prefix.
- Evidence or examples:
  - Existing command file path is `cmd/payrune/main.go`.

## Goals

- G1: Rename app command directory to `cmd/app`.
- G2: Update affected build/deploy references to new command path.
- G3: Configure compose project prefix to `payrune`.

## Non-goals (out of scope)

- NG1: Change application module name or package import root.
- NG2: Change service runtime behavior/endpoints.

## Assumptions

- A1: Binary output file can remain `payrune` while command path changes.
- A2: Compose prefix requirement targets docker compose project/resource names.

## Open questions

- Q1: Should future make targets explicitly pass `-p payrune` as an extra safeguard?
- Q2: Should old default-prefixed compose volumes be cleaned manually when needed?

## Success metrics

- Metric: Command layout.
- Target: main entrypoint is located at `cmd/app/main.go` and builds successfully.
- Metric: Compose naming.
- Target: `docker compose ... config` resolves project name `payrune`.
- Metric: Stack health.
- Target: `make up` still boots services and `make down` still cleans up.
