## ADDED Requirements

### Requirement: System deletes monitors through HTTP API
System SHALL allow clients to permanently delete eligible nested monitor configurations through HTTP API.

#### Scenario: Client deletes monitor
- **WHEN** client sends `DELETE /api/v1/services/{serviceId}/monitors/{monitorId}` for an existing monitor in the current tenant
- **THEN** system removes the monitor from normal monitor list and read APIs
- **AND** system returns `204 No Content`

#### Scenario: Client deletes missing monitor
- **WHEN** client sends `DELETE /api/v1/services/{serviceId}/monitors/{monitorId}` for a monitor that does not exist under that service in the current tenant
- **THEN** system returns not found

#### Scenario: Client deletes last monitor from active service
- **WHEN** client sends `DELETE /api/v1/services/{serviceId}/monitors/{monitorId}` for the only monitor under an active service
- **THEN** system rejects the request with conflict
- **AND** system preserves the monitor configuration

### Requirement: System updates service rollup after monitor deletion
System SHALL update service summary state after a monitor is deleted.

#### Scenario: Monitor deletion changes service coverage
- **WHEN** system successfully deletes a monitor
- **THEN** system recalculates the parent service monitor count, enabled monitor count, lifecycle state, and rollup status from remaining monitors
- **AND** subsequent service reads reflect the recalculated state

### Requirement: System records monitor deletion audit events
System SHALL retain audit evidence for successful monitor deletion.

#### Scenario: Monitor deletion succeeds
- **WHEN** system successfully deletes a monitor
- **THEN** system records an audit event identifying the parent service, deleted monitor, and deletion action
- **AND** the audit event is not removed as part of the same monitor configuration deletion
