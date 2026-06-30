## Why

Operators can create and manage services and monitors, but there is no permanent removal path for configuration that was created by mistake, is obsolete, or should no longer appear in operational views. Archive and disable preserve history and state; this change adds explicit destructive deletion for cases where the resource should be removed from active management surfaces.

## What Changes

- Add a permanent service deletion API that removes a service and its child monitor configuration from active service and monitor management.
- Add a permanent monitor deletion API for removing a single nested monitor from a service without deleting the service.
- Require deletion endpoints to return not found for missing resources and to be safe to call only for resources within the current tenant context.
- Write audit records for successful service and monitor deletion.
- Update dashboard service and monitor management views with guarded delete actions and post-delete navigation/refresh behavior.
- Clarify storage behavior for deleting primary configuration records while preserving append-only operational history unless explicitly covered by existing retention policies.
- **BREAKING**: Deleted services and monitors no longer appear in normal list/read APIs and direct reads for deleted resources return not found.

## Capabilities

### New Capabilities
- `service-deletion-api`: Permanent service deletion behavior, including cascading child monitor configuration removal and audit recording.

### Modified Capabilities
- `monitor-crud-api`: Add permanent nested monitor deletion behavior to monitor CRUD APIs.
- `dashboard-web-app`: Add guarded service and monitor delete controls and define dashboard state after deletion.
- `dynamodb-single-table-storage`: Clarify configuration deletion and operational-history preservation expectations for deleted services and monitors.

## Impact

- `services/monitor-api`: new delete handlers/routes for services and nested monitors, tenant-scoped validation, storage calls, and audit writes.
- `shared/`: domain/storage contracts for deleting service and monitor configuration records safely.
- `infra/stacks/bootstrap.ts`: API route wiring for service and monitor delete endpoints.
- `apps/dashboard`: delete actions, confirmation UX, refresh/navigation behavior, and error handling for missing/deleted resources.
- DynamoDB single-table records: service and monitor configuration items are removed from active reads; existing run/status/incident/audit history remains governed by current retention/read behavior.
