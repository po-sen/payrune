---
doc: 00_problem
spec_date: 2026-03-03
slug: deploy-service-compose-dockerfile
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

# Problem & Goals

## Context

- Background: The repository already has a runnable Go service, but it has no container packaging or compose deployment manifest.
- Users or stakeholders: Developers who need a fast local deployment workflow.
- Why now: We need one-command startup/shutdown for the service container during development.

## Constraints (optional)

- Technical constraints:
  - Dockerfile must be created at `build/app/Dockerfile`.
  - Compose manifest must be created at `deployments/compose/compose.yaml`.
  - `Makefile` should stay minimal and expose `make up` and `make down`.
- Timeline/cost constraints:
  - Keep implementation small and immediately usable.
- Compliance/security constraints:
  - Follow repository spec-first workflow before code/config edits.

## Problem statement

- Current pain:
  - No standard Docker image build path exists for the service.
  - No compose file exists to run the service container.
  - No concise top-level command exists for container lifecycle.
- Evidence or examples:
  - `build/app/` and `deployments/compose/` do not contain required deployment files.
  - No root `Makefile` exists.

## Goals

- G1: Add `build/app/Dockerfile` to package the `payrune` service.
- G2: Add `deployments/compose/compose.yaml` to deploy the service container.
- G3: Add a minimal root `Makefile` with `up` and `down` targets.

## Non-goals (out of scope)

- NG1: Production orchestration (Kubernetes, Helm, cloud IaC).
- NG2: Multi-service topology, databases, or message brokers.

## Assumptions

- A1: Service runtime port remains `8080`.
- A2: Local environment has Docker Engine and Docker Compose v2 available.

## Open questions

- Q1: Should image naming/tagging strategy be parameterized later for CI publishing?
- Q2: Should environment-variable based configuration be added in a follow-up spec?

## Success metrics

- Metric: Container packaging availability.
- Target: `docker build -f build/app/Dockerfile .` succeeds.
- Metric: Local deployment availability.
- Target: `docker compose -f deployments/compose/compose.yaml up -d --build` starts one running service container.
- Metric: Developer ergonomics.
- Target: `make up` and `make down` complete the same lifecycle without additional arguments.
