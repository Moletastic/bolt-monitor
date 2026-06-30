## Context

Project already has shared monitor configuration contract, probe-location catalog, and DynamoDB single-table storage contract. Missing piece is first real product API that lets users manage monitors. Basic CRUD API is right next step because it validates whether the shared contracts actually fit real request/response and persistence flows.

This change is cross-cutting across API routing, request validation, DynamoDB repository operations, and audit writes. It should remain intentionally narrow: monitor configuration management only, not execution or alerting.

## Goals / Non-Goals

**Goals:**
- Expose monitor CRUD endpoints under `/api/v1/monitors`.
- Validate create and update payloads using shared monitor and probe-location contracts.
- Persist monitor config and monitor listing items in single-table DynamoDB.
- Write audit events for configuration mutations.
- Support enable/disable lifecycle endpoints without introducing full soft-delete complexity.

**Non-Goals:**
- Run checks or write `CheckRun` items.
- Compute real status from execution results.
- Add authentication, RBAC, or API-key enforcement.
- Add incidents, alert policies, or notification delivery.

## Decisions

### Start with six endpoints
- Decision: implement create, list, get, patch, enable, and disable endpoints.
- Rationale: enough surface to manage monitor lifecycle without overloading first API slice.
- Alternative considered: only create and get.
- Why not: too incomplete for meaningful product usage or dashboard bootstrap.

### Reuse shared contracts as source of truth
- Decision: route handlers should build on `shared/monitorconfig`, `shared/probelocationcatalog`, and `shared/dynamodbschema` rather than local request-specific models.
- Rationale: proves shared model work and reduces divergence.
- Alternative considered: separate DTOs and storage models inside API service.
- Why not: duplicates early contracts and invites drift.

### Use single-table write set for monitor creation
- Decision: create/update flows should persist at least canonical monitor item, tenant listing ref, and initial status/audit items as appropriate.
- Rationale: list and detail endpoints need both direct and tenant-scoped read paths.
- Alternative considered: store only monitor item first.
- Why not: list endpoint becomes awkward and inconsistent with table design.

### Treat enable/disable as explicit lifecycle operations
- Decision: lifecycle toggles should be separate endpoint operations from generic patching.
- Rationale: keeps state changes auditable and easy to reason about.
- Alternative considered: only patch `enabled` field.
- Why not: weaker intent and harder to model as clear actions later.

## Risks / Trade-offs

- [Risk] CRUD handlers may force small shared-model refinements. -> Mitigation: keep handlers thin and reuse shared contracts directly.
- [Risk] Lack of auth means endpoints are not production-safe yet. -> Mitigation: keep scope personal/internal until auth change lands.
- [Risk] Initial status representation may be speculative before execution pipeline exists. -> Mitigation: keep initial status minimal and lifecycle-oriented.

## Migration Plan

1. Add monitor CRUD routes and handler package.
2. Wire request validation to shared monitor and probe-location contracts.
3. Add DynamoDB repository operations using single-table schema helpers.
4. Verify create/list/get/update/enable/disable flows and audit writes.

Rollback is straightforward because this introduces first monitor API slice only; routes and repository code can be removed without affecting health endpoint bootstrap.

## Open Questions

- Should `PATCH /api/v1/monitors/:id` permit all mutable fields at once, or a narrower subset first?
- Should disabled monitors remain in list responses by default, or require filter semantics later?
- Should create response include initial status snapshot even before execution pipeline exists?
