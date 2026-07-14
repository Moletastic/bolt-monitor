## Requirements

### Requirement: System exposes monitor audit history through HTTP API
System SHALL allow clients to read audit history for an individual monitor through HTTP API.

#### Scenario: Operator requests monitor audit history
- **WHEN** operator calls `GET /api/v1/monitors/{id}/audit` for existing monitor
- **THEN** system returns audit events associated with that monitor

### Requirement: Audit read responses expose stable mutation history fields
System SHALL return audit-event metadata suitable for monitor history views and operator investigation.

#### Scenario: Monitor has recorded mutations
- **WHEN** system returns audit history for monitor
- **THEN** each audit event includes stable identity, event type, event timestamp, and actor or origin metadata when available

### Requirement: Audit history is read-only through public API
System SHALL keep audit-event creation under system business process and expose audit history through read routes only.

#### Scenario: Client inspects audit API shape
- **WHEN** client needs mutation history for monitor
- **THEN** system provides read access at `GET /api/v1/monitors/{id}/audit`
- **AND** does not expose generic create, update, or delete audit endpoints

### Requirement: Audit history tolerates empty monitor history
System SHALL allow clients to request audit history even when a monitor has no persisted audit events.

#### Scenario: Monitor has no audit events yet
- **WHEN** operator requests audit history for existing monitor with no recorded mutations beyond current implementation scope
- **THEN** system returns successful empty audit collection response

### Requirement: Audit history reads are resource-scoped and bounded
System SHALL read monitor and service audit history through a resource-scoped storage access pattern, newest first, limited to 20 matching events per request.

#### Scenario: Operator requests monitor audit history in tenant with unrelated events
- **WHEN** operator requests audit history for one monitor and tenant contains audit events for other services or monitors
- **THEN** system reads only audit events indexed for requested monitor resource
- **AND** system does not retrieve unrelated tenant audit events for application-side filtering

#### Scenario: Operator requests service audit history
- **WHEN** operator requests audit history for one service
- **THEN** system reads only service-level audit events indexed for requested service resource
- **AND** system returns at most 20 events and cursor continuation metadata when more exist

### Requirement: Audit history preserves existing events during index migration
System SHALL preserve read visibility of audit events created before resource-index writes are enabled.

#### Scenario: Historical audit event is backfilled
- **WHEN** an existing audit event lacks resource index keys during migration
- **THEN** migration writes its resource index keys idempotently
- **AND** monitor or service audit readers return that event after index reader cutover
