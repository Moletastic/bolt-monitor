## Context

The system already has a canonical probe-location catalog in shared code and monitor CRUD already validates monitor selections against that catalog. However, the catalog is not exposed through HTTP API, so frontend create and edit flows cannot discover valid options at runtime. Dashboard v1 can temporarily work with hardcoded assumptions, but a stable product surface should expose probe locations explicitly.

The current in-process catalog is intentionally small and only includes one enabled location. That is enough to define the API shape now, before the catalog grows or gains environment-specific configuration.

## Goals / Non-Goals

**Goals:**
- Expose selectable probe locations through a read-only HTTP API.
- Return a response shape that is directly usable by dashboard forms.
- Keep the API aligned with the canonical shared `probelocationcatalog` model.
- Avoid leaking unnecessary worker-routing internals into the frontend contract.

**Non-Goals:**
- Managing probe locations through product UI.
- Tenant-specific filtering, entitlement, or RBAC rules.
- Worker deployment topology, routing, or billing semantics.
- Reworking monitor CRUD or scheduler behavior beyond enabling option discovery.

## Decisions

### Expose a read-only collection endpoint
- Decision: add a simple collection read endpoint for probe locations rather than embedding locations into monitor endpoints.
- Rationale: keeps location discovery explicit, cacheable, and reusable across create/edit flows.
- Alternative considered: include probe-location catalog in monitor list or monitor detail responses.
- Why not: duplicates data and couples unrelated reads.

### Return selection-oriented fields first
- Decision: the API should expose `locationId`, `displayName`, and `enabled` as the primary frontend contract.
- Rationale: enough for operator selection and form rendering.
- Alternative considered: expose the full internal model including `executionTarget`.
- Why not: frontend does not need routing details for v1 and exposing them too early hardens internals into API surface.

### Expose selectable locations for the current environment
- Decision: endpoint should return the set of locations valid for selection in the current deployment context.
- Rationale: matches the monitor validation story and leaves room for future environment-specific configuration.
- Alternative considered: return disabled and internal-only locations by default.
- Why not: creates confusion in the UI and weakens the meaning of the picker.

### Return enabled/selectable locations only in v1
- Decision: the public v1 read endpoint should return only enabled probe locations that operators can actually select.
- Rationale: keeps dashboard forms simple and makes the endpoint semantics match user intent exactly.
- Alternative considered: return all catalog entries with an `enabled` flag.
- Why not: useful for future admin surfaces, but unnecessary noise for the first product-facing picker.

### Keep the endpoint independent from monitor CRUD
- Decision: do not expand monitor CRUD payloads to include catalog snapshots.
- Rationale: location discovery and monitor mutation are separate concerns with different caching and evolution patterns.
- Alternative considered: add locations to create/update form boot payloads later.
- Why not: premature composition; a simple read endpoint is clearer.

### Use `/api/v1/probe-locations` as the public route
- Decision: expose the collection at `/api/v1/probe-locations`.
- Rationale: matches existing API versioning and uses direct resource naming.
- Alternative considered: nest the route under monitor-specific or catalog-specific paths.
- Why not: probe locations are a reusable top-level resource, not monitor detail.

### Guarantee API-side display-name sorting
- Decision: collection responses should be sorted by display name before returning to clients.
- Rationale: keeps picker behavior stable across clients and reduces duplicated sorting logic.
- Alternative considered: leave ordering unspecified and let clients sort.
- Why not: creates inconsistent UX and unnecessary frontend work for a canonical catalog.

## Risks / Trade-offs

- [Risk] If frontend later needs more metadata, the response shape may need expansion. -> Mitigation: keep the endpoint additive and minimal initially.
- [Risk] Future tenant-specific rules may complicate what "selectable" means. -> Mitigation: define current behavior around environment-valid selectable locations and extend later with explicit semantics.
- [Risk] Hiding `executionTarget` may frustrate internal debugging. -> Mitigation: keep worker-routing metadata available in backend code and add admin/internal surface later if actually needed.

## Migration Plan

1. Define probe-location read capability and requirements.
2. Add API route and handler that expose the current catalog in selection-friendly shape.
3. Update dashboard implementation plan to consume this endpoint for create/edit monitor flows.

Rollback is low risk because this adds a read-only surface without changing persisted monitor data.

## Open Questions
