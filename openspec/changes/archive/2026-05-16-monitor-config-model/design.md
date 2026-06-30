## Context

Repository has SST bootstrap and one health endpoint, but no domain model for monitors themselves. Before adding CRUD APIs, DynamoDB records, schedulers, or probers, system needs one agreed monitor configuration contract so every downstream component speaks same shape.

This change is cross-cutting because monitor configuration will be created by API handlers, stored in DynamoDB, consumed by schedulers, executed by probers, surfaced in dashboards, and referenced by incidents and audit records.

## Goals / Non-Goals

**Goals:**
- Define canonical monitor configuration fields for v1.
- Support HTTP monitors first with room for more monitor types later.
- Specify validation and lifecycle state expectations.
- Establish ownership boundary through `tenantId` and stable monitor identity.

**Non-Goals:**
- Implement monitor CRUD endpoints.
- Implement DynamoDB table schema details or GSIs.
- Implement check execution, alerting, or scheduling logic.
- Define non-HTTP monitor-specific fields beyond extensibility needs.

## Decisions

### Define one canonical monitor object early
- Decision: use one monitor configuration contract as source of truth across API, persistence, and execution layers.
- Rationale: avoids schema drift once multiple services start reading and writing monitor data.
- Alternative considered: let each subsystem define its own local shape first.
- Why not: creates migration and translation work later.

### Optimize v1 for HTTP monitors
- Decision: v1 required fields and semantics target HTTP checks first.
- Rationale: current product path and `/api/health` slice already center HTTP behavior; easiest real monitor type to land next.
- Alternative considered: fully generic model for all future monitor types.
- Why not: over-design now; weakens validation and makes first API harder to reason about.

### Include `tenantId` from day 1
- Decision: every monitor belongs to a tenant/workspace boundary even if v1 product runs with one default tenant.
- Rationale: future-safe ownership, audit, and query partitioning at low cost.
- Alternative considered: add tenant later.
- Why not: painful backfill and key redesign later.

### Target DynamoDB single-table persistence
- Decision: future persistence for monitor configuration should assume single-table DynamoDB design, not separate tables per model.
- Rationale: aligns monitor identity, status, runs, incidents, and audit records around shared key design and access patterns.
- Alternative considered: multiple DynamoDB tables split by entity.
- Why not: weakens cross-entity query planning and pushes avoidable migration work into later data-model changes.

### Separate identity, behavior, and lifecycle concerns
- Decision: group fields into stable identity (`monitorId`, `tenantId`), check behavior (`target`, `method`, expectations, timeout), cadence (`intervalSeconds`, `regions`), and lifecycle (`enabled`).
- Rationale: simplifies validation and keeps future API/resource evolution clear.
- Alternative considered: one flat unstructured config blob.
- Why not: harder to validate, query, and audit.

## Risks / Trade-offs

- [Risk] HTTP-first model may need extension for TCP/DNS/SSL later. -> Mitigation: reserve `type` field and keep HTTP-only constraints explicit.
- [Risk] Too many optional fields can weaken validation. -> Mitigation: define minimum required field set and clear per-field semantics.
- [Risk] Missing lifecycle semantics can confuse future scheduler behavior. -> Mitigation: lock basic enabled/disabled contract now.

## Migration Plan

1. Define monitor configuration spec and required fields.
2. Implement shared model/types from spec.
3. Reuse model in future monitor CRUD API, persistence layer, and scheduler/prober work.

Rollback is simple at this stage because this change defines contract only; implementation changes can be dropped before dependent features ship.

## Open Questions

- Should `name` be unique per tenant, or only `monitorId` globally unique?
- Should `regions` be optional in v1 with one default region, or always explicit?
- Should expected response matching support only status code in first implementation, or also body substring from day 1?
