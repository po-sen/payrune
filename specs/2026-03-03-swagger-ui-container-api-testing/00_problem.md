---
doc: 00_problem
spec_date: 2026-03-03
slug: swagger-ui-container-api-testing
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-03-deploy-service-compose-dockerfile
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background: Swagger UI is available in compose, but the current testing flow is not aligned with direct calls to API host port 8080.
- Users or stakeholders: Developers and QA testing APIs via Swagger UI.
- Why now: The required workflow is Swagger UI on 8081 directly calling `http://localhost:8080` for API testing.

## Constraints (optional)

- Technical constraints:
  - Keep Swagger UI container exposed on `http://localhost:8081`.
  - Swagger-generated requests must target `http://localhost:8080` directly.
  - API service must return valid CORS headers for browser requests from Swagger origin.
- Timeline/cost constraints:
  - Keep changes minimal and compatible with existing compose + make flow.
- Compliance/security constraints:
  - Follow spec-first process and keep spec-lint clean before implementation.

## Problem statement

- Current pain:
  - Current Swagger request pattern uses proxy-path style and does not satisfy direct-host calling requirement.
  - Browser cross-origin calls from 8081 to 8080 need explicit CORS handling.
- Evidence or examples:
  - Generated request example currently shows `http://localhost:8081/api/health` instead of direct `http://localhost:8080/health`.

## Goals

- G1: Make Swagger call `http://localhost:8080` directly from UI.
- G2: Add CORS support in payrune HTTP path for Swagger origin `http://localhost:8081`.
- G3: Preserve simple local workflow with `make up` and `make down`.

## Non-goals (out of scope)

- NG1: Introduce production API gateway behavior.
- NG2: Implement authentication/authorization flows in Swagger.

## Assumptions

- A1: Swagger UI origin remains `http://localhost:8081`.
- A2: `/health` remains the baseline endpoint for testing.

## Open questions

- Q1: Should CORS allow list be configurable later via environment variables?
- Q2: Should additional local origins (for alternate ports) be allowed in a future change?

## Success metrics

- Metric: Swagger target host.
- Target: OpenAPI `servers` URL is `http://localhost:8080`.
- Metric: CORS compliance.
- Target: `curl -i -H "Origin: http://localhost:8081" http://localhost:8080/health` includes allow-origin header.
- Metric: Local developer flow.
- Target: `make up` starts both services and Swagger can execute `/health` successfully.
