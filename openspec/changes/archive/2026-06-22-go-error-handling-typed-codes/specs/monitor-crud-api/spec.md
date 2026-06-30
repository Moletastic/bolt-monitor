## MODIFIED Requirements

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
