## 1. monitor-api: Incident Activities Endpoint

- [x] 1.1 Add `incidentActivityRecord` struct to repository.go (PK, SK, EntityType, TenantID, IncidentID, ActivityID, Action, Timestamp)
- [x] 1.2 Add `incidentActivityResponse` type to types.go with fields: activityId, incidentId, action, timestamp
- [x] 1.3 Add `ListIncidentActivities(ctx, tenantID, incidentID) ([]incidentActivityRecord, error)` to repository interface and implementation
- [x] 1.4 Add `toIncidentActivityResponse` converter function to types.go
- [x] 1.5 Add `getIncidentActivities` handler method to handler.go
- [x] 1.6 Add route: `case method == http.MethodGet && strings.HasSuffix(path, "/activities") && incidentID != "":` in handler.go route switch

## 2. monitor-api: Service Audit Endpoint

- [x] 2.1 Add `ListServiceAuditEvents(ctx, tenantID, serviceID) ([]auditEventView, error)` to repository interface and implementation
- [x] 2.2 Add `getServiceAudit` handler method to handler.go
- [x] 2.3 Add route: `case method == http.MethodGet && serviceID != "" && strings.HasSuffix(path, "/audit") && !isMonitorAuditPath(path):` in handler.go route switch (ensure monitor audit takes precedence)
- [x] 2.4 Add `isMonitorAuditPath` helper or reorder route matching to distinguish service vs monitor audit paths

## 3. Dashboard: API Client

- [x] 3.1 Add `getIncidentActivities(incidentId)` to lib/api.ts returning `IncidentActivity[]`
- [x] 3.2 Add `listServiceAuditEvents(serviceId)` to lib/api.ts returning `AuditEvent[]`
- [x] 3.3 Add `IncidentActivity` and `ServiceAuditEvents` response types to lib/api.ts

## 4. Dashboard: Incident Detail Page — Tab Layout

- [x] 4.1 Create reusable `Tabs` component (if not already present in component library) or use existing shadcn/ui tabs pattern
- [x] 4.2 Refactor `/incidents/[id]/page.tsx` to use three-tab layout replacing current two-card grid (Timeline, Alert History, Audit)

## 5. Dashboard: Timeline Tab

- [x] 5.1 Create `TimelineTab` component
- [x] 5.2 Call `getIncidentActivities(incidentId)` on mount
- [x] 5.3 Render activity records sorted by timestamp ascending
- [x] 5.4 Display action/event type and timestamp for each activity
- [x] 5.5 Handle empty state (no activities yet)
- [x] 5.6 Handle loading and error states

## 6. Dashboard: Alert History Tab

- [x] 6.1 Create `AlertHistoryTab` component
- [x] 6.2 Call existing `listMonitorRuns(serviceId, monitorId)` (already used elsewhere in dashboard)
- [x] 6.3 Filter runs to incident time window on frontend: runs where `finishedAt >= openedAt` and (for closed incidents) `finishedAt <= resolvedAt`
- [x] 6.4 Visually annotate runs that correspond to incident state transitions (opened, acknowledged, resolved) using incident activity timestamps
- [x] 6.5 Handle empty, loading, and error states

## 7. Dashboard: Audit Tab

- [x] 7.1 Create `AuditTab` component
- [x] 7.2 Call `listMonitorAuditEvents(serviceId, monitorId)` and `listServiceAuditEvents(serviceId)` in parallel
- [x] 7.3 Merge both result sets on frontend, sort by `occurredAt` timestamp ascending
- [x] 7.4 Render merged audit events in a unified timeline
- [x] 7.5 Handle empty, loading, and error states (partial failure: show one tab's error while displaying other's data gracefully)
