## 1. API And Storage Behavior

- [x] 1.1 Verify `DELETE /api/v1/services/{serviceId}` and `DELETE /api/v1/services/{serviceId}/monitors/{monitorId}` are routed from SST/API Gateway to `services/monitor-api`.
- [x] 1.2 Align service deletion behavior with the spec: allow draft/archived deletion, reject active deletion with conflict, return not found for missing services, and return `204 No Content` on success.
- [x] 1.3 Align monitor deletion behavior with the spec: remove existing nested monitors, reject deleting the last monitor from an active service with conflict, return not found for missing monitors, and return `204 No Content` on success.
- [x] 1.4 Update repository delete logic so successful service and monitor deletion writes deletion audit records that are not removed by the same delete operation.
- [x] 1.5 Ensure service deletion removes active service metadata, tenant references, service status, child monitor metadata, child monitor references, current monitor status, and notification links from active read paths.
- [x] 1.6 Ensure monitor deletion recalculates parent service monitor count, enabled count, lifecycle state, and rollup status from remaining monitors.

## 2. Dashboard API And Actions

- [x] 2.1 Add dashboard API helpers for deleting a service and deleting a nested monitor.
- [x] 2.2 Add server actions for service deletion and monitor deletion with consistent error redirect handling.
- [x] 2.3 Revalidate `/services`, affected service detail, and affected monitor detail paths after successful deletion.
- [x] 2.4 Redirect successful service deletion to `/services` and successful monitor deletion to the parent service detail route.

## 3. Dashboard UI

- [x] 3.1 Add a guarded delete control to service detail for eligible services.
- [x] 3.2 Show clear conflict/error feedback when service deletion is blocked or fails.
- [x] 3.3 Add a guarded delete control to monitor detail for eligible monitors.
- [x] 3.4 Show clear conflict/error feedback when monitor deletion is blocked or fails.
- [x] 3.5 Ensure delete copy distinguishes permanent deletion from archive and disable flows.

## 4. Tests And Verification

- [x] 4.1 Add or update monitor API handler tests for service delete success, missing service, and active-service conflict.
- [x] 4.2 Add or update monitor API handler tests for monitor delete success, missing monitor, and last-active-monitor conflict.
- [x] 4.3 Add or update repository tests covering deleted configuration absence, service status recalculation, notification-link cleanup, and deletion audit preservation.
- [x] 4.4 Add or update dashboard tests or component coverage for delete controls and server-action redirects where the project test setup supports it.
- [x] 4.5 Run `make test-go-all`.
- [x] 4.6 Run `make lint-dashboard` and `make check-dashboard`.
