## Context

The incident detail page (`/incidents/[id]`) currently shows three static timestamps — Opened, Acknowledged, Resolved — on a Timeline card. The underlying data (`IncidentActivity`, `CheckRun`, `AuditEvent` records) already exists in DynamoDB but has no read path to the dashboard. The check-runtime writes `IncidentActivity` records on every incident state transition, but monitor-api has no repository method or HTTP endpoint to retrieve them.

## Goals / Non-Goals

**Goals:**
- Add three tabs to the incident detail page: Timeline, Alert History, Audit
- Add `GET /api/v1/incidents/{id}/activities` endpoint to read `IncidentActivity` records
- Reuse existing `GET /api/v1/services/{sid}/monitors/{mid}/runs` for Alert History
- Extend Audit tab to include service-level events via second API call on frontend

**Non-Goals:**
- Split-pane layout (separate pages MVP)
- Actor column or affected users metric
- Global `/audit-trail` page (separate future change)
- Changes to write paths or schema
- Tab for "Monitor Changes" (folds into Audit tab)

## Decisions

### 1. New endpoint `GET /api/v1/incidents/{id}/activities` vs. embedding in existing `GET /incident/{id}`

**Decision:** New dedicated endpoint.

**Rationale:** Embedding activities into the incident detail response would change the existing response shape and require versioned handling. A dedicated endpoint keeps the existing incident detail response unchanged and follows the existing pattern of separate endpoints for separate read models (e.g., `/incidents/{id}/ack`, `/incidents/{id}/resolve`).

### 2. Frontend-merged Audit tab vs. new backend endpoint for combined monitor + service audit

**Decision:** Frontend merges two calls.

**Rationale:** The Audit tab needs both monitor-level events (`GET /api/v1/services/{sid}/monitors/{mid}/audit`) and service-level events (no existing endpoint — would need `GET /api/v1/services/{sid}/audit`). A new combined backend endpoint would be cleaner long-term but adds more implementation surface. For MVP, the dashboard tab makes two calls and merges on the frontend.

### 3. Tab layout vs. extending existing card layout

**Decision:** Add tabs to the incident detail page, replacing the current two-card grid (Details + Timeline) with a tabbed layout.

**Rationale:** The current page has a "Details" card and a "Timeline" card. The Alert History and Audit tabs are distinct data domains that don't belong in a card within the existing layout. Tabs are the natural Next.js/React pattern for this (already used elsewhere in the dashboard).

## Risks / Trade-offs

- **DynamoDB query efficiency**: `IncidentActivity` records use `PK = INCIDENT#<id>`, `SK = ACTIVITY#<ts>#<id>`. A simple query with `PK = incidentID` and `SK begins_with "ACTIVITY#"` retrieves all activities. This is a single-table pattern — efficient with no fan-out.

- **Audit tab performance**: Two sequential API calls (monitor events + service events) will be slower than a single call. Acceptable for MVP; a combined endpoint is a future improvement.

- **Time-window filtering for Alert History**: There is no field linking a `CheckRun` to an incident. The frontend filters runs by comparing timestamps to `openedAt` and `resolvedAt` (for closed incidents) or `openedAt` (for open incidents). This is heuristic, not causal — but sufficient for MVP annotation.

- **Service-level audit events**: No endpoint exists to fetch service-level audit events (archive, reactivate). The Audit tab's second call will be a new endpoint to be added as part of this change (`GET /api/v1/services/{sid}/audit`).
