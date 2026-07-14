## Requirements

### Requirement: System creates monitors through HTTP API
System SHALL allow clients to create monitor configurations through HTTP API.

#### Scenario: Client creates monitor with server-generated ID
- **WHEN** client submits valid create request to monitor collection under existing service (no `monitorId`)
- **THEN** system generates a unique `monitorId` derived from monitor type and target URL (or name fallback)
- **AND** system persists nested monitor records with generated `monitorId`
- **AND** system returns `201 Created` with `monitorId` in response body

#### Scenario: Client creates monitor
- **WHEN** client submits valid create request to monitor API
- **THEN** system validates monitor payload, persists monitor records, and returns created monitor resource

### Requirement: System lists and reads monitors through HTTP API
System SHALL allow clients to list monitors and fetch a single monitor through HTTP API.

#### Scenario: Client lists monitors
- **WHEN** client requests monitor collection
- **THEN** system returns monitor resources scoped to current tenant/workspace context and may include current status summary

#### Scenario: Client fetches monitor by ID
- **WHEN** client requests existing monitor by ID
- **THEN** system returns persisted monitor resource for that monitor with operational status data available through the read surface

### Requirement: System updates monitor configuration through HTTP API
System SHALL allow clients to update mutable monitor configuration fields through HTTP API.

#### Scenario: Client updates monitor
- **WHEN** client submits valid update request for existing monitor
- **THEN** system validates changed fields and persists updated monitor configuration

### Requirement: System enables and disables monitor lifecycle through HTTP API
System SHALL allow clients to explicitly enable and disable monitors through HTTP API.

#### Scenario: Client disables monitor
- **WHEN** client calls disable operation for existing monitor
- **THEN** system persists disabled lifecycle state for that monitor

#### Scenario: Client enables monitor
- **WHEN** client calls enable operation for existing monitor
- **THEN** system persists enabled lifecycle state for that monitor

### Requirement: System validates monitor CRUD payloads against canonical monitor contracts
System SHALL validate monitor CRUD payloads against the canonical monitor contract.

#### Scenario: Client submits valid monitor payload
- **WHEN** client submits a monitor create or update payload with valid name, type, cadence, enabled state, thresholds, and HTTP configuration
- **THEN** system accepts the payload without requiring probe-location selection

#### Scenario: Client submits obsolete probe-location selection
- **WHEN** client submits monitor payload fields for probe-location or region selection
- **THEN** system does not treat those fields as part of the accepted monitor contract

### Requirement: System writes audit records for monitor mutations
System SHALL write audit records for monitor create, update, enable, and disable operations.

#### Scenario: Monitor configuration changes
- **WHEN** client successfully changes monitor configuration or lifecycle state
- **THEN** system persists corresponding audit event records for that mutation

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

## Removed Requirements

### Requirement: System accepts client-provided monitorId
**Reason**: Monitor ID is now server-generated slug from type+URL or name fallback. Clients no longer provide `monitorId` on create.

**Migration**: Remove `monitorId` field from create request. Use returned `monitorId` from response.

### Requirement: Validation errors carry field paths in reason.details

When validation rejects a request, the response body's `reason.code` SHALL be `VALIDATION_FAILED` and `reason.details` SHALL include a `"field"` key identifying the offending field as a dotted path. A `"reason"` key MAY accompany `"field"` for human context.

#### Scenario: Service name validation surfaces the field
- **WHEN** a client POSTs a service with an empty `name`
- **THEN** the response body's `reason.code` is `VALIDATION_FAILED` and `reason.details.field` is `"name"`

#### Scenario: HTTP target URL validation surfaces the nested path
- **WHEN** a client POSTs a monitor with a malformed `http.target`
- **THEN** the response body's `reason.code` is `VALIDATION_FAILED` and `reason.details.field` is `"http.target"`

### Requirement: Repository sentinels surface typed errors

The monitor API repository SHALL return `*shared/errors.TypedError` for all sentinel conditions (`ServiceAlreadyExists`, `MonitorAlreadyExists`, `CannotDeleteActiveService`, `CannotDeleteLastMonitorFromActiveService`, `IncidentNotActionable`, `MissingTableName`). The handler SHALL NOT use `errors.Is` to switch on these sentinels; errors SHALL route through `shared/errors.Respond`.

#### Scenario: Duplicate service surfaces SERVICE_ALREADY_EXISTS
- **WHEN** a client POSTs a service with an ID that already exists
- **THEN** the response body's `reason.code` is `SERVICE_ALREADY_EXISTS` and the status is 409

#### Scenario: Cannot delete active service surfaces SERVICE_ACTIVE
- **WHEN** a client DELETEs a service whose lifecycleState is "active"
- **THEN** the response body's `reason.code` is `SERVICE_ACTIVE` and the status is 409

### Requirement: Not-found responses carry no details.id

When a CRUD endpoint returns a not-found response, the response body's `reason.details` SHALL NOT include an `id` key. The dashboard has no consumer of that key and the wire stays cleaner without it.

#### Scenario: Unknown service omits details.id
- **WHEN** a client GETs `/api/v1/services/UNKNOWN`
- **THEN** the response body's `reason.code` is `SERVICE_NOT_FOUND` and `reason.details` is empty
