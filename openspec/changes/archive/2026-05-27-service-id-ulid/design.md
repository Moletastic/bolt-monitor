## Context

Currently the API accepts `serviceId` and `monitorId` as client-provided slugs in create requests. This requires clients to generate IDs and creates tight coupling. ULIDs (Universally Unique Lexicographically Sortable Identifiers) provide monotonic ordering, collision resistance, and can be generated server-side using the existing ULID infrastructure already present in `ids.go`.

## Goals / Non-Goals

**Goals:**
- Service ID generated server-side as ULID with `SVC_` prefix
- Monitor ID generated server-side as slug derived from type + URL or name
- IDs returned in response body and Location header
- Remove slug validation for IDs

**Non-Goals:**
- ID migration for existing services (existing IDs remain as-is)
- Changing ID formats for other entity types (incident, run, etc.)

## Decisions

### Decision 1: Service ID format

**Option A: Raw ULID with `SVC_` prefix** — `SVC_01ARZ3NDEKTSV4RRFFQ69G5FAV`
**Option B: ULID without prefix** — `01ARZ3NDEKTSV4RRFFQ69G5FAV`
**Option C: Slugified name + ULID suffix** — `my-service-01ARZ3N`

**Chosen: Option A** — Consistent with existing patterns in `ids.go` (e.g., `MON_`, `AUD_`, `RUN_` prefixes). Easy to identify entity type from ID alone.

### Decision 2: Monitor ID generation algorithm

Monitor ID should be deterministic-ish (same inputs → same output) but unique enough to avoid collisions.

**Option A: `type-target-slug` derived** — `https-api-example-com-abc123` where target is URL host/path truncated and hashed
**Option B: `type-name-slug` derived** — `http-health-check-abc123` from name + type
**Option C: ULID-based like services** — `MON_01ARZ3NDEKTSV4RRFFQ69G5FAV`

**Chosen: Option A** — The monitor's identity should derive from what it monitors (the URL), not an arbitrary name. This makes monitors more identifiable by their target. Fallback to name if URL parsing fails.

### Decision 3: Remove serviceId/monitorId from request entirely vs accept-and-ignore

**Chosen: Remove entirely** — Clean API, no confusion. Clients must handle generated IDs.

## Risks / Trade-offs

[Risk] Breaking API change for existing clients
→ **Mitigation**: Version the API or coordinate with dashboard client changes

[Risk] Monitor ID collisions if URL changes
→ **Mitigation**: Include a short hash suffix to reduce collision probability; handle 409 on collision

## Open Questions

1. Should we emit audit events for ID generation changes?
2. Do we need backward compatibility shim for transition period?
