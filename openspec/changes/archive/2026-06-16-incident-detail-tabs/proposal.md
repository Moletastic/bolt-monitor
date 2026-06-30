## Why

The incident detail page at `/incidents/[id]` currently shows only 3 static timestamps (Opened, Acknowledged, Resolved). The underlying data — `IncidentActivity` records, `CheckRun` history, and `AuditEvent` records — already exist in DynamoDB but have no read path to the dashboard. Operators have no visibility into what happened around an incident beyond its lifecycle state changes.

## What Changes

- **Timeline tab** on `/incidents/[id]`: Shows `IncidentActivity` records for the incident — all state transitions (opened, acknowledged, resolved) in chronological order. Currently written but never read back.
- **Alert History tab** on `/incidents/[id]`: Shows `CheckRun` records for the incident's monitor, filtered to the incident's time window. Visually annotates runs that correspond to state transitions.
- **Audit tab** on `/incidents/[id]`: Shows `AuditEvent` records for both the monitor and its parent service. Merges two queries on the frontend (monitor events + service events), sorted by timestamp.
- **New API endpoint**: `GET /api/v1/incidents/{id}/activities` — returns `IncidentActivity` records for an incident.
- **Audit API extension**: Monitor audit events are already exposed at `GET /api/v1/monitors/{id}/audit`. This endpoint is reused; service-level audit events (archive, reactivate) are fetched from a second call.

## Capabilities

### New Capabilities

- `incident-activity-read-api`: System SHALL allow clients to read activity/timeline events for an individual incident through HTTP API. This surfaces the full state-transition history of an incident, not just its current status.

### Modified Capabilities

- `audit-event-read-api`: Extend to support per-incident audit views that include service-level events alongside monitor-level events. The existing scenario (`GET /api/v1/monitors/{id}/audit`) is unchanged; the extension is a new read pattern scoped to the incident's monitor + service.

## Impact

- **New endpoint**: `GET /api/v1/incidents/{id}/activities` in `monitor-api` service
- **New repository method**: `ListIncidentActivities` in `monitor-api` repository
- **Dashboard**: New tab components on `/incidents/[id]` page (Timeline, Alert History, Audit)
- **Existing endpoints reused**: `GET /api/v1/services/{sid}/monitors/{mid}/runs` (Alert History), `GET /api/v1/services/{sid}/monitors/{mid}/audit` (Audit)
- **No schema changes**: All data already persisted; only read paths are added
