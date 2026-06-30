## Context

Repository already separates monitor configuration and probe execution concerns reasonably well. `Monitor`, `ExecutionRequest`, `ExecutionResult`, `CheckRun`, and `MonitorStatus` establish internal domain flow, but operator-visible runtime actions and read models are still implicit or missing from API contracts. This change should add explicit API surfaces only for models users or admins need to interact with directly, while keeping queue, lease, and pipeline artifacts internal.

## Goals / Non-Goals

**Goals:**
- Define one spec for manual run command behavior.
- Define one spec for incident read and operator action behavior.
- Define one spec for admin scheduler configuration behavior.
- Define one spec for audit-event read behavior.
- Keep internal execution pipeline models out of public CRUD/API contracts.

**Non-Goals:**
- No implementation of execution workers, queues, leases, or scheduler internals in this change.
- No generic CRUD for incidents, runs, or audit events.
- No auth or RBAC design beyond path-level distinction between operator and admin surfaces.
- No redesign of existing monitor CRUD or status/run read routes except where new surfaces must reference them.

## Decisions

### Use command endpoints for operator actions, not CRUD for derived runtime objects
- Decision: model manual run, incident acknowledgement, and incident resolution as command endpoints.
- Rationale: these actions trigger business workflows and should not pretend to be direct record creation by clients.

### Keep incidents system-owned
- Decision: expose incidents through read endpoints and operator action endpoints only.
- Rationale: incidents should open and close from business rules derived from check results, not from arbitrary client-side CRUD.

### Keep scheduler controls on an admin-only surface
- Decision: expose recurring execution control through `/api/v1/admin/...` routes.
- Rationale: scheduler behavior is platform control-plane state, not normal monitor configuration.

### Scope audit reads to monitor history first
- Decision: require at least a monitor-scoped audit history route before any broader audit search surface.
- Rationale: monitor detail views already have a natural ownership context, and this keeps the first audit read contract small.

## Risks / Trade-offs

- [Risk] Manual run API can imply synchronous execution expectations. -> Mitigation: specify command acceptance semantics and require a returned run identifier without requiring inline completion.
- [Risk] Incident action routes can overlap with automatic incident closure logic. -> Mitigation: spec operator actions as state transitions layered on top of system-owned incident lifecycle.
- [Risk] Admin scheduler control can become a hidden second source of truth relative to deploy-time config. -> Mitigation: define one read/write API contract for runtime scheduler control state.
- [Risk] Audit APIs can expose noisy internal event detail. -> Mitigation: scope first contract to operator-relevant mutation history with stable event fields.

## Migration Plan

1. Add new API capability specs for manual run, incidents, scheduler admin config, and audit-event reads.
2. Implement route declarations and handler behavior for new surfaces.
3. Update OpenAPI docs to describe the added routes.
4. Wire dashboard/admin clients to new surfaces where needed.

Rollback: remove the new route declarations and leave internal execution models unchanged.

## Open Questions

- Should manual run command create a dedicated top-level run lookup route, or is returned `runId` plus existing monitor-scoped run history enough for v1?
- Should incident resolve remain available to operators after system auto-recovery, or should that action only annotate closed incidents?
- Does scheduler config need tenant scope later, or is one global admin config acceptable for current single-tenant bootstrap?
