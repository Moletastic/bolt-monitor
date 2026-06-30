## MODIFIED Requirements

### Requirement: System exposes monitor audit history through HTTP API
System SHALL allow clients to read audit history for an individual nested monitor through HTTP API.

#### Scenario: Operator requests monitor audit history
- **WHEN** operator calls `GET /api/v1/services/{serviceId}/monitors/{monitorId}/audit` for existing nested monitor
- **THEN** system returns audit events associated with that nested monitor

### Requirement: Audit read responses expose stable mutation history fields
System SHALL return audit-event metadata suitable for monitor history views and operator investigation.

#### Scenario: Monitor has recorded mutations
- **WHEN** system returns audit history for nested monitor
- **THEN** each audit event includes stable identity, event type, event timestamp, and actor or origin metadata when available

### Requirement: Audit history is read-only through public API
System SHALL keep audit-event creation under system business process and expose audit history through read routes only.

#### Scenario: Client inspects audit API shape
- **WHEN** client needs mutation history for monitor
- **THEN** system provides read access at `GET /api/v1/services/{serviceId}/monitors/{monitorId}/audit`
- **AND** does not expose generic create, update, or delete audit endpoints

### Requirement: Audit history tolerates empty monitor history
System SHALL allow clients to request audit history even when a monitor has no persisted audit events.

#### Scenario: Monitor has no audit events yet
- **WHEN** operator requests audit history for existing nested monitor with no recorded mutations beyond current implementation scope
- **THEN** system returns successful empty audit collection response
